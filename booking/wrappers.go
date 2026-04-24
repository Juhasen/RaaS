package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	DB          *pgxpool.Pool
	RedisClient *redis.Client
	KafkaWriter *kafka.Writer
)

func InitPostgres(connString string) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	DB = pool
	log.Println("Connected to PostgreSQL")
}

func InitRedis(addr, password string, db int) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	log.Println("Initialized Redis Client")
}

func InitKafkaWriter(brokers []string, topic string) {
	KafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	log.Println("Initialized Kafka Writer")
}
