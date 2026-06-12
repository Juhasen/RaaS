package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Juhasen/RaaS/media/handlers"
	"github.com/Juhasen/RaaS/media/repository"
	"github.com/Juhasen/RaaS/media/service"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. Load configuration from .env file (if exists)
	if err := service.LoadEnv(".env"); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	// 2. Read environment variables
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDBName := os.Getenv("MONGO_DB_NAME")
	if mongoDBName == "" {
		mongoDBName = "media_db"
	}

	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	var kafkaBrokers []string
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
	} else {
		kafkaBrokers = []string{"localhost:9092"}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 3. Initialize MongoDB client
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

	// 4. Initialize clean architecture layers
	repo := repository.NewMongoMediaRepository(client, mongoDBName)
	mediaService := service.NewMediaService(repo, kafkaBrokers, "system.events")
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

	// 6. Start server
	log.Printf("Starting media-service on :%s...", port)
	e.Logger.Fatal(e.Start(":" + port))
}
