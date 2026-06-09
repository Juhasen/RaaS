# Tasks: 007 - Listing Availability

This file breaks the implementation work into ordered, testable tasks and PR-sized units.

## Phases

### Setup

- T1.1: Create feature tasks (this file) and link to spec. [P]
- T1.2: Add checklist skeleton in `checklists/requirements.md`.

### Tests

- T2.1: Add acceptance test fixtures (events + listings) under `specs/007-listing-availability/tests/`.
- T2.2: Add integration test that verifies search result before/after booking.confirmed and booking.cancelled.

### Core Implementation

- T3.1: Data model: add `AvailabilityBlock` per-day representation in the listing service storage layer.
- T3.2: Event consumer: implement consumer for `booking.confirmed` and `booking.cancelled` events; ensure idempotency and tombstone handling.
- T3.3: Query service: implement availability query method that checks per-day blocks for overlap (check-in inclusive, check-out exclusive) and applies `location_id` and `min_price`/`max_price` filters.
- T3.4: API: add HTTP endpoint `GET /v1/listings/available` with params `checkin, checkout, location_id, min_price?, max_price?` and validation (400 on bad input).

*Progress*: T3.1, T3.2 and T3.4 implemented in working branch `007-listing-availability` as initial PRs (model, consumer wiring, API). T3.3 query service implemented in `main.go` handler; further optimization (aggregation) left as future work.

### Integration

- T4.1: Wire event consumer into service startup and config (kafka topic/subscription or mock transport if local).
- T4.2: Add metrics & simple caching (optional): count searches, latency, and event processing lag.

### Polish / Docs

- T5.1: Add README section with local test instructions and example requests/responses.
- T5.2: Create PR template notes and changelog entry.

## PR suggestions

- PR-001: `feat/007-availability-model` — implement T3.1 and schema changes.
- PR-002: `feat/007-event-consumer` — implement T3.2 with unit tests for idempotency and tombstones.
- PR-003: `feat/007-availability-api` — implement T3.3 + T3.4 and acceptance tests.
- PR-004: `chore/007-tests-and-docs` — add acceptance fixtures, README updates, and checklists.

## Acceptance criteria for main tasks

- PR-001 done when model compiled, migrations (if any) included, and unit tests for block storage pass.
- PR-002 done when consumer processes confirmed/cancelled events, idempotent, and tombstone semantics verified by tests.
- PR-003 done when `GET /v1/listings/available` returns correct results for acceptance fixtures and validates inputs (400 returned for invalid ranges/prices).
