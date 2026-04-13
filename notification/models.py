from sqlalchemy import Column, String, DateTime, Enum as SQLEnum
from sqlalchemy.ext.declarative import declarative_base
from datetime import datetime
import uuid

Base = declarative_base()

class NotificationLog(Base):
    """Notification log model for PostgreSQL"""
    __tablename__ = "notification_logs"
    
    id = Column(String(36), primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String(36), nullable=False, index=True)
    message = Column(String(1024), nullable=False)
    notification_type = Column(String(50), nullable=False, default="EMAIL")
    status = Column(String(50), nullable=False, default="PENDING")
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)
    
    def __repr__(self):
        return f"<NotificationLog(id={self.id}, user_id={self.user_id}, status={self.status})>"
