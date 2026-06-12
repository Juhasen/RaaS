package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"listing/models"
	"listing/repository"
	"listing/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAvailabilityService_BookingFlow(t *testing.T) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("Skipping integration test: MongoDB connection failed at %s", mongoURI)
		return
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("Skipping integration test: MongoDB ping failed")
		return
	}

	dbName := "raas_test"
	db := client.Database(dbName)
	_ = db.Collection("listings").Drop(ctx)
	_ = db.Collection("availability_blocks").Drop(ctx)
	_ = db.Collection("cancellation_tombstones").Drop(ctx)

	repo := repository.NewMongoRepository(client, dbName)
	listingService := service.NewListingService(repo)
	availService := service.NewAvailabilityService(repo)

	// Create test listing
	l := &models.Listing{
		HostID:        "host-test",
		Title:         "Test Cabin",
		Description:   "Cozy testing environment",
		PricePerDay:   120.0,
		LocationID:    "loc-test",
		LocationLabel: "Test Location, NY",
	}

	err = listingService.CreateListing(ctx, l)
	if err != nil {
		t.Fatalf("Failed to create listing: %v", err)
	}

	// 1. Query available listings (should contain our Test Cabin)
	// Search dates: tomorrow until 4 days from now
	tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	fourDaysOut := time.Now().AddDate(0, 0, 4).Truncate(24 * time.Hour)

	criteria := service.AvailabilityCriteria{
		Checkin:    tomorrow,
		Checkout:   fourDaysOut,
		LocationID: "loc-test",
	}
	results, err := availService.GetAvailableListings(ctx, criteria)
	if err != nil {
		t.Fatalf("GetAvailableListings failed: %v", err)
	}

	found := false
	for _, res := range results {
		if res.ListingID == l.ID.Hex() {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected listing %s to be available initially", l.ID.Hex())
	}

	// 2. Process booking confirmation (overlaps search range: tomorrow+1 to tomorrow+2)
	bookingID := "booking-test-1"
	startStr := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	endStr := time.Now().AddDate(0, 0, 3).Format("2006-01-02")
	err = availService.ProcessBookingConfirmed(ctx, l.ID.Hex(), bookingID, startStr, endStr)
	if err != nil {
		t.Fatalf("ProcessBookingConfirmed failed: %v", err)
	}

	// 3. Query availability again (should not find the cabin due to overlap)
	results, err = availService.GetAvailableListings(ctx, criteria)
	if err != nil {
		t.Fatalf("GetAvailableListings failed: %v", err)
	}

	found = false
	for _, res := range results {
		if res.ListingID == l.ID.Hex() {
			found = true
			break
		}
	}
	if found {
		t.Errorf("Expected listing %s to NOT be available after booking confirmation", l.ID.Hex())
	}

	// 4. Process booking cancellation
	err = availService.ProcessBookingCancelled(ctx, l.ID.Hex(), bookingID)
	if err != nil {
		t.Fatalf("ProcessBookingCancelled failed: %v", err)
	}

	// 5. Query availability again (should be available again)
	results, err = availService.GetAvailableListings(ctx, criteria)
	if err != nil {
		t.Fatalf("GetAvailableListings failed: %v", err)
	}

	found = false
	for _, res := range results {
		if res.ListingID == l.ID.Hex() {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected listing %s to be available again after booking cancellation", l.ID.Hex())
	}
}
