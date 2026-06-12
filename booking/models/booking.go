package models

import "time"

// Booking represents a reservation request in RaaS.
type Booking struct {
	ID         string    `json:"id"`
	ListingID  string    `json:"listing_id"`
	GuestID    string    `json:"guest_id"`
	StartDate  string    `json:"start_date"` // YYYY-MM-DD
	EndDate    string    `json:"end_date"`   // YYYY-MM-DD
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"` // PENDING, CONFIRMED, REJECTED
	CreatedAt  time.Time `json:"created_at"`
}
