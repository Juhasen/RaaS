package handlers

import (
	"net/http"

	"github.com/Juhasen/RaaS/media/models"
	"github.com/Juhasen/RaaS/media/service"
	"github.com/labstack/echo/v4"
)

// MediaHandler processes HTTP requests for media upload and retrieval.
type MediaHandler struct {
	srv *service.MediaService
}

// NewMediaHandler creates a new MediaHandler instance.
func NewMediaHandler(srv *service.MediaService) *MediaHandler {
	return &MediaHandler{srv: srv}
}

// UploadMedia handles uploading property pictures.
func (h *MediaHandler) UploadMedia(c echo.Context) error {
	listingID := c.FormValue("listing_id")
	if listingID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "listing_id is required"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "file is required"})
	}

	media, err := h.srv.UploadMedia(c.Request().Context(), listingID, file)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, media)
}

// GetMedia retrieves metadata for a single media asset.
func (h *MediaHandler) GetMedia(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "id parameter is required"})
	}

	media, err := h.srv.FindByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	if media == nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "media not found"})
	}

	return c.JSON(http.StatusOK, media)
}

// GetListingMedia retrieves all media metadata for a specific listing.
func (h *MediaHandler) GetListingMedia(c echo.Context) error {
	listingID := c.Param("listing_id")
	if listingID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "listing_id parameter is required"})
	}

	list, err := h.srv.FindByListingID(c.Request().Context(), listingID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	if list == nil {
		list = make([]*models.Media, 0)
	}

	return c.JSON(http.StatusOK, list)
}
