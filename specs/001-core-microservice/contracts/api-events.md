# Service Contracts (Events & REST APIs)

## 1. Asynchronous Events (Kafka)

### Topic: user.created
- **Producer**: User Service
- **Consumers**: Notification Service, Analytics Service
- **Payload**: {"user_id": UUID, "email": String, "created_at": Timestamp}

### Topic: booking.created (Saga Started)
- **Producer**: Booking Service
- **Consumers**: Payment Service, Analytics Service
- **Payload**: {"booking_id": UUID, "guest_id": UUID, "amount": Decimal, "status": "PENDING"}

### Topic: payment.succeeded
- **Producer**: Payment Service
- **Consumers**: Booking Service, Notification Service, Analytics Service
- **Payload**: {"payment_id": UUID, "booking_id": UUID, "status": "SUCCESS"}

### Topic: payment.failed
- **Producer**: Payment Service
- **Consumers**: Booking Service, Notification Service, Analytics Service
- **Payload**: {"payment_id": UUID, "booking_id": UUID, "status": "FAILED", "reason": String}

### Topic: booking.confirmed (Saga Completed)
- **Producer**: Booking Service
- **Consumers**: Listing Service, Notification Service, Analytics Service
- **Payload**: {"booking_id": UUID, "status": "CONFIRMED"}

### Topic: booking.rejected (Saga Compensated)
- **Producer**: Booking Service
- **Consumers**: Notification Service, Analytics Service
- **Payload**: {"booking_id": UUID, "status": "REJECTED"}

### Topic: media.uploaded
- **Producer**: Media Service
- **Consumers**: Listing Service, Analytics Service
- **Payload**: {"media_id": UUID, "listing_id": ObjectId, "url": String}

## 2. API Endpoints

(Sample, to be implemented within each service)

- **User**: POST /api/users/register, POST /api/users/login, GET /api/users/{id}
- **Listing**: POST /api/listings, GET /api/listings, GET /api/listings/{id}
- **Media**: POST /api/media/upload, GET /api/media/listing/{listing_id}
- **Booking**: POST /api/bookings, GET /api/bookings/{id}, GET /api/bookings/user/{id}
- **Review**: POST /api/reviews, GET /api/reviews/listing/{id}
- **Favorites**: POST /api/favorites, DELETE /api/favorites/{listingId}, GET /api/favorites
- **Analytics**: GET /api/analytics/events, GET /api/analytics/dashboard

*All microservices implement GET /health.*
