from pydantic import BaseModel
from datetime import datetime

class NotificationLogResponse(BaseModel):
    """Schema for notification log response"""
    id: str
    user_id: str
    message: str
    notification_type: str
    status: str
    created_at: datetime
    
    class Config:
        from_attributes = True
