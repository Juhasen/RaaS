from sqlalchemy import Column, String, DateTime, JSON
from sqlalchemy.ext.declarative import declarative_base
from datetime import datetime
import uuid

Base = declarative_base()

class Event(Base):
    """Event model for MongoDB/PostgreSQL"""
    __tablename__ = "events"
    
    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    event_type = Column(String(255), nullable=False, index=True)
    user_id = Column(String(36), nullable=False, index=True)
    metadata = Column(JSON, nullable=True)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow, index=True)
    
    def __repr__(self):
        return f"<Event(id={self.id}, event_type={self.event_type}, user_id={self.user_id})>"
