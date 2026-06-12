package repository

import (
	"context"
	"errors"
	"fmt"

	"booking/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository handles direct database queries using pgx v5.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository initializes and returns a PostgresRepository instance.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CheckOverlap counts existing bookings overlapping with the desired dates.
func (r *PostgresRepository) CheckOverlap(ctx context.Context, listingID, startDate, endDate, ignoreStatus string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT count(1) FROM bookings WHERE listing_id=$1 AND start_date < $2 AND end_date > $3 AND status != $4",
		listingID, endDate, startDate, ignoreStatus,
	).Scan(&count)
	return count, err
}

// CreateBooking inserts a new booking record.
func (r *PostgresRepository) CreateBooking(ctx context.Context, b *models.Booking) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO bookings (id, listing_id, guest_id, start_date, end_date, total_price, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		b.ID, b.ListingID, b.GuestID, b.StartDate, b.EndDate, b.TotalPrice, b.Status, b.CreatedAt,
	)
	return err
}

// GetBookingByID retrieves a single booking record by ID.
func (r *PostgresRepository) GetBookingByID(ctx context.Context, id string) (*models.Booking, error) {
	var b models.Booking
	err := r.pool.QueryRow(ctx,
		"SELECT id, listing_id, guest_id, start_date, end_date, total_price, status, created_at FROM bookings WHERE id=$1",
		id,
	).Scan(&b.ID, &b.ListingID, &b.GuestID, &b.StartDate, &b.EndDate, &b.TotalPrice, &b.Status, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &b, nil
}

// UpdateBooking updates a booking's properties.
func (r *PostgresRepository) UpdateBooking(ctx context.Context, b *models.Booking) error {
	res, err := r.pool.Exec(ctx,
		"UPDATE bookings SET listing_id=$1, guest_id=$2, start_date=$3, end_date=$4, total_price=$5, status=$6 WHERE id=$7",
		b.ListingID, b.GuestID, b.StartDate, b.EndDate, b.TotalPrice, b.Status, b.ID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// UpdateBookingStatus updates only the status of a booking.
func (r *PostgresRepository) UpdateBookingStatus(ctx context.Context, id, status string) error {
	res, err := r.pool.Exec(ctx,
		"UPDATE bookings SET status=$1 WHERE id=$2",
		status, id,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteBooking deletes a booking record from Postgres.
func (r *PostgresRepository) DeleteBooking(ctx context.Context, id string) error {
	res, err := r.pool.Exec(ctx,
		"DELETE FROM bookings WHERE id=$1",
		id,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ListBookings queries all bookings matching optional guest_id and listing_id filters.
func (r *PostgresRepository) ListBookings(ctx context.Context, guestID, listingID string) ([]models.Booking, error) {
	query := "SELECT id, listing_id, guest_id, start_date, end_date, total_price, status, created_at FROM bookings WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if guestID != "" {
		query += fmt.Sprintf(" AND guest_id=$%d", argIdx)
		args = append(args, guestID)
		argIdx++
	}
	if listingID != "" {
		query += fmt.Sprintf(" AND listing_id=$%d", argIdx)
		args = append(args, listingID)
		argIdx++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Booking
	for rows.Next() {
		var b models.Booking
		err = rows.Scan(&b.ID, &b.ListingID, &b.GuestID, &b.StartDate, &b.EndDate, &b.TotalPrice, &b.Status, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, b)
	}

	if list == nil {
		list = []models.Booking{}
	}
	return list, nil
}
