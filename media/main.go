package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Juhasen/RaaS/media/handlers"
	"github.com/Juhasen/RaaS/media/repository"
	"github.com/Juhasen/RaaS/media/service"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. Load configuration from .env file (if exists)
	if err := service.LoadEnv(".env"); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	// 2. Load configuration struct using go-env
	cfg, err := service.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 3. Initialize MongoDB client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()
	log.Println("Connected to MongoDB")

	// 4. Initialize clean architecture layers
	repo := repository.NewMongoMediaRepository(client, cfg.MongoDBName)
	mediaService := service.NewMediaService(repo, cfg, cfg.KafkaBrokers, "system.events")
	defer mediaService.Close()

	handler := handlers.NewMediaHandler(mediaService)

	// 5. Setup Router
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Upload routes (POST /media/upload maps to API Gateway proxy)
	e.POST("/media", handler.UploadMedia)
	e.POST("/media/upload", handler.UploadMedia)

	// Retrieval routes
	e.GET("/media/:id", handler.GetMedia)
	e.GET("/media/listing/:listing_id", handler.GetListingMedia)

	// Fallback static files serving (e.g. locally uploaded images)
	e.Static("/uploads", "./uploads")

	// Start Kafka consumer in background
	go startKafkaConsumers(cfg.KafkaBrokers, mediaService)

	// 6. Start server
	log.Printf("Starting media-service on :%s...", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}

func startKafkaConsumers(brokers []string, mediaService *service.MediaService) {
	log.Printf("Starting Media Kafka consumer on brokers %v (system.events)", brokers)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "media-service-group",
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
		}

		if err := json.Unmarshal(m.Value, &evt); err == nil {
			if evt.Event == "listing.deleted" && evt.ListingID != "" {
				go func(lID string) {
					errProc := mediaService.DeleteMediaByListingID(context.Background(), lID)
					if errProc != nil {
						log.Printf("Failed to process listing.deleted event for listing %s: %v", lID, errProc)
					}
				}(evt.ListingID)
			}
		}
	}
}
