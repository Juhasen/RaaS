<!--
Sync Impact Report:
- Version change: Initial → 1.0.0
- List of modified principles:
  - Added 1. Core Principles
  - Added 2. Architecture
  - Added 3. Service Design
  - Added 4. Event-Driven Design
  - Added 5. Data Management
  - Added 6. Concurrency & Consistency
  - Added 7. Security
  - Added 8. Performance
  - Added 9. Observability
  - Added 10. DevOps & Deployment
  - Added 11. Code Standards
  - Added 12. API Design
  - Added 13. Testing
  - Added 14. UI / Frontend
  - Added 15. AI Usage
  - Added 16. Development Workflow
- Templates requiring updates (✅ updated / ⚠ pending):
  - .specify/templates/plan-template.md: ✅
  - .specify/templates/spec-template.md: ✅
  - .specify/templates/tasks-template.md: ✅
- Follow-up TODOs: None
-->

# RaaS Project Constitution

## 1. Core Principles

System MUST be scalable, fault-tolerant, and loosely coupled. Each microservice MUST own its domain and data (no shared databases).
Prefer asynchronous communication over synchronous where possible. Design for failure: every service MUST handle partial outages gracefully.

## 2. Architecture

Follow microservices architecture with clear bounded contexts.
Use API Gateway as the single entry point.
Communication MUST be:

- Sync: REST / gRPC (only when necessary)
- Async: Event Bus (Kafka) as default
  Implement Saga pattern for distributed transactions to avoid tight coupling between services.

## 3. Service Design

Each service MUST have a single responsibility, be independently deployable, and have its own database.
Use DTOs for communication (never expose internal models).
Version all public APIs.

## 4. Event-Driven Design

Events MUST be immutable. Use clear naming convention: `<domain>.<action>` (e.g., booking.created).
Ensure idempotency of event consumers. Handle retries and dead-letter queues (DLQ). Avoid event chaining loops.

## 5. Data Management

Polyglot persistence MUST be enforced:

- PostgreSQL for transactional data
- MongoDB for flexible listing data
- Redis for caching & locks
  Never access another service’s database directly. Use eventual consistency where needed.

## 6. Concurrency & Consistency

Use optimistic locking where possible.
Use Redis for distributed locks (critical sections).
Design the Booking Service to strictly prevent double-booking. Accept eventual consistency in non-critical flows only.

## 7. Security

Use JWT / OAuth2 for authentication. Validate all incoming requests.
Never expose secrets (use environment variables / vaults). Enforce HTTPS everywhere.

## 8. Performance

Cache frequently accessed data via Redis. Avoid blocking operations in critical paths.
Use pagination and filtering for all list endpoints. Monitor latency between services.

## 9. Observability

Implement centralized logging and use tracing (OpenTelemetry).
Monitor request latency, error rates, and event lag (Kafka). Each service MUST expose health checks.

## 10. DevOps & Deployment

All services MUST be containerized (Docker).
Use local Kubernetes (e.g., Docker Desktop with Kubernetes) for local development to mirror production, and Kubernetes for production. CI/CD pipelines are required for each service.

## 11. Code Standards

Go services MUST use idiomatic Go and keep packages small and focused.
Java services MUST follow Spring Boot best practices.
Keep code simple and readable. Avoid over-engineering.

## 12. API Design

REST endpoints MUST be predictable and consistent.
Use proper HTTP status codes and validate input at the API boundary. Document APIs using OpenAPI/Swagger.

## 13. Testing

Write unit tests for business logic. Write integration tests for service communication.
Test event flows (critical in Saga). Mock external services appropriately.

## 14. UI / Frontend

Frontend MUST communicate only via the API Gateway.
Keep UI simple and responsive. Handle async states (loading, retries, failures) gracefully.

## 15. AI Usage

AI-generated code MUST be reviewed before use.
Do not trust AI blindly in critical logic (e.g., payments, booking). Validate all generated logic manually.

## 16. Development Workflow

Use feature branches. Keep commits small and meaningful.
Each PR MUST be reviewable and focused. Prefer working features over premature optimization.

## Governance

Amendments to this constitution require documentation and approval. All PRs and reviews MUST verify compliance with these principles.

**Version**: 1.0.0 | **Ratified**: 2026-04-08 | **Last Amended**: 2026-04-08
