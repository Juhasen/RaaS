# Local Quickstart Guide

## Prerequisites
- Docker & Docker Compose
- JDK 21+ (for Java: Payment, Review, Favorites)
- Go 1.25+ (for Go: Listing, Media, Booking)
- Python 3.11+ (for Python: Notification, User, Analytics)

## Starting the Infrastructure
Ensure Kafka, Postgres, MongoDB, Redis, and MinIO/Localstack (S3) are running locally via Docker Compose.

`ash
docker-compose up -d
`

## Running the Services
Since there are 9 microservices, you can run them via Docker once implemented. 

For local development without Docker:
- Go services: cd <service_name> && go run main.go
- Java services: cd <service_name> && ./mvnw spring-boot:run
- Python services: cd <service_name> && pip install -r requirements.txt && fastapi run main.py

Health checks are available at http://localhost:<SERVICE_PORT>/health.
