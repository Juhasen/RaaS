# Research: API Gateway Entry Point & Auth Context

## 1. Gateway Routing Approach
Decision: Use a configurable API gateway runtime with a custom auth/authorization component rather than building a bespoke gateway service.
Rationale: Aligns with the architecture requirement for a single entry point while leveraging proven routing features and reducing maintenance.
Alternatives considered: Custom edge service in Go/Java (rejected due to higher maintenance and duplication of gateway features).

## 2. Token Validation
Decision: Validate JWTs locally at the gateway using configured signing keys or cached JWKS; no per-request calls to the user service.
Rationale: Meets FR-010 and keeps latency and availability independent from the user service.
Alternatives considered: Token introspection via user service (rejected due to latency and availability coupling).

## 3. Identity Context Propagation
Decision: Gateway injects `x-raas-user-id` and `x-raas-user-role` headers and strips any client-supplied `x-raas-*` headers.
Rationale: Prevents spoofing and provides a consistent, trusted identity context to downstream services.
Alternatives considered: Forward raw JWT to downstream services (rejected due to duplicated validation and inconsistent policy enforcement).

## 4. Role and Ownership Enforcement
Decision: Evaluate role policies at the gateway; for ownership checks, call a lightweight internal ownership endpoint on the target service with gateway-injected identity headers.
Rationale: Centralizes authorization while using the owning service as the source of truth for resource ownership.
Alternatives considered: Embed ownership data in JWTs (rejected as stale/impractical), or leave ownership checks to services only (rejected due to FR-011).

## 5. Public Route Allowlist
Decision: Default all routes to protected; declare public endpoints in an explicit allowlist config.
Rationale: Meets FR-013 and reduces accidental exposure.
Alternatives considered: Blocklist of protected routes (rejected as risky and hard to maintain).

## 6. Error Response Contract
Decision: Standardize auth and routing error responses with JSON `error_code`, `message`, and `request_id`, using 401/403/404.
Rationale: Consistent client handling and meets FR-007/FR-012.
Alternatives considered: Passthrough upstream errors or plain text responses (rejected for inconsistency).
