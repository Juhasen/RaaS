from pydantic import BaseModel
from datetime import datetime
from typing import Optional, Any

class EventIngestionRequest(BaseModel):
    """Schema for ingesting events"""
    event_type: str
    user_id: str
    metadata: Optional[dict[str, Any]] = {}

class EventResponse(BaseModel):
    """Schema for event response"""
    id: str
    event_type: str
    user_id: str
    metadata: dict
    created_at: datetime
    
    class Config:
        from_attributes = True

class EventListResponse(BaseModel):
    """Schema for paginated event list"""
    total: int
    count: int
    items: list[EventResponse]
