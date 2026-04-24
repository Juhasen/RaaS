from sqlalchemy import Column, String, DateTime, Enum as SQLEnum
from sqlalchemy.ext.declarative import declarative_base
from datetime import datetime
import uuid

Base = declarative_base()

class User(Base):
    """User model for PostgreSQL"""
    __tablename__ = "users"
    
    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    email = Column(String(255), unique=True, nullable=False, index=True)
    password_hash = Column(String(255), nullable=False)
    role = Column(String(50), nullable=False, default="guest")
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)
    
    def __repr__(self):
        return f"<User(id={self.id}, email={self.email}, role={self.role})>"
