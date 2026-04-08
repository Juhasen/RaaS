package main

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Listing struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HostID      string             `json:"host_id" bson:"host_id"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	PricePerDay float64            `json:"price_per_day" bson:"price_per_day"`
	Status      string             `json:"status" bson:"status"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

func main() {
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/listings", createListing)
	e.GET("/listings/:id", getListing)

	e.Logger.Fatal(e.Start(":8080"))
}

func createListing(c echo.Context) error {
	var listing Listing
	if err := c.Bind(&listing); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request payload"})
	}

	listing.ID = primitive.NewObjectID()
	listing.CreatedAt = time.Now()
	if listing.Status == "" {
		listing.Status = "AVAILABLE" // Default status
	}

	if MongoClient != nil {
		collection := MongoClient.Database("raas").Collection("listings")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, listing)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create listing"})
		}
	} else {
		// Mock logic if connection not initiated for some reason
		c.Logger().Warn("MongoClient is nil, skipping mongodb insert")
	}

	return c.JSON(http.StatusCreated, listing)
}

func getListing(c echo.Context) error {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid id format"})
	}

	if MongoClient == nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{"error": "MongoClient is nil"})
	}

	collection := MongoClient.Database("raas").Collection("listings")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var listing Listing
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&listing)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "listing not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}

	return c.JSON(http.StatusOK, listing)
}
