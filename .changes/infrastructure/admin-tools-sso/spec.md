# Admin tools SSO through the gateway

Issue: #52 (follow-up to #50 / PR #51). Spans `gateway`, `user`, `ui` and `k8s/infra/admin-tools/`; filed under infrastructure.

## Problem Statement

Admin infra tools (MinIO, Grafana, Adminer, Redpanda Console) sit on their own NodePorts, so the browser treats each as another origin: the RaaS token (in `localStorage`, sent only by the app's own requests) never reaches them and each asks for its own password. Nothing server-side limits who can reach them, and anyone can self-register as `admin`. The JWT secret also has a hard-coded fallback (`raas-secret-key-12345`, not set in any manifest), so admin tokens can be forged.

## Solution

Serve the tools through the gateway on the UI's origin at `/admin/{grafana,minio,adminer,kafka}/`. Every `/admin/*` request passes an `auth_request` that accepts only `role == admin`. Login additionally sets an HttpOnly cookie so plain link clicks authenticate. Grafana trusts a gateway-set user header (auth-proxy mode). Tool NodePorts and the Redpanda basic-auth proxy are removed. `admin` can no longer be self-registered; one admin is seeded from a generated Secret.

## User Stories

- As an admin, I want each tool button to open the tool without a RaaS re-login, so that I use one session.
- As an admin, I want Grafana to sign me in automatically, so that I never type a Grafana password.
- As a guest, host or logged-out visitor, I want nothing under `/admin/*` to be reachable (401 logged out, 403 other roles), so that infra tools are not exposed.
- As an operator, I want `admin` unobtainable through registration, so that the gateway check protects something.
- As an operator, I want tokens signed with a generated secret, so that admin tokens cannot be forged.
- As an operator, I want the tools off their NodePorts, so that the gateway is the only way in.

## Implementation Decisions

### user (`user/`)

- Login (`POST /users/login`) also sets cookie `raas_token`: HttpOnly, SameSite=Lax, Path=/, Max-Age = token lifetime, `Secure` only when env `COOKIE_SECURE=true` (default off; TLS is out of scope). Body unchanged, so the UI keeps working.
- `POST /users/logout` clears the cookie (204).
- `get_current_user` takes the token from the Bearer header, else the cookie. Header wins.
- `GET /users/admin-check`: 200 for `role == admin` with response header `X-Auth-User: <email>`, 403 for other roles, 401 for no or invalid token. Reads the role from the DB row, not the token claim, so a demoted user loses access immediately.
- Register/create: `UserCreateRequest.role` pattern becomes `^(guest|host)$`; `role=admin` gives 422.
- Startup seed: if `ADMIN_EMAIL` and `ADMIN_PASSWORD` are set and no user has that email, create it with role `admin` (bcrypt, no Kafka event). Idempotent; never changes an existing user. Existing `admin` rows from earlier self-registration stay (see Further Notes).
- `JWT_SECRET` has no fallback: the service fails to start without it. Provided by Secret `user-secrets` (`JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`), added to `user.yaml` via `secretKeyRef`. If other services verify JWTs (`grep -rn JWT_SECRET` at implementation), they read the same Secret.

### gateway (`gateway/default.conf` and `k8s/apps/gateway/gateway-configmap.yaml`, kept identical)

- Internal location `/api/auth/verify-admin` rewrites to `/users/admin-check`, forwards `Cookie` and `Authorization`, no body.
- `location /admin/<tool>/` blocks, each with `auth_request /api/auth/verify-admin;`, plain 401/403 (no redirect), WebSocket upgrade headers, `proxy_read_timeout 300s`:
  - `/admin/grafana/` to the Grafana Service. `auth_request_set $auth_user $upstream_http_x_auth_user;` then `proxy_set_header X-WEBAUTH-USER $auth_user;`, which overwrites any client-sent value.
  - `/admin/minio/` to the MinIO console (9001), `/admin/adminer/` to `adminer:8080`, `/admin/kafka/` to `redpanda-console:8080`.
- Base paths per tool:
  - Grafana: `GF_SERVER_ROOT_URL=%(protocol)s://%(domain)s/admin/grafana/`, `GF_SERVER_SERVE_FROM_SUB_PATH=true`, `GF_AUTH_PROXY_ENABLED=true`, `GF_AUTH_PROXY_HEADER_NAME=X-WEBAUTH-USER`, `GF_AUTH_PROXY_AUTO_SIGN_UP=true`, `GF_USERS_AUTO_ASSIGN_ORG_ROLE=Admin`.
  - MinIO console: gateway strips the prefix, `MINIO_BROWSER_REDIRECT_URL` points at `/admin/minio`.
  - Redpanda Console: `SERVER_BASEPATH=/admin/kafka` (exact keys for the pinned v3.3.0 verified at implementation).
  - Adminer: no base-path setting; the gateway strips the prefix and relative links are checked in the deployment check.
  - If a tool cannot work under a sub-path, ask the user then: drop that tool from the gateway rather than ship a broken button.
- Grafana's own login stays enabled; its Service becomes ClusterIP only, so nothing but the gateway can send the auth header.
- MinIO and Adminer keep their own login (user decision): the gateway gate keeps non-admins out but a second password remains. Redpanda Console has no login, so the gate is its only protection.
- The Traefik ingress already forwards `/` to the gateway; unchanged.

### cluster (`k8s/infra/`)

- Delete `admin-tools/nodeports.yaml` and the `redpanda-console-auth` Deployment, ConfigMap and Secret handling (manifest and `scripts/admin-secrets.ps1`). The `adminer` Service becomes ClusterIP. Grafana and MinIO console get ClusterIP Services if the existing ones lack the port; no NodePort remains for the four tools.
- `scripts/admin-secrets.ps1` also generates `JWT_SECRET` (random), `ADMIN_EMAIL` (`admin@raas.local`) and `ADMIN_PASSWORD`, writes Secret `user-secrets`, restarts `user-deployment`, and lists the admin login in `.passwords`. Rotating `JWT_SECRET` logs everyone out; accepted.
- Memory: no new pods; `redpanda-console-auth` (16Mi request, 32Mi limit) is removed. Before/after `kubectl top` recorded.

### ui (`ui/`)

- `config/infra-tools.ts`: urls become `/admin/grafana/`, `/admin/minio/`, `/admin/adminer/`, `/admin/kafka/`.
- `AuthService.logout()` also calls `POST /api/users/logout` (fire and forget) so the cookie is cleared.
- Route, guard and page unchanged.

## Testing Decisions

Levels needed: unit tests (user, ui), deployment check, manual.

Seams: the user service HTTP API through FastAPI `TestClient` (existing pattern in `test_metrics.py`: stub `db_utils` and `kafka_client`; DB replaced with in-memory SQLite via `app.dependency_overrides[get_db]`), and the gateway over HTTP in the deployment check. UI: `INFRA_TOOLS` and `AuthService.logout` specs. No gateway unit tests (nginx has no test framework here; YAGNI).

- Unit, test-first (`pytest` in `user/`): login sets the cookie with HttpOnly and SameSite; `admin-check` is 200 with `X-Auth-User` for admin via cookie, 403 for host and guest, 401 with no token and with a bad token; Bearer still works; register with `role=admin` returns 422; seed creates the admin once and is a no-op on the second run; startup without `JWT_SECRET` fails.
- Unit (`npm run test -- --watch=false`): `INFRA_TOOLS` urls all start with `/admin/`; `logout()` posts to `/api/users/logout`.
- Manifests/gateway: `nginx -t` in the built gateway image; `kubectl apply --dry-run=server -k k8s/infra/admin-tools`; `grep -rn nodePort k8s/infra` shows none for the four tools.
- Deployment (after `scripts/admin-secrets.ps1`, apply, `redeploy.ps1` for `user`, `gateway`, `ui`), with curl against `localhost`:
  - each `/admin/<tool>/` with no cookie gives 401, with a host cookie 403, with the admin cookie 200 or 302 into the tool;
  - `/admin/grafana/api/user` as admin returns the admin email (signed in without password);
  - a non-admin client sending `X-WEBAUTH-USER` still gets 401/403;
  - old NodePorts 30901, 30090, 30808, 30300 refuse connections;
  - `POST /api/users/register` with `role=admin` gives 422;
  - a token forged with the old default secret is rejected;
  - pods Ready, 0 restarts after 2 minutes; `kubectl top pod` before/after recorded.
- Manual (human only): log in as the seeded admin, click each of the four buttons; Grafana opens signed in; MinIO and Adminer pages render fully under the sub-path and still ask for their own login; Console topics list loads; after logout `/admin/grafana/` returns 401.

## Out of Scope

- TLS/ingress rollout (the `Secure` cookie flag exists but is off locally); production exposure.
- OpenID/IdP for MinIO; SSO for Adminer's DB login.
- Rotating the Postgres password.
- An endpoint to create more admins (seed only).
- Gateway unit tests; per-tool authorization beyond `admin`.

## Further Notes

- Self-registered `admin` rows already in the DB remain admins. Operator check: `SELECT email FROM users WHERE role='admin'` in Adminer, then demote with an `UPDATE`. Not automated.
- The JWT secret fix goes beyond the issue text, but without it the admin gate is bypassable with a forged token, so it is in scope.
- CSRF: SameSite=Lax blocks cross-site POSTs to the tools; their own CSRF protections apply.
- `X-Auth-User` is a response header of an internal subrequest; clients never see it.

### Adversarial review outcome

Accepted: leftover self-registered admins (demoted at startup); seed collision and password rotation; gateway crashloop on missing upstream (resolver + variables); `nginx -t` AC dropped; `JWT_SECRET` vs existing test and CI; Grafana prefix, Host and probe; cookie stripped upstream; `/admin/<tool>` slash redirect; `auto_error=False` and route order; body size and buffering; default.conf vs configmap drift stated; no pruning on `kubectl apply -k`; spoofed header AC; sub-path sentinel AC; supported origin; re-login after deploy; Grafana trust risk stated as accepted.
Rejected: `auth_request` caching (YAGNI, single admin user, revisit if the user service shows load); login rate limiting (random 24-char password, separate concern); deny-all `/admin/*` (would break the UI route `/admin/tools`); reconciling old default.conf/configmap drift (out of scope).
