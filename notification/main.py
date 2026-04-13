import os
import logging
import threading
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session
import uuid

from models import NotificationLog, Base
from schemas import NotificationLogResponse
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

app = FastAPI(title="Notification Service", version="1.0.0")

# Dependency to get database session
def get_db():
    session = postgres.SessionLocal()
    try:
        yield session
    finally:
        session.close()

# Kafka consumer setup
kafka_consumer = None
consumer_thread = None

def handle_notification_event(event: dict):
    """Handle incoming notification events"""
    try:
        session = postgres.SessionLocal()
        event_type = event.get("event_type", "unknown")
        user_id = event.get("user_id")
        
        # Create appropriate notification based on event type
        message = ""
        if event_type == "user.created":
            message = f"Welcome to RaaS! Your account has been created."
        elif event_type == "payment.succeeded":
            message = f"Payment successful. Your booking is being processed."
        elif event_type == "payment.failed":
            message = f"Payment failed. Please try again or contact support."
        elif event_type == "booking.confirmed":
            message = f"Your booking has been confirmed!"
        elif event_type == "booking.rejected":
            message = f"Your booking was rejected. Please try another date."
        else:
            message = f"Event notification: {event_type}"
        
        notification = NotificationLog(
            id=str(uuid.uuid4()),
            user_id=user_id,
            message=message,
            notification_type="EMAIL",
            status="SENT"
        )
        
        session.add(notification)
        session.commit()
        logger.info(f"Notification sent for event {event_type} to user {user_id}")
        return True
        
    except Exception as e:
        logger.error(f"Error handling notification event: {e}")
        return False
    finally:
        session.close()

def start_kafka_consumer():
    """Start Kafka consumer in background thread"""
    global kafka_consumer
    try:
        topics = [
            "user.created",
            "payment.succeeded", 
            "payment.failed",
            "booking.confirmed",
            "booking.rejected"
        ]
        kafka_consumer = KafkaConsumerWrapper(
            group_id="notification-service-group",
            topics=topics
        )
        kafka_consumer.consume_messages(handle_notification_event)
    except Exception as e:
        logger.error(f"Kafka consumer error: {e}")

@app.on_event("startup")
async def startup_event():
    """Start Kafka consumer on app startup"""
    global consumer_thread
    consumer_thread = threading.Thread(target=start_kafka_consumer, daemon=True)
    consumer_thread.start()
    logger.info("Kafka consumer started")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on app shutdown"""
    if kafka_consumer:
        kafka_consumer.close()
    logger.info("Kafka consumer closed")

# Health check endpoint
@app.get("/health")
async def health():
    """Health check endpoint for Kubernetes probes"""
    return {
        "status": "healthy",
        "service": "notification-service",
        "version": "1.0.0"
    }

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "Notification Service is running"}

# Get notification log endpoint
@app.get("/notifications/{log_id}", response_model=NotificationLogResponse)
async def get_notification(log_id: str, db: Session = Depends(get_db)):
    """Get notification log by ID"""
    try:
        notification = db.query(NotificationLog).filter(NotificationLog.id == log_id).first()
        if not notification:
            raise HTTPException(status_code=404, detail="Notification not found")
        return notification
    except Exception as e:
        logger.error(f"Error retrieving notification: {e}")
        raise HTTPException(status_code=500, detail="Failed to retrieve notification")

# List notifications for user
@app.get("/notifications/user/{user_id}")
async def list_user_notifications(
    user_id: str,
    skip: int = 0,
    limit: int = 10,
    db: Session = Depends(get_db)
):
    """List notifications for a specific user"""
    try:
        total = db.query(NotificationLog).filter(NotificationLog.user_id == user_id).count()
        notifications = db.query(NotificationLog).filter(
            NotificationLog.user_id == user_id
        ).order_by(NotificationLog.created_at.desc()).offset(skip).limit(limit).all()
        
        return {
            "total": total,
            "count": len(notifications),
            "items": notifications
        }
    except Exception as e:
        logger.error(f"Error listing notifications: {e}")
        raise HTTPException(status_code=500, detail="Failed to list notifications")

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8003))
    uvicorn.run(app, host="0.0.0.0", port=port)
