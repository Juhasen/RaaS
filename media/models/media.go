package models

import "time"

// Media represents an uploaded media asset (e.g. property image) metadata.
type Media struct {
	ID        string    `json:"id" bson:"_id"`
	ListingID string    `json:"listing_id" bson:"listing_id"`
	URL       string    `json:"url" bson:"url"`
	Type      string    `json:"type" bson:"type"` // e.g. "IMAGE"
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
