# Research & Technical Decisions: Core Microservice Implementation

## 1. Needs Clarification Solutions

**Topic**: Kafka Idempotency & Consumer Offset Management
**Decision**: We will implement the Outbox Pattern or store a processed event_id in the microservice's database alongside the transaction to ensure exactly-once processing semantic for the Saga steps.
**Rationale**: Saga relies heavily on avoiding duplicate event processing to prevent incorrect payment triggers or booking confirmations.
**Alternatives considered**: Pure Kafka transaction offsets (requires exact tech stack alignment and might be complex with polyglot Spring Boot & Go combination).

## 2. Technology Best Practices

**Topic**: Polyglot Architecture & API Gateway
**Decision**: Use an established API Gateway (like Kong or NGINX/Envoy) to route requests to respective microservices based on paths (e.g., /api/users/ -> User Service).
**Rationale**: Constitution mandates a single entry point.
**Alternatives considered**: Dedicated custom-built edge service in Go (rejected due to excessive maintenance overhead).

**Topic**: Distributed Locks (Booking Service)
**Decision**: Use Redis with Redlock algorithm to lock listing IDs during the booking process.
**Rationale**: High-performance, concurrent-safe lock to prevent double bookings across multiple Go container replicas.
**Alternatives considered**: PostgreSQL database locking (slower and might cause contention under high load).

**Topic**: Saga Pattern Orchestration vs. Choreography
**Decision**: Choreography (event-driven).
**Rationale**: Services should remain loosely coupled without a centralized orchestrator, relying on Kafka events (ooking.placed, payment.succeeded).
**Alternatives considered**: Centralized Orchestrator (adds a potential single point of failure and tight coupling).
