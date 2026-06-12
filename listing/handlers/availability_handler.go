package handlers

import (
	"fmt"
	"net/http"
	"time"

	"listing/service"

	"github.com/labstack/echo/v4"
)

// AvailabilityHandler handles HTTP requests relating to available listings.
type AvailabilityHandler struct {
	availService *service.AvailabilityService
}

// NewAvailabilityHandler creates and returns an AvailabilityHandler instance.
func NewAvailabilityHandler(as *service.AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{availService: as}
}

// GetAvailableListings parses query parameters, validates them, and queries matching listings.
func (h *AvailabilityHandler) GetAvailableListings(c echo.Context) error {
	checkinStr := c.QueryParam("checkin")
	checkoutStr := c.QueryParam("checkout")
	locationID := c.QueryParam("location_id")
	minPriceStr := c.QueryParam("min_price")
	maxPriceStr := c.QueryParam("max_price")

	if checkinStr == "" || checkoutStr == "" || locationID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "checkin, checkout and location_id are required"})
	}

	checkin, err := time.Parse("2006-01-02", checkinStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid checkin date format"})
	}
	checkout, err := time.Parse("2006-01-02", checkoutStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid checkout date format"})
	}
	if !checkin.Before(checkout) {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "checkin must be before checkout"})
	}

	// Calculate nights
	nights := int(checkout.Sub(checkin).Hours() / 24)
	if nights > 30 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "search range cannot exceed 30 nights"})
	}

	var minPricePtr, maxPricePtr *float64

	if minPriceStr != "" {
		var minPrice float64
		if _, err := fmt.Sscanf(minPriceStr, "%f", &minPrice); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid min_price"})
		}
		minPricePtr = &minPrice
	}

	if maxPriceStr != "" {
		var maxPrice float64
		if _, err := fmt.Sscanf(maxPriceStr, "%f", &maxPrice); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid max_price"})
		}
		maxPricePtr = &maxPrice
	}

	if minPricePtr != nil && maxPricePtr != nil && *minPricePtr > *maxPricePtr {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "min_price cannot be greater than max_price"})
	}

	criteria := service.AvailabilityCriteria{
		Checkin:    checkin,
		Checkout:   checkout,
		LocationID: locationID,
		MinPrice:   minPricePtr,
		MaxPrice:   maxPricePtr,
	}

	results, err := h.availService.GetAvailableListings(c.Request().Context(), criteria)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to search available listings"})
	}

	return c.JSON(http.StatusOK, results)
}
