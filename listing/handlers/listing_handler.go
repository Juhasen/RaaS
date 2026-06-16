package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"listing/models"
	"listing/service"

	"github.com/labstack/echo/v4"
)

// ListingHandler processes Listing HTTP requests.
type ListingHandler struct {
	listingService      *service.ListingService
	availabilityHandler *AvailabilityHandler
}

// NewListingHandler creates and returns a ListingHandler instance.
func NewListingHandler(ls *service.ListingService, ah *AvailabilityHandler) *ListingHandler {
	return &ListingHandler{
		listingService:      ls,
		availabilityHandler: ah,
	}
}

// CreateListing handles the POST /listings request.
func (h *ListingHandler) CreateListing(c echo.Context) error {
	var l models.Listing
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request payload"})
	}

	if err := h.listingService.CreateListing(c.Request().Context(), &l); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create listing"})
	}

	return c.JSON(http.StatusCreated, l)
}

// GetListing handles the GET /listings/:id request.
func (h *ListingHandler) GetListing(c echo.Context) error {
	id := c.Param("id")
	l, err := h.listingService.GetListingByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
		}
		if errors.Is(err, service.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "listing not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}

	return c.JSON(http.StatusOK, l)
}

// UpdateListing handles the PUT /listings/:id request.
func (h *ListingHandler) UpdateListing(c echo.Context) error {
	id := c.Param("id")
	var l models.Listing
	if err := c.Bind(&l); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request payload"})
	}

	updated, err := h.listingService.UpdateListing(c.Request().Context(), id, &l)
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
		}
		if errors.Is(err, service.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "listing not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to update listing"})
	}

	return c.JSON(http.StatusOK, updated)
}

// DeleteListing handles the DELETE /listings/:id request.
func (h *ListingHandler) DeleteListing(c echo.Context) error {
	id := c.Param("id")
	err := h.listingService.DeleteListing(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidID) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
		}
		if errors.Is(err, service.ErrNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "listing not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to delete listing"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListListings handles the GET /listings request, supporting full combination of filters.
func (h *ListingHandler) ListListings(c echo.Context) error {
	checkin := c.QueryParam("checkin")
	if checkin == "" {
		checkin = c.QueryParam("start_date")
	}
	checkout := c.QueryParam("checkout")
	if checkout == "" {
		checkout = c.QueryParam("end_date")
	}
	location := c.QueryParam("location")
	if location == "" {
		location = c.QueryParam("location_id")
	}
	name := c.QueryParam("name")
	if name == "" {
		name = c.QueryParam("q")
	}
	hostID := c.QueryParam("host_id")

	page := int64(1)
	if v := c.QueryParam("page"); v != "" {
		if p, err := strconv.ParseInt(v, 10, 64); err == nil && p > 0 {
			page = p
		}
	}
	limit := int64(20)
	if v := c.QueryParam("limit"); v != "" {
		if l, err := strconv.ParseInt(v, 10, 64); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 100 {
		limit = 100
	}

	// Validate date range if checkin and checkout are provided
	if checkin != "" && checkout != "" {
		checkinTime, err := time.Parse("2006-01-02", checkin)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid checkin date format"})
		}
		checkoutTime, err := time.Parse("2006-01-02", checkout)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid checkout date format"})
		}
		if !checkinTime.Before(checkoutTime) {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "checkin must be before checkout"})
		}
	}

	list, total, err := h.listingService.ListListings(c.Request().Context(), hostID, checkin, checkout, location, name, page, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to list listings"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data":  list,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
