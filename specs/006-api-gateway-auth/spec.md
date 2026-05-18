# Feature Specification: API Gateway Entry Point & Auth Context

**Feature Branch**: `006-api-gateway-auth`  
**Created**: 2026-05-18  
**Status**: Draft  
**Input**: User description: "API gateway entrypoint with auth verification and user context propagation"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Single Entry Point for the UI (Priority: P1)

As a user, I access the application through one stable base URL, and the system routes each request to the correct backend capability without exposing internal service boundaries.

**Why this priority**: The UI must not call multiple services directly, and this is a prerequisite for safe production access control.

**Independent Test**: Can be fully tested by sending multiple public and protected requests through the single entry point and confirming the correct responses are returned.

**Acceptance Scenarios**:

1. **Given** the UI calls a listing search endpoint through the entry point, **When** the request is sent, **Then** the system returns listing results from the correct backend service.
2. **Given** the UI calls a booking endpoint through the entry point, **When** the request is sent, **Then** the system returns the booking response from the correct backend service.

---

### User Story 2 - Protected Access Uses Valid Identity (Priority: P1)

As a signed-in user, I can access protected actions only when my access token is valid, and the system carries my identity to downstream services.

**Why this priority**: Preventing unauthorized access is critical for production readiness.

**Independent Test**: Can be fully tested by attempting the same protected request with a valid token and with an invalid or missing token.

**Acceptance Scenarios**:

1. **Given** a valid access token, **When** I request a protected endpoint, **Then** the request succeeds and includes my identity context for the backend to use.
2. **Given** a missing or invalid access token, **When** I request a protected endpoint, **Then** the system denies access with a clear error response.

---

### User Story 3 - Role-Appropriate Access (Priority: P2)

As a host or guest, I can only access actions appropriate to my role, and role-restricted actions are blocked otherwise.

**Why this priority**: Role separation is required to prevent unauthorized management actions.

**Independent Test**: Can be fully tested by attempting a host-only action using a guest role.

**Acceptance Scenarios**:

1. **Given** a guest role token, **When** I attempt a host-only action, **Then** the system denies access with a role-based error response.

---

### Edge Cases

- Missing, expired, or malformed access token for a protected endpoint
- Client attempts to spoof identity context in request metadata
- Request targets an unmapped route
- Backend service is unavailable or times out

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST provide a single public entry point for all UI requests.
- **FR-002**: The system MUST route requests to the correct backend capability based on path and method.
- **FR-003**: The system MUST clearly distinguish public endpoints from protected endpoints.
- **FR-004**: The system MUST validate access tokens for protected endpoints and reject invalid or expired tokens.
- **FR-005**: The system MUST attach a standardized user identity context (user identifier and role) to backend requests after validation.
- **FR-006**: The system MUST prevent clients from overriding or injecting identity context.
- **FR-007**: The system MUST return consistent error responses for authentication failures and route not found conditions.
- **FR-008**: The system MUST enforce role-based access for endpoints that are restricted by role alone.

Acceptance coverage: The acceptance scenarios in "User Scenarios & Testing" collectively validate FR-001 through FR-008.

### Key Entities *(include if feature involves data)*

- **Access Token**: A signed credential representing a user session and role.
- **User Identity Context**: The user identifier and role passed to backend services for authorization decisions.
- **Route**: A public endpoint mapping to a backend capability.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of UI requests in production are routed through the single entry point.
- **SC-002**: 100% of protected endpoints reject requests without a valid access token.
- **SC-003**: 95% of valid protected requests complete successfully end to end.
- **SC-004**: 100% of role-restricted actions are denied when attempted with the wrong role.

## Assumptions

- The UI will only call the public entry point and will not access internal services directly.
- A separate identity feature provides access tokens to signed-in users.
- The initial role set is limited to Host and Guest.
- Ownership checks for specific resources are handled by downstream services in a separate feature.
