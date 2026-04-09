package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Media struct {
	ID        string    `json:"id"`
	ListingID string    `json:"listing_id"`
	URL       string    `json:"url"`
	Type      string    `json:"type"` // IMAGE, VIDEO
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/media", uploadMedia)
	e.GET("/media/:id", getMedia)

	e.Logger.Fatal(e.Start(":8080"))
}

func uploadMedia(c echo.Context) error {
	listingID := c.FormValue("listing_id")
	if listingID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "listing_id is required"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "file is required"})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	defer src.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	id := uuid.New().String()
	key := fmt.Sprintf("%s/%s-%s", listingID, id, file.Filename)

	mediaType := "IMAGE"
	// simplified type check for MVP
	if file.Header.Get("Content-Type") == "video/mp4" {
		mediaType = "VIDEO"
	}

	// Upload to S3 if configured
	if S3Client != nil {
		_, err = S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String("raas-media-bucket"),
			Key:    aws.String(key),
			Body:   bytes.NewReader(buf.Bytes()),
			ContentType: aws.String(file.Header.Get("Content-Type")),
		})
		if err != nil {
			c.Logger().Errorf("Failed to upload to S3: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to upload file"})
		}
	} else {
		c.Logger().Warn("S3Client is nil, skipping actual upload")
	}

	media := Media{
		ID:        id,
		ListingID: listingID,
		URL:       fmt.Sprintf("https://s3.amazonaws.com/raas-media-bucket/%s", key),
		Type:      mediaType,
		CreatedAt: time.Now(),
	}

	if KafkaWriter != nil {
		eventMsg := fmt.Sprintf(`{"event":"media.uploaded","media_id":"%s","listing_id":"%s","url":"%s","type":"%s"}`, media.ID, media.ListingID, media.URL, media.Type)
		err := KafkaWriter.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(media.ID),
				Value: []byte(eventMsg),
			},
		)
		if err != nil {
			c.Logger().Errorf("Failed to emit media.uploaded event: %v", err)
			// Continue since media was uploaded successfully
		} else {
			c.Logger().Info("Emitted media.uploaded event")
		}
	} else {
		c.Logger().Warn("KafkaWriter not initialized, skipping event emission")
	}

	return c.JSON(http.StatusCreated, media)
}

func getMedia(c echo.Context) error {
	id := c.Param("id")
	// Since we are not storing media metadata in a DB for the MVP, we just return a placeholder.
	// In a real app this would query Mongo or Postgres.
	return c.JSON(http.StatusOK, echo.Map{
		"id": id,
		"message": "Media metadata retrieval normally requires a database. Fetching directly from S3 requires listing_id/key.",
	})
}

