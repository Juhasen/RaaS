package service

import (
	"context"
	"time"

	"listing/models"
	"listing/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AvailabilityCriteria specifies the criteria when querying available listings.
type AvailabilityCriteria struct {
	Checkin    time.Time
	Checkout   time.Time
	LocationID string
	MinPrice   *float64
	MaxPrice   *float64
}

// AvailableListingResponse is the format returned by availability search.
type AvailableListingResponse struct {
	ListingID     string  `json:"listing_id"`
	DisplayName   string  `json:"display_name"`
	LocationLabel string  `json:"location_label"`
	BasePrice     float64 `json:"base_price"`
}

// AvailabilityService handles availability search and booking event streams.
type AvailabilityService struct {
	repo *repository.MongoRepository
}

// NewAvailabilityService creates and returns an AvailabilityService instance.
func NewAvailabilityService(repo *repository.MongoRepository) *AvailabilityService {
	return &AvailabilityService{repo: repo}
}

// GetAvailableListings finds all listings matching the criteria that do not have overlapping blocks.
func (s *AvailabilityService) GetAvailableListings(ctx context.Context, criteria AvailabilityCriteria) ([]AvailableListingResponse, error) {
	filter := bson.M{"location_id": criteria.LocationID}

	if criteria.MinPrice != nil && criteria.MaxPrice != nil {
		filter["price_per_day"] = bson.M{"$gte": *criteria.MinPrice, "$lte": *criteria.MaxPrice}
	} else if criteria.MinPrice != nil {
		filter["price_per_day"] = bson.M{"$gte": *criteria.MinPrice}
	} else if criteria.MaxPrice != nil {
		filter["price_per_day"] = bson.M{"$lte": *criteria.MaxPrice}
	}

	listings, err := s.repo.ListListings(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []AvailableListingResponse
	for _, l := range listings {
		count, err := s.repo.CountAvailabilityBlocks(ctx, l.ID, criteria.Checkin, criteria.Checkout)
		if err != nil {
			continue
		}
		if count == 0 {
			results = append(results, AvailableListingResponse{
				ListingID:     l.ID.Hex(),
				DisplayName:   l.Title,
				LocationLabel: l.LocationLabel,
				BasePrice:     l.PricePerDay,
			})
		}
	}

	if results == nil {
		results = []AvailableListingResponse{}
	}
	return results, nil
}

// ProcessBookingConfirmed processes booking.confirmed events.
func (s *AvailabilityService) ProcessBookingConfirmed(ctx context.Context, listingIDHex, bookingID, startStr, endStr string) error {
	listingID, err := primitive.ObjectIDFromHex(listingIDHex)
	if err != nil {
		return err
	}
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return err
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return err
	}

	// If there's an out-of-order cancellation tombstone for this booking, we ignore the confirmation
	// so that availability is not blocked, and clean up the tombstone.
	tombstoneCount, err := s.repo.DeleteTombstoneByBooking(ctx, bookingID)
	if err == nil && tombstoneCount > 0 {
		return nil
	}

	// Write availability blocks for each day in [start, end)
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		block := &models.AvailabilityBlock{
			ID:        primitive.NewObjectID(),
			ListingID: listingID,
			Date:      d,
			BookingID: bookingID,
			CreatedAt: time.Now(),
		}
		err := s.repo.UpsertAvailabilityBlock(ctx, block)
		if err != nil {
			return err
		}
	}

	return nil
}

// ProcessBookingCancelled processes booking.cancelled events.
func (s *AvailabilityService) ProcessBookingCancelled(ctx context.Context, listingIDHex, bookingID string) error {
	deletedCount, err := s.repo.DeleteAvailabilityBlocksByBooking(ctx, bookingID)
	if err != nil {
		return err
	}

	// If no blocks were deleted, it's an out-of-order cancellation. We record a tombstone.
	if deletedCount == 0 {
		tombstone := &models.CancellationTombstone{
			BookingID: bookingID,
			CreatedAt: time.Now(),
		}
		err := s.repo.CreateTombstone(ctx, tombstone)
		if err != nil {
			return err
		}
	}

	return nil
}
