import os
import logging
from prometheus_fastapi_instrumentator import Instrumentator
from fastapi import FastAPI, HTTPException, Depends, Request, Response
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from contextlib import asynccontextmanager
from sqlalchemy.orm import Session
from sqlalchemy.exc import IntegrityError
import bcrypt
import uuid
import jwt
from datetime import datetime, timedelta

from models import User, Base
from schemas import UserCreateRequest, UserLoginRequest, UserResponse, UserListResponse, TokenResponse
from db_utils import get_postgres
from kafka_client import KafkaProducerWrapper

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize database
postgres = get_postgres()
kafka_producer = KafkaProducerWrapper()

# Create tables on startup
try:
    Base.metadata.create_all(postgres.engine)
    logger.info("Database tables created successfully")
except Exception as e:
    logger.error(f"Failed to create database tables: {e}")

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Fail startup instead of running with no admin or with leftover self-registered admins.
    email, password = os.getenv("ADMIN_EMAIL"), os.getenv("ADMIN_PASSWORD")
    if not email or not password:
        raise RuntimeError("ADMIN_EMAIL and ADMIN_PASSWORD must be set")
    session = postgres.SessionLocal()
    try:
        seed_admin(session, email, password)
    except IntegrityError:
        session.rollback()  # another replica seeded at the same time
    finally:
        session.close()
    yield

app = FastAPI(title="User Service", version="1.0.0", lifespan=lifespan)

Instrumentator().instrument(app).expose(app)

# Dependency to get database session
def get_db():
    session = postgres.SessionLocal()
    try:
        yield session
    finally:
        session.close()

JWT_SECRET = os.environ["JWT_SECRET"]  # no fallback: a known default lets anyone forge admin tokens
JWT_ALGORITHM = os.getenv("JWT_ALGORITHM", "HS256")
ACCESS_TOKEN_EXPIRE_MINUTES = 1440
COOKIE_NAME = "raas_token"
COOKIE_SECURE = os.getenv("COOKIE_SECURE", "false").lower() == "true"

security = HTTPBearer(auto_error=False)

def create_access_token(data: dict, expires_delta: timedelta = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, JWT_SECRET, algorithm=JWT_ALGORITHM)
    return encoded_jwt

def user_from_token(token: str | None, db: Session) -> User:
    if not token:
        raise HTTPException(status_code=401, detail="Not authenticated")
    try:
        payload = jwt.decode(token, JWT_SECRET, algorithms=[JWT_ALGORITHM])
        user_id: str = payload.get("sub")
        if user_id is None:
            raise HTTPException(status_code=401, detail="Invalid token claims")
    except jwt.PyJWTError:
        raise HTTPException(status_code=401, detail="Could not validate credentials")
    
    user = db.query(User).filter(User.id == user_id).first()
    if user is None:
        raise HTTPException(status_code=401, detail="User not found")
    return user

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
    db: Session = Depends(get_db)
) -> User:
    return user_from_token(credentials.credentials if credentials else None, db)

async def get_admin_check_user(
    request: Request,
    credentials: HTTPAuthorizationCredentials = Depends(security),
    db: Session = Depends(get_db)
) -> User:
    # Only the gateway auth_request uses the cookie (plain link clicks cannot send a header).
    # Header wins. Every other endpoint stays header-only, so it is not exposed to CSRF.
    token = credentials.credentials if credentials else request.cookies.get(COOKIE_NAME)
    return user_from_token(token, db)

# Health check endpoint
@app.get("/health")
async def health():
    """Health check endpoint for Kubernetes probes"""
    return {
        "status": "healthy",
        "service": "user-service",
        "version": "1.0.0"
    }

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "User Service is running"}

def seed_admin(db: Session, email: str, password: str):
    """Make `email` the only admin. Idempotent; see .changes/infrastructure/admin-tools-sso/spec.md."""
    if len(password) < 8:
        raise ValueError("ADMIN_PASSWORD must be at least 8 characters")
    for u in db.query(User).filter(User.role == "admin", User.email != email).all():
        logger.warning(f"Demoting admin {u.email} to guest")
        u.role = "guest"
    user = db.query(User).filter(User.email == email).first()
    if user is None:
        db.add(User(id=str(uuid.uuid4()), email=email, password_hash=_hash(password), role="admin"))
    elif user.role != "admin":
        logger.error(f"{email} exists with role {user.role}; not promoting it to admin")
    elif not bcrypt.checkpw(password.encode(), user.password_hash.encode()):
        user.password_hash = _hash(password)
    db.commit()

def _hash(password: str) -> str:
    return bcrypt.hashpw(password.encode(), bcrypt.gensalt()).decode()

# User Creation Endpoint
@app.post("/users", response_model=UserResponse, status_code=201)
async def create_user(
    request: UserCreateRequest,
    db: Session = Depends(get_db)
):
    """Create a new user"""
    try:
        # Hash password
        salt = bcrypt.gensalt()
        password_hash = bcrypt.hashpw(request.password.encode('utf-8'), salt)
        
        # Create user
        user = User(
            id=str(uuid.uuid4()),
            email=request.email,
            password_hash=password_hash.decode('utf-8'),
            role=request.role
        )
        
        db.add(user)
        db.commit()
        db.refresh(user)
        
        # Emit user.created event
        event_data = {
            "event_type": "user.created",
            "user_id": user.id,
            "email": user.email,
            "role": user.role,
            "created_at": user.created_at.isoformat()
        }
        kafka_success = kafka_producer.send_event("user.created", event_data, key=user.id)
        if not kafka_success:
            logger.warning(f"Failed to emit user.created event for user {user.id}")
        
        logger.info(f"User created: {user.id}")
        return user
        
    except IntegrityError:
        db.rollback()
        logger.error(f"User with email {request.email} already exists")
        raise HTTPException(status_code=409, detail="Email already registered")
    except Exception as e:
        db.rollback()
        logger.error(f"Error creating user: {e}")
        raise HTTPException(status_code=500, detail="Failed to create user")

# User Registration Endpoint (Alias)
@app.post("/users/register", response_model=UserResponse, status_code=201)
async def register_user(
    request: UserCreateRequest,
    db: Session = Depends(get_db)
):
    """Register a new user"""
    return await create_user(request, db)

# User Login Endpoint
@app.post("/users/login", response_model=TokenResponse)
async def login_user(
    request: UserLoginRequest,
    response: Response,
    db: Session = Depends(get_db)
):
    """Authenticate user and return access token"""
    try:
        user = db.query(User).filter(User.email == request.email).first()
        if not user:
            raise HTTPException(status_code=401, detail="Invalid email or password")
        
        # Verify password
        is_password_correct = bcrypt.checkpw(
            request.password.encode('utf-8'),
            user.password_hash.encode('utf-8')
        )
        if not is_password_correct:
            raise HTTPException(status_code=401, detail="Invalid email or password")
        
        # Create token
        access_token = create_access_token(
            data={"sub": user.id, "email": user.email, "role": user.role}
        )
        
        logger.info(f"User logged in successfully: {user.id}")
        response.set_cookie(
            COOKIE_NAME, access_token, max_age=ACCESS_TOKEN_EXPIRE_MINUTES * 60,
            httponly=True, samesite="lax", secure=COOKIE_SECURE, path="/",
        )
        return TokenResponse(access_token=access_token, user=user)
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error during login: {e}")
        raise HTTPException(status_code=500, detail="Internal server error during login")

@app.post("/users/logout", status_code=204)
async def logout_user(response: Response):
    """Clear the auth cookie"""
    response.delete_cookie(COOKIE_NAME, path="/")


# Gateway auth_request target for the admin tools: 200 only for admins (role read from the DB row).
# Declared before /users/{user_id}.
@app.get("/users/admin-check")
async def admin_check(response: Response, current_user: User = Depends(get_admin_check_user)):
    if current_user.role != "admin":
        raise HTTPException(status_code=403, detail="Admin only")
    response.headers["X-Auth-User"] = current_user.email
    return {"status": "ok"}


# Get Current User Info
@app.get("/users/me", response_model=UserResponse)
async def get_me(current_user: User = Depends(get_current_user)):
    """Get info of the currently logged-in user"""
    return current_user

# Get User by ID Endpoint
@app.get("/users/{user_id}", response_model=UserResponse)
async def get_user(user_id: str, db: Session = Depends(get_db)):
    """Get user by ID"""
    try:
        user = db.query(User).filter(User.id == user_id).first()
        if not user:
            raise HTTPException(status_code=404, detail="User not found")
        return user
    except Exception as e:
        logger.error(f"Error retrieving user: {e}")
        raise HTTPException(status_code=500, detail="Failed to retrieve user")

# List Users Endpoint (with pagination)
@app.get("/users", response_model=UserListResponse)
async def list_users(
    skip: int = 0,
    limit: int = 10,
    db: Session = Depends(get_db)
):
    """List all users with pagination"""
    try:
        total = db.query(User).count()
        users = db.query(User).offset(skip).limit(limit).all()
        return UserListResponse(
            total=total,
            count=len(users),
            items=users
        )
    except Exception as e:
        logger.error(f"Error listing users: {e}")
        raise HTTPException(status_code=500, detail="Failed to list users")

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8002))
    uvicorn.run(app, host="0.0.0.0", port=port)
