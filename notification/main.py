import os
import logging
import threading
import urllib.request
import json
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
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
    Base.metadata.create_all(bind=postgres.engine)
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

BOOKING_SERVICE_URL = os.getenv("BOOKING_SERVICE_URL", "http://booking-service.raas.svc.cluster.local:80")
USER_SERVICE_URL = os.getenv("USER_SERVICE_URL", "http://user-service.raas.svc.cluster.local:80")

def fetch_booking_details(booking_id: str) -> dict:
    url = f"{BOOKING_SERVICE_URL}/bookings/{booking_id}"
    try:
        req = urllib.request.Request(url, headers={"Accept": "application/json"})
        with urllib.request.urlopen(req, timeout=5) as response:
            if response.status == 200:
                return json.loads(response.read().decode('utf-8'))
    except Exception as e:
        logger.error(f"Error fetching booking {booking_id}: {e}")
    return {}

def fetch_user_details(user_id: str) -> dict:
    url = f"{USER_SERVICE_URL}/users/{user_id}"
    try:
        req = urllib.request.Request(url, headers={"Accept": "application/json"})
        with urllib.request.urlopen(req, timeout=5) as response:
            if response.status == 200:
                return json.loads(response.read().decode('utf-8'))
    except Exception as e:
        logger.error(f"Error fetching user {user_id}: {e}")
    return {}

def send_email(to_email: str, subject: str, body_text: str) -> bool:
    smtp_host = os.getenv("SMTP_HOST", "smtp.gmail.com")
    smtp_port = int(os.getenv("SMTP_PORT", "587"))
    smtp_user = os.getenv("SMTP_USER")
    smtp_password = os.getenv("SMTP_PASSWORD")
    smtp_from = os.getenv("SMTP_FROM", smtp_user)
    
    if not smtp_user or not smtp_password:
        logger.warning("SMTP credentials not configured. Skipping email delivery.")
        return False
        
    try:
        msg = MIMEMultipart()
        msg['From'] = smtp_from
        msg['To'] = to_email
        msg['Subject'] = subject
        
        msg.attach(MIMEText(body_text, 'plain'))
        
        # Connect using STARTTLS
        server = smtplib.SMTP(smtp_host, smtp_port, timeout=10)
        server.starttls()
        server.login(smtp_user, smtp_password)
        server.sendmail(smtp_from, to_email, msg.as_string())
        server.quit()
        
        logger.info(f"Email successfully sent to {to_email}")
        return True
    except Exception as e:
        logger.error(f"Failed to send email to {to_email}: {e}")
        return False

def handle_notification_event(event: dict):
    """Handle incoming notification events"""
    try:
        session = postgres.SessionLocal()
        event_type = event.get("event") or event.get("event_type") or "unknown"
        
        # Only process booking.confirmed and booking.rejected
        if event_type not in ["booking.confirmed", "booking.rejected"]:
            logger.info(f"Ignoring event type: {event_type}")
            return True
            
        booking_id = event.get("booking_id")
        if not booking_id:
            logger.error(f"Missing booking_id in event: {event}")
            return False
            
        # Fetch booking to get guest_id
        booking = fetch_booking_details(booking_id)
        guest_id = booking.get("guest_id")
        if not guest_id:
            logger.error(f"Could not find guest_id for booking {booking_id}")
            return False
            
        # Fetch user to get email address
        user = fetch_user_details(guest_id)
        email = user.get("email")
        if not email:
            logger.error(f"Could not find email for user {guest_id}")
            return False
            
        # Compose message
        subject = ""
        message = ""
        if event_type == "booking.confirmed":
            subject = "Booking Confirmed - Rent as a Service"
            message = f"Hello,\n\nYour booking (ID: {booking_id}) has been CONFIRMED by the host! Enjoy your stay."
        elif event_type == "booking.rejected":
            subject = "Booking Rejected - Rent as a Service"
            message = f"Hello,\n\nUnfortunately, your booking (ID: {booking_id}) was REJECTED by the host. Please try booking another space or different dates."
            
        # Send the email
        email_sent = send_email(email, subject, message)
        status = "SENT" if email_sent else "FAILED"
        
        notification = NotificationLog(
            id=str(uuid.uuid4()),
            user_id=guest_id,
            message=message,
            notification_type="EMAIL",
            status=status
        )
        
        session.add(notification)
        session.commit()
        logger.info(f"Notification sent for event {event_type} to user {guest_id} with status {status}")
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
            "system.events"
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
