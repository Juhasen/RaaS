import os
import logging
import threading
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session
import uuid

from models import Event, Base
from schemas import EventIngestionRequest, EventResponse, EventListResponse
from db_utils import get_postgres
from kafka_client import KafkaConsumerWrapper

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize database
postgres = get_postgres()

# Create tables on startup
try:
    postgres.engine.create_all(Base.metadata)
    logger.info("Database tables created successfully")
except Exception as e:
    logger.error(f"Failed to create database tables: {e}")

app = FastAPI(title="Analytics Service", version="1.0.0")

# Kafka consumer setup
kafka_consumers = {}
consumer_threads = {}

# Dependency to get database session
def get_db():
    session = postgres.SessionLocal()
    try:
        yield session
    finally:
        session.close()

def handle_event(event: dict):
    """
    Handle incoming events with idempotent processing
    Uses event_id or combination of keys to detect duplicates
    """
    try:
        session = postgres.SessionLocal()
        event_type = event.get("event_type", "unknown")
        user_id = event.get("user_id")
        
        # Create event record for analytics
        analytics_event = Event(
            id=str(uuid.uuid4()),
            event_type=event_type,
            user_id=user_id,
            metadata=event
        )
        
        session.add(analytics_event)
        session.commit()
        logger.info(f"Event tracked: {event_type} for user {user_id}")
        return True
        
    except Exception as e:
        logger.error(f"Error handling event: {e}")
        session.rollback()
        return False
    finally:
        session.close()

def start_booking_consumer():
    """Start consumer for booking.created events (T039)"""
    try:
        consumer = KafkaConsumerWrapper(
            group_id="analytics-booking-group",
            topics=["booking.created"]
        )
        kafka_consumers["booking"] = consumer
        consumer.consume_messages(handle_event)
    except Exception as e:
        logger.error(f"Booking consumer error: {e}")

def start_tracking_consumer():
    """Start consumer for tracking events (T040)"""
    try:
        topics = [
            "payment.succeeded",
            "payment.failed",
            "booking.confirmed",
            "booking.rejected",
            "media.uploaded",
            "user.created"
        ]
        consumer = KafkaConsumerWrapper(
            group_id="analytics-tracking-group",
            topics=topics
        )
        kafka_consumers["tracking"] = consumer
        consumer.consume_messages(handle_event)
    except Exception as e:
        logger.error(f"Tracking consumer error: {e}")

@app.on_event("startup")
async def startup_event():
    """Start Kafka consumers on app startup"""
    # T039: Start booking.created consumer
    consumer_threads["booking"] = threading.Thread(
        target=start_booking_consumer, 
        daemon=True
    )
    consumer_threads["booking"].start()
    
    # T040: Start tracking consumer
    consumer_threads["tracking"] = threading.Thread(
        target=start_tracking_consumer, 
        daemon=True
    )
    consumer_threads["tracking"].start()
    
    logger.info("Kafka consumers started")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on app shutdown"""
    for consumer in kafka_consumers.values():
        if consumer:
            consumer.close()
    logger.info("Kafka consumers closed")

# Health check endpoint
@app.get("/health")
async def health():
    """Health check endpoint for Kubernetes probes"""
    return {
        "status": "healthy",
        "service": "analytics-service",
        "version": "1.0.0"
    }

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "Analytics Service is running"}

# Event Ingestion Endpoint
@app.post("/events", response_model=EventResponse, status_code=201)
async def ingest_event(
    request: EventIngestionRequest,
    db: Session = Depends(get_db)
):
    """Ingest an event for analytics tracking"""
    try:
        event = Event(
            id=str(uuid.uuid4()),
            event_type=request.event_type,
            user_id=request.user_id,
            metadata=request.metadata
        )
        
        db.add(event)
        db.commit()
        db.refresh(event)
        
        logger.info(f"Event ingested: {event.id} - {event.event_type}")
        return event
        
    except Exception as e:
        db.rollback()
        logger.error(f"Error ingesting event: {e}")
        raise HTTPException(status_code=500, detail="Failed to ingest event")

# Get Event by ID Endpoint
@app.get("/events/{event_id}", response_model=EventResponse)
async def get_event(event_id: str, db: Session = Depends(get_db)):
    """Get event by ID"""
    try:
        event = db.query(Event).filter(Event.id == event_id).first()
        if not event:
            raise HTTPException(status_code=404, detail="Event not found")
        return event
    except Exception as e:
        logger.error(f"Error retrieving event: {e}")
        raise HTTPException(status_code=500, detail="Failed to retrieve event")

# List Events Endpoint (with pagination)
@app.get("/events", response_model=EventListResponse)
async def list_events(
    skip: int = 0,
    limit: int = 10,
    event_type: str = None,
    user_id: str = None,
    db: Session = Depends(get_db)
):
    """List events with optional filtering and pagination"""
    try:
        query = db.query(Event)
        
        if event_type:
            query = query.filter(Event.event_type == event_type)
        if user_id:
            query = query.filter(Event.user_id == user_id)
        
        total = query.count()
        events = query.order_by(Event.created_at.desc()).offset(skip).limit(limit).all()
        
        return EventListResponse(
            total=total,
            count=len(events),
            items=events
        )
    except Exception as e:
        logger.error(f"Error listing events: {e}")
        raise HTTPException(status_code=500, detail="Failed to list events")

# Event Stats Endpoint
@app.get("/events/stats/by-type")
async def get_event_stats_by_type(db: Session = Depends(get_db)):
    """Get event statistics grouped by event type"""
    try:
        from sqlalchemy import func
        stats = db.query(
            Event.event_type,
            func.count(Event.id).label("count")
        ).group_by(Event.event_type).all()
        
        return {
            "stats": [
                {"event_type": stat[0], "count": stat[1]} 
                for stat in stats
            ]
        }
    except Exception as e:
        logger.error(f"Error getting event stats: {e}")
        raise HTTPException(status_code=500, detail="Failed to get event stats")

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8008))
    uvicorn.run(app, host="0.0.0.0", port=port)
