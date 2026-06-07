# Implementation Plan: API Gateway Entry Point & Auth Context

**Branch**: `006-api-gateway-auth` | **Date**: 2026-05-18 | **Spec**: specs/006-api-gateway-auth/spec.md
**Input**: Feature specification from `/specs/006-api-gateway-auth/spec.md`

**Note**: This plan follows the speckit template and references research and design artifacts in this folder.

## Summary

Provide a single API gateway entry point for all UI traffic that routes by path and method, validates JWTs locally, enforces role and ownership policies, and propagates a trusted user identity context to downstream services while leaving responses unmodified.

## Technical Context

**Language/Version**: Gateway configuration with a policy component; recommended Go 1.21+ if custom auth service is required  
**Primary Dependencies**: API gateway runtime (Envoy/NGINX/Kong class), JWT validation library, policy engine/middleware, HTTP client for ownership checks  
**Storage**: N/A (stateless; configuration via files or config maps)  
**Testing**: Unit tests for policy evaluation; integration tests for routing/auth; contract tests for header propagation  
**Target Platform**: Linux containers on Kubernetes (local k8s and production)  
**Project Type**: API gateway/edge proxy service  
**Performance Goals**: Minimal added latency; keep gateway overhead small enough to preserve downstream SLOs  
**Constraints**: JWT validated locally; no per-request user service calls; protected-by-default routes with explicit public allowlist; role + ownership enforcement; 401/403/404 for auth and routing errors; strip/override client identity headers  
**Scale/Scope**: All UI traffic for listing, booking, media, review, favorites, payment, user, notification, analytics; roles limited to Host and Guest in this phase

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- API Gateway as single entry point: PASS (explicit requirement)
- Security (JWT/OAuth2, validate all incoming requests): PASS (local JWT validation, header stripping)
- Service design (independent services, DTOs): PASS (gateway only routes and enforces auth; no shared data)
- API design (consistent status codes, validation): PASS (401/403/404 defined; standardized error body)
- Observability (logging, tracing, health checks): PASS (gateway adds request id and metrics)
- DevOps (containerized, k8s): PASS (gateway deployed as container)

Post-design check: PASS (no constitution violations identified).

## Project Structure

### Documentation (this feature)

```text
specs/006-api-gateway-auth/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── gateway-auth-context.md
└── tasks.md
```

### Source Code (repository root)

```text
gateway/
├── Dockerfile
├── config/
│   ├── routes.yaml
│   └── public-routes.yaml
├── policies/
│   └── ownership.yaml
├── authz/
│   └── src/
└── tests/
    ├── integration/
    └── unit/

k8s/
└── apps/
    └── gateway.yaml
```

**Structure Decision**: Add a new gateway service at repo root, deployed via a dedicated k8s manifest; routing and policy configuration live alongside the gateway, while tests cover routing, auth, and ownership enforcement.

## Architecture Overview

- Single public entry point routes requests by method + path to backend services with no response aggregation.
- JWT validation occurs at the gateway using locally available signing keys or JWKS cache.
- Public endpoints are explicitly allowlisted; everything else is protected by default.
- Identity context is injected by the gateway into downstream requests; client-supplied identity headers are stripped.
- Role-based access is enforced by policy rules; ownership checks call a lightweight internal ownership endpoint on the target service when required.
- Error responses are standardized (401, 403, 404) for auth and routing failures; upstream failures return 502/503.

## Risks & Mitigations

- Misconfigured allowlist could expose protected routes; mitigate with tests that assert all routes are protected unless explicitly public.
- Ownership checks add latency and can fail if a backend is down; mitigate with strict timeouts and clear 502/503 responses.
- Identity header spoofing; mitigate by stripping inbound `x-raas-*` headers and overwriting at the gateway.
- JWT key rotation or configuration drift; mitigate with JWKS refresh and config validation on startup.
- Gateway becomes a single point of failure; mitigate with horizontal scaling and health checks.

## Milestones

1. Define routing map and public allowlist, plus k8s deployment for gateway entry point.
2. Implement JWT validation and identity context propagation, including header stripping.
3. Add role-based policies and ownership verification flow with service contracts.
4. Standardize error responses, observability, and integration tests for routing/auth.

## Complexity Tracking

No constitution violations identified.
