package handlers

import (
	"errors"
	"net/http"

	"booking/models"
	"booking/service"

	"github.com/labstack/echo/v4"
)

// BookingHandler processes HTTP requests for bookings.
type BookingHandler struct {
	bookingService *service.BookingService
}

// NewBookingHandler creates and returns a BookingHandler instance.
func NewBookingHandler(bs *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bs}
}

// CreateBooking handles the POST /bookings request.
func (h *BookingHandler) CreateBooking(c echo.Context) error {
	var b models.Booking
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid structure"})
	}

	if b.ListingID == "" || b.GuestID == "" || b.StartDate == "" || b.EndDate == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing required fields"})
	}

	err := h.bookingService.CreateBooking(c.Request().Context(), &b)
	if err != nil {
		if errors.Is(err, service.ErrLockField) {
			return c.JSON(http.StatusConflict, echo.Map{"error": err.Error()})
		}
		if errors.Is(err, service.ErrConflict) {
			return c.JSON(http.StatusConflict, echo.Map{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create booking"})
	}

	return c.JSON(http.StatusCreated, b)
}

// GetBooking handles the GET /bookings/:id request.
func (h *BookingHandler) GetBooking(c echo.Context) error {
	id := c.Param("id")
	b, err := h.bookingService.GetBookingByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "booking not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}

	return c.JSON(http.StatusOK, b)
}

// UpdateBooking handles the PUT /bookings/:id request.
func (h *BookingHandler) UpdateBooking(c echo.Context) error {
	id := c.Param("id")
	var b models.Booking
	if err := c.Bind(&b); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid structure"})
	}

	updated, err := h.bookingService.UpdateBooking(c.Request().Context(), id, &b)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "booking not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to update booking"})
	}

	return c.JSON(http.StatusOK, updated)
}

// DeleteBooking handles the DELETE /bookings/:id request.
func (h *BookingHandler) DeleteBooking(c echo.Context) error {
	id := c.Param("id")
	err := h.bookingService.DeleteBooking(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "booking not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to delete booking"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListBookings handles the GET /bookings request, with optional guest_id and listing_id filters.
func (h *BookingHandler) ListBookings(c echo.Context) error {
	guestID := c.QueryParam("guest_id")
	listingID := c.QueryParam("listing_id")

	list, err := h.bookingService.ListBookings(c.Request().Context(), guestID, listingID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to list bookings"})
	}

	return c.JSON(http.StatusOK, list)
}
