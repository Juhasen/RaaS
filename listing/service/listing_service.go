package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"listing/models"
	"listing/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrInvalidID = errors.New("invalid ID format")
	ErrNotFound  = errors.New("listing not found")
)

// ListingService orchestrates listing-related operations.
type ListingService struct {
	repo *repository.MongoRepository
}

// NewListingService creates and returns a ListingService instance.
func NewListingService(repo *repository.MongoRepository) *ListingService {
	return &ListingService{repo: repo}
}

// CreateListing prepares and stores a new Listing.
func (s *ListingService) CreateListing(ctx context.Context, l *models.Listing) error {
	l.ID = primitive.NewObjectID()
	l.CreatedAt = time.Now()
	if l.Status == "" {
		l.Status = "AVAILABLE"
	}
	return s.repo.CreateListing(ctx, l)
}

// GetListingByID fetches a Listing by its string hexadecimal ID.
func (s *ListingService) GetListingByID(ctx context.Context, idStr string) (*models.Listing, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInvalidID
	}
	l, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return l, nil
}

// UpdateListing replaces fields on an existing Listing, keeping ID and CreatedAt intact.
func (s *ListingService) UpdateListing(ctx context.Context, idStr string, updated *models.Listing) (*models.Listing, error) {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInvalidID
	}
	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Preserve identity and creation timestamp
	updated.ID = existing.ID
	updated.CreatedAt = existing.CreatedAt
	if updated.Status == "" {
		updated.Status = existing.Status
	}

	err = s.repo.UpdateListing(ctx, updated)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteListing deletes a Listing by its string hexadecimal ID.
func (s *ListingService) DeleteListing(ctx context.Context, idStr string) error {
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ErrInvalidID
	}
	err = s.repo.DeleteListing(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ListListings retrieves Listings matching the optional filters.
func (s *ListingService) ListListings(ctx context.Context, hostID, checkinStr, checkoutStr, location, name string) ([]models.Listing, error) {
	var filterParts []bson.M

	if hostID != "" {
		filterParts = append(filterParts, bson.M{"host_id": hostID})
	}

	if checkinStr != "" && checkoutStr != "" {
		blockedIDs, err := s.getBlockedListingIDsFromBookingService(ctx, checkinStr, checkoutStr)
		if err != nil {
			log.Printf("Booking service unavailable, falling back to local MongoDB blocks: %v", err)
			start, errStart := time.Parse("2006-01-02", checkinStr)
			end, errEnd := time.Parse("2006-01-02", checkoutStr)
			if errStart == nil && errEnd == nil {
				var mongoErr error
				blockedIDs, mongoErr = s.repo.GetBlockedListingIDs(ctx, start, end)
				if mongoErr != nil {
					log.Printf("Failed to get local MongoDB blocked listing IDs: %v", mongoErr)
				}
			}
		}

		if len(blockedIDs) > 0 {
			filterParts = append(filterParts, bson.M{"_id": bson.M{"$nin": blockedIDs}})
		}
	}

	if location != "" {
		filterParts = append(filterParts, bson.M{"$or": []bson.M{
			{"location_label": bson.M{"$regex": location, "$options": "i"}},
			{"location_id": bson.M{"$regex": location, "$options": "i"}},
		}})
	}

	if name != "" {
		filterParts = append(filterParts, bson.M{"$or": []bson.M{
			{"title": bson.M{"$regex": name, "$options": "i"}},
			{"description": bson.M{"$regex": name, "$options": "i"}},
		}})
	}

	var filter bson.M
	if len(filterParts) > 0 {
		filter = bson.M{"$and": filterParts}
	} else {
		filter = bson.M{}
	}

	return s.repo.ListListings(ctx, filter)
}

func (s *ListingService) getBlockedListingIDsFromBookingService(ctx context.Context, start, end string) ([]primitive.ObjectID, error) {
	bookingURL := os.Getenv("BOOKING_SERVICE_URL")
	if bookingURL == "" {
		bookingURL = "http://booking-service"
	}
	url := fmt.Sprintf("%s/bookings/active?start_date=%s&end_date=%s", bookingURL, start, end)

	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var bookings []struct {
		ListingID string `json:"listing_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&bookings); err != nil {
		return nil, err
	}

	var blockedIDs []primitive.ObjectID
	for _, b := range bookings {
		if b.ListingID != "" {
			if oid, err := primitive.ObjectIDFromHex(b.ListingID); err == nil {
				blockedIDs = append(blockedIDs, oid)
			}
		}
	}
	return blockedIDs, nil
}

