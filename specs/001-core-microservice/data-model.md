# Data Model & Schema definitions

## 1. User Service (Python / Postgres)
- **User**: id (UUID), email (String), password_hash (String), ole (Enum), created_at (Timestamp)

## 2. Listing Service (Go / MongoDB)
- **Listing**: id (ObjectId), host_id (UUID), 	itle (String), description (String), price_per_day (Decimal), status (Enum), created_at (Timestamp)

## 3. Media Service (Go / S3/R2)
- **Media**: id (UUID), listing_id (ObjectId), url (String), 	ype (Enum: IMAGE, VIDEO), created_at (Timestamp)
*(Stores metadata locally or relies entirely on S3 keys structure, usually backed by Postgres/Mongo for fast lookups).*

## 4. Booking Service (Go / Postgres / Redis)
- **Booking**: id (UUID), listing_id (ObjectId), guest_id (UUID), start_date (Date), end_date (Date), 	otal_price (Decimal), status (Enum: PENDING, CONFIRMED, REJECTED), created_at (Timestamp)

## 5. Payment Service (Java / Postgres)
- **Transaction**: id (UUID), booking_id (UUID), mount (Decimal), status (Enum: SUCCESS, FAILED), payment_method (String), created_at (Timestamp)

## 6. Review Service (Java / Postgres)
- **Review**: id (UUID), booking_id (UUID), eviewer_id (UUID), ating (Int 1-5), comment (String), created_at (Timestamp)

## 7. Favorites Service (Java / Redis)
- **UserFavorites** (Redis Hash): Key: user:{uuid}:favorites, Values: list of listing_ids.

## 8. Notification Service (Python)
- **NotificationLog**: id (UUID), user_id (UUID), message (String), 	ype (Enum: EMAIL, SMS, PUSH), status (Enum: SENT, FAILED), created_at (Timestamp)

## 9. Analytics Service (Python / MongoDB/Postgres)
- **Event**: id (ObjectId/UUID), event_type (String), user_id (UUID), metadata (JSON), created_at (Timestamp)
