# RaaS Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-05-18

## Active Technologies
- Go 1.21+ (Listing, Media, Booking), Java 17+ (Payment, Review, Favorites), Python 3.11+ (Notification, User, Analytics) + Go web framework (Gin/Mux), Spring Boot, FastAPI/Flask, Kafka Client (001-core-microservice)
- PostgreSQL, MongoDB, Redis, S3 (R2) (001-core-microservice)
- Go 1.25+ (Listing, Media, Booking), Java 21+ (Payment, Review, Favorites), Python 3.11+ (Notification, User, Analytics) + Go web framework (Gin/Mux), Spring Boot, FastAPI/Flask, Kafka Client (001-core-microservice)
- Go 1.25+ (Listing, Media, Booking), Java 21+ (Payment, Review, Favorites), Python 3.11+ (Notification, User, Analytics) + Go web framework (Echo), Spring Boot, FastAPI/Flask, Kafka Client, Stripe Java SDK (001-core-microservice)
- Go 1.25+ (Listing, Media, Booking), Java 21+ (Payment, Review, Favorites), Python 3.11+ (Notification, User, Analytics) + Go web framework (Echo), Spring Boot, FastAPI, Kafka Client, Stripe Java SDK (001-core-microservice)
- Gateway configuration with a policy component; recommended Go 1.21+ if custom auth service is required + API gateway runtime (Envoy/NGINX/Kong class), JWT validation library, policy engine/middleware, HTTP client for ownership checks (006-api-gateway-auth)
- N/A (stateless; configuration via files or config maps) (006-api-gateway-auth)

- Go 1.21+ (Listing, Booking, Notification), Java 17+ (User, Payment, Review, Favorites) + Go web framework (Gin/Mux), Spring Boot, Kafka Client (001-core-microservice)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.21+ (Listing, Booking, Notification), Java 17+ (User, Payment, Review, Favorites)

## Code Style

Go 1.21+ (Listing, Booking, Notification), Java 17+ (User, Payment, Review, Favorites): Follow standard conventions

## Recent Changes
- 006-api-gateway-auth: Added Gateway configuration with a policy component; recommended Go 1.21+ if custom auth service is required + API gateway runtime (Envoy/NGINX/Kong class), JWT validation library, policy engine/middleware, HTTP client for ownership checks
- 001-core-microservice: Added Go 1.25+ (Listing, Media, Booking), Java 21+ (Payment, Review, Favorites), Python 3.11+ (Notification, User, Analytics) + Go web framework (Echo), Spring Boot, FastAPI, Kafka Client, Stripe Java SDK
- 001-core-microservice: Added Go 1.25+ (Listing, Media, Booking), Java 21+ (Payment, Review, Favorites), Python 3.11+ (Notification, User, Analytics) + Go web framework (Echo), Spring Boot, FastAPI/Flask, Kafka Client, Stripe Java SDK


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
