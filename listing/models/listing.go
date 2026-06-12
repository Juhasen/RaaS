package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Listing represents a rentable unit or property in the system.
type Listing struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HostID        string             `json:"host_id" bson:"host_id"`
	Title         string             `json:"title" bson:"title"`
	Description   string             `json:"description" bson:"description"`
	PricePerDay   float64            `json:"price_per_day" bson:"price_per_day"`
	LocationID    string             `json:"location_id" bson:"location_id"`
	LocationLabel string             `json:"location_label" bson:"location_label"`
	Status        string             `json:"status" bson:"status"`
	MediaURLs     []string           `json:"media_urls" bson:"media_urls,omitempty"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
}
