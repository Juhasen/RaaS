package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"booking/models"
	"booking/repository"
	"booking/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBookingService_Flow(t *testing.T) {
	pgURI := os.Getenv("PG_URI")
	if pgURI == "" {
		pgURI = "postgres://raas_user:raas_password@localhost:5432/raas_db"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, pgURI)
	if err != nil {
		t.Skipf("Skipping Postgres integration tests; connection failed at %s", pgURI)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skip("Skipping Postgres integration tests; ping failed")
		return
	}

	// Create test table if it doesn't exist to ensure test isolated environment runs smoothly
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS bookings (
			id VARCHAR(255) PRIMARY KEY,
			listing_id VARCHAR(255) NOT NULL,
			guest_id VARCHAR(255) NOT NULL,
			start_date VARCHAR(10) NOT NULL,
			end_date VARCHAR(10) NOT NULL,
			total_price DOUBLE PRECISION NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to prepare test table: %v", err)
	}

	// Clean up table
	_, _ = pool.Exec(ctx, "DELETE FROM bookings")

	repo := repository.NewPostgresRepository(pool)
	// We pass nil locker and writer to test core logic in database isolation
	bookingService := service.NewBookingService(repo, nil, nil)

	// Test booking creation
	today := time.Now()
	startDateStr := today.AddDate(0, 0, 3).Format("2006-01-02")
	endDateStr := today.AddDate(0, 0, 8).Format("2006-01-02")
	b := &models.Booking{
		ListingID:  "listing-test-1",
		GuestID:    "guest-test-1",
		StartDate:  startDateStr,
		EndDate:    endDateStr,
		TotalPrice: 500.0,
	}

	err = bookingService.CreateBooking(ctx, b)
	if err != nil {
		t.Fatalf("CreateBooking failed: %v", err)
	}

	if b.ID == "" || b.Status != "PENDING" {
		t.Errorf("Expected populated ID and PENDING status, got %+v", b)
	}

	// Test booking with past start date fails
	pastBooking := &models.Booking{
		ListingID:  "listing-test-1",
		GuestID:    "guest-test-1",
		StartDate:  today.AddDate(0, 0, -2).Format("2006-01-02"),
		EndDate:    today.AddDate(0, 0, 5).Format("2006-01-02"),
		TotalPrice: 500.0,
	}
	err = bookingService.CreateBooking(ctx, pastBooking)
	if err == nil {
		t.Error("Expected error for past start date, got nil")
	} else if err.Error() != "start date cannot be in the past" {
		t.Errorf("Expected 'start date cannot be in the past' error, got: %v", err)
	}

	// Test booking overlap check (same listing, overlapping dates)
	b2 := &models.Booking{
		ListingID:  "listing-test-1",
		GuestID:    "guest-test-2",
		StartDate:  today.AddDate(0, 0, 5).Format("2006-01-02"),
		EndDate:    today.AddDate(0, 0, 7).Format("2006-01-02"),
		TotalPrice: 200.0,
	}
	err = bookingService.CreateBooking(ctx, b2)
	if err == nil {
		t.Errorf("Expected date overlap error, got nil")
	}

	// Test listing bookings
	list, err := bookingService.ListBookings(ctx, "guest-test-1", "")
	if err != nil {
		t.Fatalf("ListBookings failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected 1 booking in list, got %d", len(list))
	}

	// Test payment succeeded event
	err = bookingService.ProcessPaymentEvent(ctx, "payment.succeeded", b.ID, b.ListingID)
	if err != nil {
		t.Fatalf("ProcessPaymentEvent succeeded failed: %v", err)
	}

	fetched, err := bookingService.GetBookingByID(ctx, b.ID)
	if err != nil {
		t.Fatalf("GetBookingByID failed: %v", err)
	}
	if fetched.Status != "CONFIRMED" {
		t.Errorf("Expected CONFIRMED status after payment success, got %s", fetched.Status)
	}

	// Test booking delete
	err = bookingService.DeleteBooking(ctx, b.ID)
	if err != nil {
		t.Fatalf("DeleteBooking failed: %v", err)
	}

	_, err = bookingService.GetBookingByID(ctx, b.ID)
	if err == nil {
		t.Errorf("Expected error fetching deleted booking, got nil")
	}
}
