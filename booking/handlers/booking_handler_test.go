package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"booking/handlers"
	"booking/models"
	"booking/repository"
	"booking/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func TestBookingHandler_CRUD(t *testing.T) {
	pgURI := os.Getenv("PG_URI")
	if pgURI == "" {
		pgURI = "postgres://raas_user:raas_password@localhost:5432/raas_db"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, pgURI)
	if err != nil {
		t.Skip("Skipping handler test: Postgres connection failed")
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skip("Skipping handler test: Postgres ping failed")
		return
	}

	// Create test table if it doesn't exist
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

	// Clean up
	_, _ = pool.Exec(ctx, "DELETE FROM bookings")

	repo := repository.NewPostgresRepository(pool)
	bs := service.NewBookingService(repo, nil, nil)
	h := handlers.NewBookingHandler(bs)

	e := echo.New()

	// 1. Test Create Booking
	payload := `{"listing_id":"list-h","guest_id":"guest-h","start_date":"2026-07-01","end_date":"2026-07-05","total_price":400.0}`
	req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateBooking(c); err != nil {
		t.Fatalf("CreateBooking failed: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	var created models.Booking
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("Failed to unmarshal created booking: %v", err)
	}
	if created.ListingID != "list-h" || created.ID == "" {
		t.Errorf("Created booking details mismatch: %+v", created)
	}

	// 2. Test Get Booking
	req = httptest.NewRequest(http.MethodGet, "/bookings/"+created.ID, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID)

	if err := h.GetBooking(c); err != nil {
		t.Fatalf("GetBooking failed: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var fetched models.Booking
	if err := json.Unmarshal(rec.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("Failed to unmarshal fetched booking: %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("Fetched booking ID mismatch: %s vs %s", fetched.ID, created.ID)
	}

	// 3. Test Update Booking
	updatePayload := `{"listing_id":"list-h","guest_id":"guest-h","start_date":"2026-07-01","end_date":"2026-07-05","total_price":450.0,"status":"CONFIRMED"}`
	req = httptest.NewRequest(http.MethodPut, "/bookings/"+created.ID, strings.NewReader(updatePayload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID)

	if err := h.UpdateBooking(c); err != nil {
		t.Fatalf("UpdateBooking failed: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var updated models.Booking
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("Failed to unmarshal updated booking: %v", err)
	}
	if updated.TotalPrice != 450.0 || updated.Status != "CONFIRMED" {
		t.Errorf("Updated booking details mismatch: %+v", updated)
	}

	// 4. Test Delete Booking
	req = httptest.NewRequest(http.MethodDelete, "/bookings/"+created.ID, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID)

	if err := h.DeleteBooking(c); err != nil {
		t.Fatalf("DeleteBooking failed: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rec.Code)
	}

	// 5. Test Get Deleted (should be 404)
	req = httptest.NewRequest(http.MethodGet, "/bookings/"+created.ID, nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(created.ID)

	if err := h.GetBooking(c); err != nil {
		t.Fatalf("GetBooking failed: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for deleted booking, got %d", rec.Code)
	}
}
