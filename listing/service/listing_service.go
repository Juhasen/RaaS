package service

import (
	"context"
	"errors"
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

// ListListings retrieves all Listings from the database.
func (s *ListingService) ListListings(ctx context.Context) ([]models.Listing, error) {
	return s.repo.ListListings(ctx, bson.M{})
}
