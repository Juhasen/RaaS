import os
import logging
from fastapi import FastAPI, HTTPException, Depends
from fastapi.responses import JSONResponse
from contextlib import asynccontextmanager
from sqlalchemy.orm import Session
from sqlalchemy.exc import IntegrityError
import bcrypt
import uuid

from models import User, Base
from schemas import UserCreateRequest, UserResponse, UserListResponse
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
    postgres.engine.create_all(Base.metadata)
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
