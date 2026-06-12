package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AvailabilityBlock represents a single day blocked for a listing due to a confirmed booking.
type AvailabilityBlock struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ListingID primitive.ObjectID `bson:"listing_id"`
	Date      time.Time          `bson:"date"`
	BookingID string             `bson:"booking_id"`
	CreatedAt time.Time          `bson:"created_at"`
}

// CancellationTombstone registers out-of-order cancellations to prevent late bookings from blocking availability.
type CancellationTombstone struct {
	BookingID string    `bson:"booking_id"`
	CreatedAt time.Time `bson:"created_at"`
}
