# Contract: Gateway Auth Context and Routing

## Scope
Defines authentication, authorization, identity propagation, and routing expectations between the API gateway, clients, and backend services.

## Client -> Gateway
- Authorization: Bearer <jwt> (required for protected routes)
- Public routes may omit Authorization.
- Client-supplied x-raas-* headers are ignored and stripped.

## Gateway -> Services (Injected Headers)
- x-raas-user-id: string
- x-raas-user-role: host|guest
- x-raas-authenticated: true
- x-raas-request-id: UUID string

## Error Responses
Content-Type: application/json

Status codes:
- 401 auth_required, auth_invalid
- 403 auth_forbidden (role or ownership)
- 404 route_not_found

Body:
{
  "error_code": "...",
  "message": "...",
  "request_id": "..."
}

## Ownership Verification
For routes with ownership enforcement, the gateway performs a preflight request:

- Method: GET (or HEAD)
- Path: /internal/ownership/{resourceType}/{resourceId}
- Headers: x-raas-user-id, x-raas-user-role, x-raas-request-id
- Responses:
  - 200: requester is owner
  - 403: requester is not owner
  - 404: resource not found

Timeouts or upstream errors return 502/503 from the gateway.

## Public Route Allowlist
Public routes are declared in config:
- method
- path_pattern
- backend_service
- public: true

All other routes are protected by default.
