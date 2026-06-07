# Listing Service - Availability

This service exposes an availability API and consumes booking events to maintain per-day availability blocks.

Endpoints:

- `GET /v1/listings/available?checkin=YYYY-MM-DD&checkout=YYYY-MM-DD&location_id=...&min_price=&max_price=`: Returns available listings for the date range and location. `checkin` is inclusive, `checkout` is exclusive. Maximum 30 nights.

Event consumption:

- Listens for `booking.confirmed` and `booking.cancelled` events on the `system.events` topic and updates `availability_blocks` and `cancellation_tombstones` collections.
