package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/segmentio/kafka-go"
)

var (
	S3Client    *s3.Client
	KafkaWriter *kafka.Writer
)

func InitS3() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	S3Client = s3.NewFromConfig(cfg)
	log.Println("Initialized S3/R2 Client")
}

func InitKafkaWriter(brokers []string, topic string) {
	KafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	log.Println("Initialized Kafka Writer")
}
