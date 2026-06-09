# Feature Specification: Listing Availability Search

**Feature Branch**: `[007-listing-availability]`  
**Created**: 2026-05-18  
**Status**: Draft  
**Input**: User description: "Go-focused listing availability search feature. Scope: listing service maintains availability by consuming booking.confirmed/booking.cancelled events, provides API to query available listings by date range and location, and supports basic price filters. Keep it tech-agnostic and focused on behavior, data, and outcomes. Include roles (Host, Guest) only if needed, and keep tasks within the Go listing service boundary (no cross-language tasks)."

## Clarifications

### Session 2026-05-18

- Q: How should out-of-order cancellations be handled? -> A: Record a cancellation tombstone so later confirmations do not block availability.
- Q: How should availability blocks be stored? -> A: Store availability blocks as individual days.
- Q: What should be used as the location filter? -> A: Use a standardized location_id.
- Q: What is the maximum search range length? -> A: Maximum 30 nights.
- Q: Which HTTP status code should be used for validation errors? -> A: Use HTTP 400 for invalid search inputs.

### Session 2026-06-07

- Q: Should price filters use base nightly price or per-date prices? -> A: Use listing's base nightly price only (per-date pricing out of scope for this version).

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Find Available Listings (Priority: P1)

Guests search for listings that are available for a specific date range and location, optionally narrowing results by price.

**Why this priority**: This is the core value of the feature and enables discovery without booking conflicts.

**Independent Test**: Can be fully tested by submitting a search request with date range, location, and price filters and verifying the returned listings are available and match filters.

**Acceptance Scenarios**:

1. **Given** listings with confirmed bookings, **When** a guest searches by date range and location, **Then** only listings without overlapping confirmed bookings are returned.
2. **Given** listings with varying prices, **When** a guest applies minimum and maximum price filters, **Then** results only include listings within the specified price range.

---

### User Story 2 - Availability Reflects Booking Changes (Priority: P2)

Guests see search results that stay in sync with booking confirmations and cancellations.

**Why this priority**: Accurate availability prevents double-booking and maintains trust in search results.

**Independent Test**: Can be fully tested by sending booking.confirmed and booking.cancelled events for a listing and verifying search results change accordingly.

**Acceptance Scenarios**:

1. **Given** a listing is returned for a date range, **When** a booking.confirmed event overlaps that range, **Then** the listing is no longer returned for that range.
2. **Given** a listing is not returned due to a confirmed booking, **When** a booking.cancelled event for that booking is processed, **Then** the listing becomes eligible to appear for that range.

---

### User Story 3 - Clear Feedback for Invalid Searches (Priority: P3)

Guests receive clear feedback when their search inputs are invalid and can correct them.

**Why this priority**: Prevents confusion and reduces retries by guiding the user to valid searches.

**Independent Test**: Can be fully tested by submitting invalid date ranges or inconsistent price filters and verifying the response explains the issue.

**Acceptance Scenarios**:

1. **Given** a search request where check-in is on or after check-out, **When** the request is submitted, **Then** the system rejects it with a clear validation message.
2. **Given** a search request with a minimum price higher than the maximum price, **When** the request is submitted, **Then** the system rejects it with a clear validation message.

---

[Add more user stories as needed, each with an assigned priority]

### Edge Cases

- Duplicate booking events are received for the same booking and date range.
- Booking events arrive out of order (cancellation before confirmation).
- A cancellation is received for an unknown booking.
- The search date range spans a high-demand period with zero availability.
- The search location does not match any listings.
- Price filters are omitted, partially provided, or inverted.

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: System MUST ingest booking.confirmed and booking.cancelled events and update availability for the affected listing and date range.
- **FR-002**: System MUST treat confirmed bookings as blocking availability for overlapping dates until a matching cancellation is processed, represented as per-day blocks.
- **FR-003**: System MUST allow availability search by date range and a standardized location_id, returning only listings whose stored location matches that identifier.
- **FR-004**: System MUST treat check-in as inclusive and check-out as exclusive when evaluating overlaps with confirmed bookings.
- **FR-005**: System MUST support optional minimum and maximum price filters using each listing's base nightly price.
  Per-date (date-varying) pricing is out of scope for filtering in this version; filters operate on the static base nightly price.
- **FR-006**: System MUST validate that the search date range has a check-in date before the check-out date and return a clear error when invalid.
- **FR-007**: System MUST validate that minimum price is less than or equal to maximum price when both are provided and return a clear error when invalid.
- **FR-008**: System MUST handle repeated booking events for the same booking so availability is updated at most once.
- **FR-009**: System MUST ignore cancellation events that do not correspond to a known confirmed booking and must not change availability based on them.
- **FR-010**: System MUST return listing identifiers, display name, location label, and base nightly price in search results.
- **FR-011**: System MUST record a cancellation tombstone for unknown bookings so a later confirmation does not create a blocked date range.
- **FR-012**: System MUST reject searches longer than 30 nights with a clear validation message.
- **FR-013**: System MUST return HTTP 400 for invalid search inputs.

### Key Entities *(include if feature involves data)*

- **Listing**: A rentable unit with location attributes and base nightly price.
- **Availability Block**: A blocked date range derived from a confirmed booking for a listing.
- **Booking Event**: A confirmed or cancelled booking message containing listing identifier, booking identifier, and date range.
- **Search Criteria**: Date range, location filter, and optional price bounds used to query availability.

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: 95% of availability searches return results in under 2 seconds for date ranges up to 30 nights.
- **SC-002**: 99% of confirmed or cancelled bookings affect search results within 1 minute of event receipt.
- **SC-003**: 0 acceptance test cases return listings with overlapping confirmed bookings for the searched date range.
- **SC-004**: 90% of users can find at least one available listing for a valid search within 3 minutes.

## Assumptions

- Listings already exist in the listing service with a normalized location identifier and base nightly price.
- Date ranges are interpreted as check-in inclusive and check-out exclusive using calendar dates local to the listing.
- Only confirmed bookings affect availability; pending or tentative holds are out of scope.
- Manual availability blocks or overrides are out of scope for this version.
- Booking events include listing identifier, booking identifier, and start/end dates.

## Implementation Notes (progress)

- Initial implementation in `listing/` service added:
  - `availability_blocks` collection storing per-day blocks (`listing_id`, `date`, `booking_id`).
  - `cancellation_tombstones` collection for out-of-order cancellations.
  - Kafka consumer handles `booking.confirmed` and `booking.cancelled` events (expects `start_date`/`end_date` in YYYY-MM-DD).
  - API `GET /v1/listings/available` implemented with validation and price/location filters.

These changes are in branch `007-listing-availability` and include basic fixtures under `specs/007-listing-availability/tests/fixtures.json`.
