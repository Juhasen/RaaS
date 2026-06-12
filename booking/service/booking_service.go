package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"booking/models"
	"booking/repository"

	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"
)

var (
	ErrNotFound  = errors.New("booking not found")
	ErrConflict  = errors.New("booking dates overlap with an existing booking")
	ErrLockField = errors.New("could not acquire lock for listing, please try again")
)

// BookingService coordinates booking processes, checks dates, locks listing blocks, and emits Kafka events.
type BookingService struct {
	repo        *repository.PostgresRepository
	rs          *redsync.Redsync
	kafkaWriter *kafka.Writer
}

// NewBookingService initializes and returns a BookingService instance.
func NewBookingService(repo *repository.PostgresRepository, rs *redsync.Redsync, kw *kafka.Writer) *BookingService {
	return &BookingService{
		repo:        repo,
		rs:          rs,
		kafkaWriter: kw,
	}
}

// CreateBooking creates a new booking request checking for date collisions and holding listing locks.
func (s *BookingService) CreateBooking(ctx context.Context, b *models.Booking) error {
	// 1. Acquire distributed lock using Redis/Redsync if configured
	if s.rs != nil {
		mutexname := fmt.Sprintf("listing-lock-%s", b.ListingID)
		mutex := s.rs.NewMutex(mutexname, redsync.WithExpiry(8*time.Second))
		if err := mutex.LockContext(ctx); err != nil {
			return ErrLockField
		}
		defer func() {
			_, _ = mutex.UnlockContext(ctx)
		}()
	}

	b.ID = uuid.NewString()
	b.Status = "PENDING"
	b.CreatedAt = time.Now()

	// 2. Check for date collisions with existing confirmed/pending bookings
	conflictCount, err := s.repo.CheckOverlap(ctx, b.ListingID, b.StartDate, b.EndDate, "REJECTED")
	if err != nil {
		return err
	}
	if conflictCount > 0 {
		return ErrConflict
	}

	// 3. Persist the booking details to database
	err = s.repo.CreateBooking(ctx, b)
	if err != nil {
		return err
	}

	// 4. Emit the booking.created event to Kafka
	if s.kafkaWriter != nil {
		eventMsg := fmt.Sprintf(`{"event":"booking.created","booking_id":"%s","listing_id":"%s","guest_id":"%s","total_price":%f}`,
			b.ID, b.ListingID, b.GuestID, b.TotalPrice,
		)
		errEvt := s.kafkaWriter.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(b.ID),
				Value: []byte(eventMsg),
			},
		)
		if errEvt != nil {
			log.Printf("Failed to emit booking.created event: %v", errEvt)
		} else {
			log.Printf("Emitted booking.created event for booking %s", b.ID)
		}
	}

	return nil
}

// GetBookingByID retrieves a single booking request by its ID.
func (s *BookingService) GetBookingByID(ctx context.Context, id string) (*models.Booking, error) {
	b, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

// UpdateBooking updates booking details keeping identity intact.
func (s *BookingService) UpdateBooking(ctx context.Context, id string, updated *models.Booking) (*models.Booking, error) {
	existing, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Preserve immutable properties
	updated.ID = existing.ID
	updated.CreatedAt = existing.CreatedAt
	if updated.Status == "" {
		updated.Status = existing.Status
	}

	err = s.repo.UpdateBooking(ctx, updated)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

// DeleteBooking deletes a booking from Postgres.
func (s *BookingService) DeleteBooking(ctx context.Context, id string) error {
	err := s.repo.DeleteBooking(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// ListBookings queries all booking records matching parameters.
func (s *BookingService) ListBookings(ctx context.Context, guestID, listingID string) ([]models.Booking, error) {
	return s.repo.ListBookings(ctx, guestID, listingID)
}

// ProcessPaymentEvent reacts to payment status change events (confirming or rejecting the booking).
func (s *BookingService) ProcessPaymentEvent(ctx context.Context, eventName, bookingID, listingID string) error {
	var targetStatus string
	var sagaEvent string

	if eventName == "payment.succeeded" {
		targetStatus = "CONFIRMED"
		sagaEvent = "booking.confirmed"
	} else if eventName == "payment.failed" {
		targetStatus = "REJECTED"
		sagaEvent = "booking.rejected"
	} else {
		return fmt.Errorf("unknown payment event: %s", eventName)
	}

	err := s.repo.UpdateBookingStatus(ctx, bookingID, targetStatus)
	if err != nil {
		return err
	}

	log.Printf("Payment status updated to %s for booking %s", targetStatus, bookingID)

	if s.kafkaWriter != nil {
		eventMsg := fmt.Sprintf(`{"event":"%s","booking_id":"%s","listing_id":"%s"}`, sagaEvent, bookingID, listingID)
		errEvt := s.kafkaWriter.WriteMessages(ctx,
			kafka.Message{
				Key:   []byte(bookingID),
				Value: []byte(eventMsg),
			},
		)
		if errEvt != nil {
			return fmt.Errorf("failed to emit %s saga event: %w", sagaEvent, errEvt)
		}
		log.Printf("Emitted saga event %s for booking %s", sagaEvent, bookingID)
	}

	return nil
}
