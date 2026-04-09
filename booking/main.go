package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"log"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
)

type Booking struct {
	ID         string    `json:"id"`
	ListingID  string    `json:"listing_id"`
	GuestID    string    `json:"guest_id"`
	StartDate  string    `json:"start_date"` // YYYY-MM-DD
	EndDate    string    `json:"end_date"`   // YYYY-MM-DD
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"` // PENDING, CONFIRMED, REJECTED
	CreatedAt  time.Time `json:"created_at"`
}

var rs *redsync.Redsync

func main() {
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/bookings", createBooking)
	e.GET("/bookings/:id", getBooking)

	go startKafkaConsumers()

	e.Logger.Fatal(e.Start(":8080"))
}

func getRedsync() *redsync.Redsync {
	if rs == nil && RedisClient != nil {
		pool := goredis.NewPool(RedisClient)
		rs = redsync.New(pool)
	}
	return rs
}

func startKafkaConsumers() {
	log.Println("Starting Booking Kafka consumers (mocked without actual brokers for MVP)")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "booking-service-group",
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
			BookingID string `json:"booking_id"`
			ListingID string `json:"listing_id"`
		}
		if err := json.Unmarshal(m.Value, &event); err == nil && DB != nil {
			if event.Event == "payment.succeeded" {
				log.Printf("Payment succeeded for booking %s, confirming...", event.BookingID)
				DB.Exec(context.Background(), "UPDATE bookings SET status='CONFIRMED' WHERE id=$1", event.BookingID)
				emitSagaEvent("booking.confirmed", event.BookingID, event.ListingID)
			} else if event.Event == "payment.failed" {
				log.Printf("Payment failed for booking %s, rejecting...", event.BookingID)
				DB.Exec(context.Background(), "UPDATE bookings SET status='REJECTED' WHERE id=$1", event.BookingID)
				emitSagaEvent("booking.rejected", event.BookingID, event.ListingID)
			}
		}
	}
}

func emitSagaEvent(eventName, bookingID, listingID string) {
	if KafkaWriter != nil {
		eventMsg := fmt.Sprintf(`{"event":"%s","booking_id":"%s","listing_id":"%s"}`, eventName, bookingID, listingID)
		KafkaWriter.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(bookingID),
				Value: []byte(eventMsg),
			},
		)
	}
}

func createBooking(c echo.Context) error {
	var req Booking
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid structure"})
	}

	if req.ListingID == "" || req.GuestID == "" || req.StartDate == "" || req.EndDate == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing required fields"})
	}

	syncMap := getRedsync()
	if syncMap != nil {
		mutexname := fmt.Sprintf("listing-lock-%s", req.ListingID)
		mutex := syncMap.NewMutex(mutexname, redsync.WithExpiry(8*time.Second))

		if err := mutex.Lock(); err != nil {
			return c.JSON(http.StatusConflict, echo.Map{"error": "Could not acquire lock for listing, please try again"})
		}
		defer mutex.Unlock()
	} else {
		c.Logger().Warn("Redis not initialized, skipping distributed lock")
	}

	req.ID = uuid.NewString()
	req.Status = "PENDING"
	req.CreatedAt = time.Now()

	if DB != nil {
		var conflictCount int
		err := DB.QueryRow(context.Background(),
			"SELECT count(1) FROM bookings WHERE listing_id=$1 AND start_date < $2 AND end_date > $3 AND status != $4",
			req.ListingID, req.EndDate, req.StartDate, "REJECTED",
		).Scan(&conflictCount)
		if err == nil && conflictCount > 0 {
			return c.JSON(http.StatusConflict, echo.Map{"error": "booking dates overlap with an existing booking"})
		}

		_, err = DB.Exec(context.Background(),
			`INSERT INTO bookings (id, listing_id, guest_id, start_date, end_date, total_price, status, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			req.ID, req.ListingID, req.GuestID, req.StartDate, req.EndDate, req.TotalPrice, req.Status, req.CreatedAt,
		)
		if err != nil {
			c.Logger().Errorf("DB Insert failed: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to save booking"})
		}
	} else {
		c.Logger().Warn("Postgres not initialized, skipping insert")
	}

	if KafkaWriter != nil {
		eventMsg := fmt.Sprintf(`{"event":"booking.created","booking_id":"%s","listing_id":"%s","guest_id":"%s","total_price":%f}`, req.ID, req.ListingID, req.GuestID, req.TotalPrice)
		err := KafkaWriter.WriteMessages(context.Background(),
			kafka.Message{
				Key:   []byte(req.ID),
				Value: []byte(eventMsg),
			},
		)
		if err != nil {
			c.Logger().Errorf("Failed to emit booking.created event: %v", err)
		} else {
			c.Logger().Info("Emitted booking.created event")
		}
	} else {
		c.Logger().Warn("KafkaWriter not initialized, skipping event emission")
	}

	return c.JSON(http.StatusCreated, req)
}

func getBooking(c echo.Context) error {
	id := c.Param("id")

	if DB == nil {
		return c.JSON(http.StatusServiceUnavailable, echo.Map{"error": "Database not initialized"})
	}

	var b Booking
	err := DB.QueryRow(context.Background(),
		"SELECT id, listing_id, guest_id, start_date, end_date, total_price, status, created_at FROM bookings WHERE id=$1",
		id,
	).Scan(&b.ID, &b.ListingID, &b.GuestID, &b.StartDate, &b.EndDate, &b.TotalPrice, &b.Status, &b.CreatedAt)

	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "booking not found"})
	}

	return c.JSON(http.StatusOK, b)
}
