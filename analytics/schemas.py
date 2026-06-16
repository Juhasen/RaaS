from pydantic import BaseModel, Field
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
    user_id: Optional[str] = None
    metadata: dict = Field(..., validation_alias="event_metadata", serialization_alias="metadata")
    created_at: datetime
    
    class Config:
        from_attributes = True

class EventListResponse(BaseModel):
    """Schema for paginated event list"""
    total: int
    count: int
    items: list[EventResponse]

class HostListingStats(BaseModel):
    listing_id: str
    title: str
    price_per_day: float
    bookings_count: int
    pending_bookings: int
    confirmed_bookings: int
    rejected_bookings: int

class HostAnalyticsResponse(BaseModel):
    host_id: str
    total_listings: int
    total_bookings: int
    total_pending: int
    total_confirmed: int
    total_rejected: int
    listings: list[HostListingStats]

