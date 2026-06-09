package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Listing struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HostID        string             `json:"host_id" bson:"host_id"`
	Title         string             `json:"title" bson:"title"`
	Description   string             `json:"description" bson:"description"`
	PricePerDay   float64            `json:"price_per_day" bson:"price_per_day"`
	LocationID    string             `json:"location_id" bson:"location_id"`
	LocationLabel string             `json:"location_label" bson:"location_label"`
	Status        string             `json:"status" bson:"status"`
	MediaURLs     []string           `json:"media_urls" bson:"media_urls,omitempty"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
}

func main() {
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/listings", createListing)
	e.GET("/listings/:id", getListing)

	go startKafkaConsumers()

	e.Logger.Fatal(e.Start(":8080"))
}

func startKafkaConsumers() {
	log.Println("Starting Listing Kafka consumers (mocked without actual brokers for MVP)")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "listing-service-group",
		Topic:   "system.events",
		MaxWait: 1 * time.Second,
	})
	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		var evt struct {
			Event     string `json:"event"`
			ListingID string `json:"listing_id"`
			BookingID string `json:"booking_id"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}
		if err := json.Unmarshal(m.Value, &evt); err == nil && MongoClient != nil {
			// Handle media.uploaded separately if necessary
			if evt.Event == "media.uploaded" && evt.ListingID != "" {
				objID, errID := primitive.ObjectIDFromHex(evt.ListingID)
				if errID == nil {
					collection := MongoClient.Database("raas").Collection("listings")
					collection.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$addToSet": bson.M{"media_urls": evt.BookingID}})
				}
			}

			// Booking events
			if evt.Event == "booking.confirmed" {
				go func(e interface{}) {
					_ = processBookingConfirmed(evt.ListingID, evt.BookingID, evt.StartDate, evt.EndDate)
				}(evt)
			} else if evt.Event == "booking.cancelled" {
				go func(e interface{}) {
					_ = processBookingCancelled(evt.ListingID, evt.BookingID)
				}(evt)
			}
		}
	}
}

func availableListingsHandler(c echo.Context) error {
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
	nights := int(checkout.Sub(checkin).Hours() / 24)
	if nights > 30 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "search range cannot exceed 30 nights"})
	}

	var minPrice, maxPrice float64
	if minPriceStr != "" {
		if _, err := fmt.Sscanf(minPriceStr, "%f", &minPrice); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid min_price"})
		}
	}
	if maxPriceStr != "" {
		if _, err := fmt.Sscanf(maxPriceStr, "%f", &maxPrice); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid max_price"})
		}
	}
	if minPriceStr != "" && maxPriceStr != "" && minPrice > maxPrice {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "min_price cannot be greater than max_price"})
	}

	if MongoClient == nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{"error": "MongoClient is nil"})
	}

	listingsColl := MongoClient.Database("raas").Collection("listings")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"location_id": locationID}
	if minPriceStr != "" {
		filter["price_per_day"] = bson.M{"$gte": minPrice}
	}
	if maxPriceStr != "" {
		if _, ok := filter["price_per_day"]; ok {
			filter["price_per_day"] = bson.M{"$gte": minPrice, "$lte": maxPrice}
		} else {
			filter["price_per_day"] = bson.M{"$lte": maxPrice}
		}
	}

	cursor, err := listingsColl.Find(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	blocksColl := MongoClient.Database("raas").Collection("availability_blocks")
	for cursor.Next(ctx) {
		var l Listing
		if err := cursor.Decode(&l); err != nil {
			continue
		}

		// count blocking blocks
		count, _ := blocksColl.CountDocuments(ctx, bson.M{"listing_id": l.ID, "date": bson.M{"$gte": checkin, "$lt": checkout}})
		if count == 0 {
			results = append(results, map[string]interface{}{
				"listing_id":     l.ID.Hex(),
				"display_name":   l.Title,
				"location_label": l.LocationLabel,
				"base_price":     l.PricePerDay,
			})
		}
	}

	return c.JSON(http.StatusOK, results)
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
