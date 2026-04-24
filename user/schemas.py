from pydantic import BaseModel, EmailStr, Field
from datetime import datetime
from typing import Optional

class UserCreateRequest(BaseModel):
    """Schema for creating a new user"""
    email: EmailStr
    password: str = Field(..., min_length=8)
    role: str = Field(default="guest", pattern="^(guest|host|admin)$")

class UserResponse(BaseModel):
    """Schema for user response"""
    id: str
    email: str
    role: str
    created_at: datetime
    
    class Config:
        from_attributes = True

class UserListResponse(BaseModel):
    """Schema for paginated user list"""
    total: int
    count: int
    items: list[UserResponse]
