package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"booking/handlers"
	"booking/repository"
	"booking/service"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func main() {
	// 1. Get configurations from environment variables
	pgURI := os.Getenv("PG_URI")
	if pgURI == "" {
		pgURI = "postgres://raas_user:raas_password@localhost:5432/raas_db"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	kafkaBrokersStr := os.Getenv("KAFKA_BROKERS")
	var kafkaBrokers []string
	if kafkaBrokersStr != "" {
		kafkaBrokers = strings.Split(kafkaBrokersStr, ",")
	} else {
		kafkaBrokers = []string{"localhost:9092"}
	}

	// 2. Initialize PostgreSQL connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool, err := pgxpool.New(ctx, pgURI)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL database: %v", err)
	}
	defer pgPool.Close()
	log.Println("Connected to PostgreSQL")

	// 3. Initialize Redis client & redsync distributed locker
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	defer redisClient.Close()
	log.Println("Initialized Redis Client")

	redisPool := goredis.NewPool(redisClient)
	rs := redsync.New(redisPool)

	// 4. Initialize Kafka writer for saga events (system.events)
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers...),
		Topic:    "system.events",
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()
	log.Println("Initialized Kafka Writer")

	// 5. Wire layers
	repo := repository.NewPostgresRepository(pgPool)
	bookingService := service.NewBookingService(repo, rs, kafkaWriter)
	handler := handlers.NewBookingHandler(bookingService)

	// 6. Set up Echo router
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Booking CRUD API endpoints
	e.POST("/bookings", handler.CreateBooking)
	e.GET("/bookings/:id", handler.GetBooking)
	e.PUT("/bookings/:id", handler.UpdateBooking)
	e.DELETE("/bookings/:id", handler.DeleteBooking)
	e.GET("/bookings", handler.ListBookings)

	// 7. Start Kafka payment event consumers
	go startKafkaConsumers(kafkaBrokers, bookingService)

	// 8. Start server
	log.Println("Starting Echo server on :8080...")
	e.Logger.Fatal(e.Start(":8080"))
}

func startKafkaConsumers(brokers []string, bookingService *service.BookingService) {
	log.Printf("Starting Booking Kafka consumers on brokers %v (system.events)", brokers)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "booking-service-group",
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

		var event struct {
			Event     string `json:"event"`
			BookingID string `json:"booking_id"`
			ListingID string `json:"listing_id"`
		}

		if err := json.Unmarshal(m.Value, &event); err == nil {
			if event.Event == "payment.succeeded" || event.Event == "payment.failed" {
				log.Printf("Received payment event: %s for booking %s", event.Event, event.BookingID)
				errProc := bookingService.ProcessPaymentEvent(context.Background(), event.Event, event.BookingID, event.ListingID)
				if errProc != nil {
					log.Printf("Failed to process payment event %s for booking %s: %v", event.Event, event.BookingID, errProc)
				}
			}
		}
	}
}
