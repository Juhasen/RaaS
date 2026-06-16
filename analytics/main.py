import os
import logging
import threading
import urllib.request
import json
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session
import uuid

from models import Event, Base
from schemas import EventIngestionRequest, EventResponse, EventListResponse, HostListingStats, HostAnalyticsResponse
from db_utils import get_postgres
from kafka_client import KafkaConsumerWrapper

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize database
postgres = get_postgres()

# Create tables on startup
try:
    Base.metadata.create_all(bind=postgres.engine)
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

def handle_event(event: dict, topic: str = None):
    """
    Handle incoming events with idempotent processing
    Uses event_id or combination of keys to detect duplicates
    """
    try:
        session = postgres.SessionLocal()
        event_type = event.get("event_type") or event.get("event") or topic or "unknown"
        user_id = event.get("user_id") or event.get("guest_id") or event.get("host_id") or "system"
        
        # Create event record for analytics
        analytics_event = Event(
            id=str(uuid.uuid4()),
            event_type=event_type,
            user_id=user_id,
            event_metadata=event
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
            event_metadata=request.metadata
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

LISTING_SERVICE_URL = os.getenv("LISTING_SERVICE_URL", "http://listing-service.raas.svc.cluster.local:80")
BOOKING_SERVICE_URL = os.getenv("BOOKING_SERVICE_URL", "http://booking-service.raas.svc.cluster.local:80")

@app.get("/analytics/host/{host_id}", response_model=HostAnalyticsResponse)
async def get_host_analytics(host_id: str, db: Session = Depends(get_db)):
    """Fetch analytics for a given host account"""
    try:
        # 1. Fetch listings for this host from the listing microservice
        url = f"{LISTING_SERVICE_URL}/listings?host_id={host_id}&limit=100"
        listings = []
        try:
            req = urllib.request.Request(url, headers={"Accept": "application/json"})
            with urllib.request.urlopen(req, timeout=5) as response:
                if response.status == 200:
                    data = json.loads(response.read().decode('utf-8'))
                    listings = data.get("data", []) or []
        except Exception as err:
            logger.error(f"Failed to fetch listings for host {host_id} from {url}: {err}")
            pass

        if not listings:
            return HostAnalyticsResponse(
                host_id=host_id,
                total_listings=0,
                total_bookings=0,
                total_pending=0,
                total_confirmed=0,
                total_rejected=0,
                listings=[]
            )

        listing_ids = [l.get("id") for l in listings if l.get("id")]
        
        # 2. Fetch all bookings from the booking microservice
        bookings_url = f"{BOOKING_SERVICE_URL}/bookings"
        bookings = []
        try:
            req = urllib.request.Request(bookings_url, headers={"Accept": "application/json"})
            with urllib.request.urlopen(req, timeout=5) as response:
                if response.status == 200:
                    bookings = json.loads(response.read().decode('utf-8'))
        except Exception as err:
            logger.error(f"Failed to fetch bookings from {bookings_url}: {err}")

        # 3. Process bookings to determine the status of each booking for the host's listings
        listing_bookings = {lid: {} for lid in listing_ids}
        for booking in bookings:
            lid = booking.get("listing_id")
            bid = booking.get("id")
            status = booking.get("status", "PENDING")
            
            if lid in listing_bookings and bid:
                # Map status: Go service status is PENDING, CONFIRMED, REJECTED
                listing_bookings[lid][bid] = status

        # 4. Aggregate metrics per listing and in total
        listings_stats = []
        total_listings = len(listings)
        total_bookings = 0
        total_pending = 0
        total_confirmed = 0
        total_rejected = 0

        for listing in listings:
            lid = listing.get("id")
            title = listing.get("title", "Unknown Title")
            price = listing.get("price_per_day", 0.0)
            
            bookings_map = listing_bookings.get(lid, {})
            b_count = len(bookings_map)
            p_count = sum(1 for status in bookings_map.values() if status == "PENDING")
            c_count = sum(1 for status in bookings_map.values() if status == "CONFIRMED")
            r_count = sum(1 for status in bookings_map.values() if status == "REJECTED")
            
            total_bookings += b_count
            total_pending += p_count
            total_confirmed += c_count
            total_rejected += r_count
            
            listings_stats.append(HostListingStats(
                listing_id=str(lid),
                title=title,
                price_per_day=price,
                bookings_count=b_count,
                pending_bookings=p_count,
                confirmed_bookings=c_count,
                rejected_bookings=r_count
            ))

        return HostAnalyticsResponse(
            host_id=host_id,
            total_listings=total_listings,
            total_bookings=total_bookings,
            total_pending=total_pending,
            total_confirmed=total_confirmed,
            total_rejected=total_rejected,
            listings=listings_stats
        )

    except Exception as e:
        logger.error(f"Error in host analytics for host {host_id}: {e}")
        raise HTTPException(status_code=500, detail="Failed to fetch host analytics")

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8008))
    uvicorn.run(app, host="0.0.0.0", port=port)
