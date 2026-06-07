# Quickstart: API Gateway Entry Point & Auth Context

## Prereqs
- Docker and local Kubernetes
- Downstream services reachable from the gateway
- JWT signing keys or JWKS URL for token validation

## Configure
- GATEWAY_ROUTES_CONFIG: path or URL to routes.yaml
- GATEWAY_PUBLIC_ROUTES: path or URL to public-routes.yaml
- JWT_ISSUER: expected issuer
- JWT_AUDIENCE: expected audience
- JWT_JWKS_URL or JWT_SIGNING_KEY
- OWNERSHIP_TIMEOUT_MS: timeout for ownership checks

## Run (local k8s)
1. Apply the gateway manifest (k8s/apps/gateway.yaml) to the cluster.
2. Port-forward or expose the gateway service.

## Smoke Tests
- Public route without token returns 200.
- Protected route without token returns 401.
- Protected route with guest token to host-only endpoint returns 403.
