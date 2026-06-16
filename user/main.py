import os
import logging
from fastapi import FastAPI, HTTPException, Depends
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

app = FastAPI(title="User Service", version="1.0.0")

# Dependency to get database session
def get_db():
    session = postgres.SessionLocal()
    try:
        yield session
    finally:
        session.close()

JWT_SECRET = os.getenv("JWT_SECRET", "raas-secret-key-12345")
JWT_ALGORITHM = os.getenv("JWT_ALGORITHM", "HS256")
ACCESS_TOKEN_EXPIRE_MINUTES = 1440

security = HTTPBearer()

def create_access_token(data: dict, expires_delta: timedelta = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, JWT_SECRET, algorithm=JWT_ALGORITHM)
    return encoded_jwt

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security),
    db: Session = Depends(get_db)
) -> User:
    token = credentials.credentials
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
        return TokenResponse(access_token=access_token, user=user)
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error during login: {e}")
        raise HTTPException(status_code=500, detail="Internal server error during login")

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
