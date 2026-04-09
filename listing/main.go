package main

import (
	"context"
	"encoding/json"
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
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HostID      string             `json:"host_id" bson:"host_id"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	PricePerDay float64            `json:"price_per_day" bson:"price_per_day"`
	Status      string             `json:"status" bson:"status"`
	MediaURLs   []string           `json:"media_urls" bson:"media_urls,omitempty"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
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

		var event struct {
			Event     string `json:"event"`
			ListingID string `json:"listing_id"`
			URL       string `json:"url"`
		}
		if err := json.Unmarshal(m.Value, &event); err == nil && MongoClient != nil {
			objID, errID := primitive.ObjectIDFromHex(event.ListingID)
			if errID == nil {
				collection := MongoClient.Database("raas").Collection("listings")
				if event.Event == "media.uploaded" && event.URL != "" {
					collection.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$addToSet": bson.M{"media_urls": event.URL}})
				} else if event.Event == "booking.confirmed" {
					collection.UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": bson.M{"status": "BOOKED"}})
				}
			}
		}
	}
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
