# Tasks: Core Microservice Implementation

**Feature Branch**: `001-core-microservice`
**Input**: Design documents from `/specs/001-core-microservice/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are omitted as they were not explicitly requested, but independent test criteria defined in user stories guide validation.

**Organization**: Tasks are grouped by user story, and further subdivided by language/developer track to enable independent branching and parallel implementation across the microservices.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

### Go Developer Track (Listing, Media, Booking)

- [ ] T001 Initialize Go 1.25+ module and Echo framework base structure for Listing Service in listing/
- [ ] T002 [P] Initialize Go 1.25+ module and Echo framework base structure for Media Service in media/
- [ ] T003 [P] Initialize Go 1.25+ module and Echo framework base structure for Booking Service in booking/

### Java Developer Track (Payment, Review, Favorites)

- [ ] T004 [P] Initialize Java 21+ Spring Boot base structure for Payment Service in payment/
- [ ] T005 [P] Initialize Java 21+ Spring Boot base structure for Review Service in review/
- [ ] T006 [P] Initialize Java 21+ Spring Boot base structure for Favorites Service in favorites/

### Python Developer Track (Notification, User, Analytics)

- [x] T007 [P] Initialize Python 3.11+ base structure (FastAPI) for Notification Service in notification/
- [x] T008 [P] Initialize Python 3.11+ base structure (FastAPI) for User Service in user/
- [x] T009 [P] Initialize Python 3.11+ base structure (FastAPI) for Analytics Service in analytics/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

### Go Developer Track

- [ ] T010 Implement `/health` endpoint and configure Dockerfile for Listing Service in listing/main.go and listing/Dockerfile
- [ ] T011 [P] Implement `/health` endpoint and configure Dockerfile for Media Service in media/main.go and media/Dockerfile
- [ ] T012 [P] Implement `/health` endpoint and configure Dockerfile for Booking Service in booking/main.go and booking/Dockerfile
- [ ] T013 Configure Kafka client wrappers and database utilities (MongoDB, PostgreSQL, Redis, S3) for Go services

### Java Developer Track

- [ ] T014 [P] Implement `/health` endpoint and configure Dockerfile for Payment Service in payment/src/ and payment/Dockerfile
- [ ] T015 [P] Implement `/health` endpoint and configure Dockerfile for Review Service in review/src/ and review/Dockerfile
- [ ] T016 [P] Implement `/health` endpoint and configure Dockerfile for Favorites Service in favorites/src/ and favorites/Dockerfile
- [ ] T017 Configure Kafka client wrappers and database connection utilities (PostgreSQL, Redis) for Java services

### Python Developer Track

- [x] T018 [P] Implement `/health` endpoint and configure Dockerfile for Notification Service in notification/main.py and notification/Dockerfile
- [x] T019 [P] Implement `/health` endpoint and configure Dockerfile for User Service in user/main.py and user/Dockerfile
- [x] T020 [P] Implement `/health` endpoint and configure Dockerfile for Analytics Service in analytics/main.py and analytics/Dockerfile
- [x] T021 Configure Kafka client wrappers and database connection utilities (PostgreSQL, MongoDB) for Python services

**Checkpoint**: Foundation ready for each specific stack - user story implementation can now begin independently

---

## Phase 3: User Story 1 - Create and Manage Domain Entity (Priority: P1) đźŽŻ MVP

**Goal**: Users must be able to create, update, and retrieve the core domain entity associated with this microservice securely.
**Independent Test**: Can be fully tested by creating a new entity via the REST API and verifying its persistence in the database and corresponding event publication.

### Go Developer Track

- [ ] T022 [P] [US1] Create Listing model (MongoDB) and POST/GET API endpoints in listing/main.go
- [ ] T023 [P] [US1] Create Media model (S3/R2) and POST/GET API endpoints for uploads in media/main.go
- [ ] T024 [P] [US1] Create Booking model (Postgres) and POST/GET API endpoints with Redis Redlock distributed locks in booking/main.go
- [ ] T025 [P] [US1] Emit `booking.created` event from Booking Service on booking placement in booking/main.go
- [ ] T026 [P] [US1] Emit `media.uploaded` event from Media Service on successful upload in media/main.go

### Java Developer Track

- [ ] T027 [P] [US1] Create Review model (Postgres) and POST/GET API endpoints in review/src/
- [ ] T028 [P] [US1] Create Favorites UserFavorites logic (Redis hash) and POST/GET API endpoints in favorites/src/

### Python Developer Track

- [x] T029 [US1] Create User model (Postgres) and POST/GET API endpoints in user/main.py
- [x] T030 [P] [US1] Emit `user.created` event from User Service on successful registration in user/main.py
- [x] T031 [P] [US1] Create REST API endpoints for Analytics event ingestion (MongoDB/Postgres) in analytics/main.py

**Checkpoint**: User Story 1 is fully functional across domains

---

## Phase 4: User Story 2 - Event Consumption and Reactivity (Priority: P2)

**Goal**: The microservice must listen to external system events and update its internal state or trigger subsequent actions idempotently.
**Independent Test**: Can be fully tested by publishing a mock event to the message broker and verifying the microservice correctly processes and reacts to the event without duplicating effort on retry.

### Go Developer Track

- [ ] T032 [P] [US2] Implement idempotent Kafka consumer for `payment.succeeded` & `payment.failed` in Booking Service (Saga complete/compensate) in booking/main.go
- [ ] T033 [US2] Emit `booking.confirmed` / `booking.rejected` from Booking Service upon handling payment results in booking/main.go
- [ ] T034 [P] [US2] Implement Kafka consumer in Listing Service to listen to `booking.confirmed` for availability logic updates, and `media.uploaded` for asset attachments in listing/main.go

### Java Developer Track

- [ ] T035 [US2] Implement idempotent Kafka consumer (Outbox Pattern or event_id tracking) for `booking.created` in Payment Service in payment/src/
- [ ] T036 [US2] Integrate Stripe Java SDK to process payment charges (creating PaymentIntents/handling webhooks) in payment/src/
- [ ] T037 [US2] Implement Event Dispatcher in Payment Service to emit `payment.succeeded` / `payment.failed` after processing with Stripe in payment/src/

### Python Developer Track

- [x] T038 [P] [US2] Implement Kafka consumer in Notification Service to listen for formatting and sending mock notifications on `user.created`, `payment.succeeded`, `payment.failed`, `booking.confirmed`, `booking.rejected` in notification/main.py
- [x] T039 [P] [US2] Implement idempotent Kafka consumer for `booking.created` in Analytics Service in analytics/main.py
- [x] T040 [P] [US2] Implement Kafka consumers in Analytics Service for tracking events (`payment.succeeded`, `payment.failed`, `booking.confirmed`, `booking.rejected`, `media.uploaded`) in analytics/main.py

**Checkpoint**: Saga event-driven lifecycle complete

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Go Developer Track

- [ ] T041 Create Kubernetes `deployment.yaml` and `service.yaml` manifests for Go services

### Java Developer Track

- [ ] T042 Create Kubernetes `deployment.yaml` and `service.yaml` manifests for Java services

### Python Developer Track

- [x] T043 Create Kubernetes `deployment.yaml` and `service.yaml` manifests for Python services

### Cross-Cutting (Any Developer)

- [ ] T044 Standardize pagination, filtering, error handling formats, and OpenTelemetry across all 9 endpoints

---

## Dependencies & Execution Order

### Parallel Opportunities

- Tasks are explicitly divided by Developer Track (Go, Java, Python).
- Each developer can create their own branch (e.g., `001-core-go`, `001-core-java`, `001-core-python`) and implement their Track independently.
- Since services do not share databases and communicate solely via REST/Kafka contracts, they can mock external endpoints or produce local mock events to test.

### Examples

**For Go developers**:
`/speckit.implement I am working on the Go services. Please create and checkout a new branch named "001-core-go", and implement ONLY the tasks listed under the "Go Developer Track" headers in tasks.md. Ignore Java and Python tasks.`

**For Java developers**:
`/speckit.implement I am working on the Java services. Please create and checkout a new branch named "001-core-java", and implement ONLY the tasks listed under the "Java Developer Track" headers in tasks.md. Ignore Go and Python tasks.`

**For Python developers**:
`/speckit.implement I am working on the Python services. Please create and checkout a new branch named "001-core-python", and implement ONLY the tasks listed under the "Python Developer Track" headers in tasks.md. Ignore Go and Java tasks.`
