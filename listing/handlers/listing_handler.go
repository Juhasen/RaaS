package handlers

import (
	"errors"
	"net/http"

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

// ListListings handles the GET /listings request, delegating to search if parameters match availability criteria.
func (h *ListingHandler) ListListings(c echo.Context) error {
	checkin := c.QueryParam("checkin")
	checkout := c.QueryParam("checkout")
	locationID := c.QueryParam("location_id")
	hostID := c.QueryParam("host_id")

	// If query parameters for search are present, delegate to availability handler
	if checkin != "" || checkout != "" || locationID != "" {
		return h.availabilityHandler.GetAvailableListings(c)
	}

	list, err := h.listingService.ListListings(c.Request().Context(), hostID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to list listings"})
	}

	return c.JSON(http.StatusOK, list)
}

