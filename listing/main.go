package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"listing/handlers"
	"listing/repository"
	"listing/service"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. Get configuration from environment variables
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	var kafkaBrokers []string
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
	} else {
		kafkaBrokers = []string{"localhost:9092"}
	}

	// 2. Initialize MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()
	log.Println("Connected to MongoDB")

	// 3. Initialize repository, services, and handlers
	repo := repository.NewMongoRepository(client, "raas")

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers...),
		Topic:    "system.events",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	listingService := service.NewListingService(repo, kafkaWriter)
	availService := service.NewAvailabilityService(repo)

	availHandler := handlers.NewAvailabilityHandler(availService)
	listHandler := handlers.NewListingHandler(listingService, availHandler)

	// 4. Set up router
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Listing CRUD API endpoints
	e.POST("/listings", listHandler.CreateListing)
	e.GET("/listings/:id", listHandler.GetListing)
	e.PUT("/listings/:id", listHandler.UpdateListing)
	e.DELETE("/listings/:id", listHandler.DeleteListing)
	e.GET("/listings", listHandler.ListListings) // Handles list & availability check delegation

	// Availability query API endpoints
	e.GET("/v1/listings/available", availHandler.GetAvailableListings)
	e.GET("/listings/available", availHandler.GetAvailableListings)

	// 5. Start Kafka consumer in background
	go startKafkaConsumers(kafkaBrokers, availService, repo)

	// 6. Start Server
	log.Println("Starting Echo server on :8080...")
	e.Logger.Fatal(e.Start(":8080"))
}

func startKafkaConsumers(brokers []string, availService *service.AvailabilityService, repo *repository.MongoRepository) {
	log.Printf("Starting Listing Kafka consumers on brokers %v (system.events)", brokers)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "listing-service-group",
		Topic:   "system.events",
		MaxWait: 1 * time.Second,
	})
	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			// Backoff on consumer read error
			time.Sleep(5 * time.Second)
			continue
		}

		var evt struct {
			Event     string `json:"event"`
			ListingID string `json:"listing_id"`
			BookingID string `json:"booking_id"`
			URL       string `json:"url"`
			StartDate string `json:"start_date"`
			EndDate   string `json:"end_date"`
		}

		if err := json.Unmarshal(m.Value, &evt); err == nil {
			// Handle media.uploaded separately
			if evt.Event == "media.uploaded" && evt.ListingID != "" {
				objID, errID := primitive.ObjectIDFromHex(evt.ListingID)
				if errID == nil {
					collection := repo.GetListingsCollection()
					mediaURL := evt.URL
					if mediaURL == "" {
						mediaURL = evt.BookingID // Fallback to booking_id field if url is empty
					}
					if mediaURL != "" {
						_, errUpdate := collection.UpdateOne(
							context.Background(),
							bson.M{"_id": objID},
							bson.M{"$addToSet": bson.M{"media_urls": mediaURL}},
						)
						if errUpdate != nil {
							log.Printf("Failed to update media_urls for listing %s: %v", evt.ListingID, errUpdate)
						}
					}
				}
			}

			// Booking availability blocks events
			if evt.Event == "booking.confirmed" {
				go func(lID, bID, start, end string) {
					errProc := availService.ProcessBookingConfirmed(context.Background(), lID, bID, start, end)
					if errProc != nil {
						log.Printf("Failed to process booking.confirmed event for booking %s: %v", bID, errProc)
					}
				}(evt.ListingID, evt.BookingID, evt.StartDate, evt.EndDate)
			} else if evt.Event == "booking.cancelled" {
				go func(lID, bID string) {
					errProc := availService.ProcessBookingCancelled(context.Background(), lID, bID)
					if errProc != nil {
						log.Printf("Failed to process booking.cancelled event for booking %s: %v", bID, errProc)
					}
				}(evt.ListingID, evt.BookingID)
			}
		}
	}
}
