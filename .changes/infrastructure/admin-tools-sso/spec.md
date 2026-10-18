# Admin tools SSO through the gateway

Issue: #52 (follow-up to #50 / PR #51). Spans `gateway`, `user`, `ui` and `k8s/infra/admin-tools/`; filed under infrastructure.

## Problem Statement

Admin infra tools (MinIO, Grafana, Adminer, Redpanda Console) sit on their own NodePorts, so the browser treats each as another origin: the RaaS token (in `localStorage`, sent only by the app's own requests) never reaches them and each asks for its own password. Nothing server-side limits who can reach them, and anyone can self-register as `admin`. The JWT secret also has a hard-coded fallback (`raas-secret-key-12345`, not set in any manifest), so admin tokens can be forged.

## Solution

Serve the tools through the gateway on the UI's origin at `/admin/{grafana,minio,adminer,kafka}/`. Every `/admin/*` request passes an `auth_request` that accepts only `role == admin`. Login additionally sets an HttpOnly cookie so plain link clicks authenticate. Grafana trusts a gateway-set user header (auth-proxy mode). Tool NodePorts and the Redpanda basic-auth proxy are removed. `admin` can no longer be self-registered; one admin is seeded from a generated Secret.

## User Stories

- As an admin, I want each tool button to open the tool without a RaaS re-login, so that I use one session.
- As an admin, I want Grafana to sign me in automatically, so that I never type a Grafana password.
- As a guest, host or logged-out visitor, I want none of the four tool paths to be reachable (401 logged out, 403 other roles), so that infra tools are not exposed.
- As an operator, I want `admin` unobtainable through registration, so that the gateway check protects something.
- As an operator, I want tokens signed with a generated secret, so that admin tokens cannot be forged.
- As an operator, I want the tools off their NodePorts, so that the gateway is the only way in.

## Implementation Decisions

### user (`user/`)

- Login (`POST /users/login`) also sets cookie `raas_token`: HttpOnly, SameSite=Lax, Path=/, Max-Age = token lifetime, `Secure` only when env `COOKIE_SECURE=true` (default off; TLS is out of scope). Body unchanged, so the UI keeps working.
- `POST /users/logout` clears the cookie (204).
- `get_current_user` takes the token from the Bearer header, else the cookie. Header wins. `HTTPBearer(auto_error=False)` so a missing header gives 401, not FastAPI's default 403. `GET /users/admin-check` is declared before `GET /users/{user_id}`.
- `GET /users/admin-check`: 200 for `role == admin` with response header `X-Auth-User: <email>`, 403 for other roles, 401 for no or invalid token. Reads the role from the DB row, not the token claim, so a demoted user loses access immediately.
- Register/create: `UserCreateRequest.role` pattern becomes `^(guest|host)$`; `role=admin` gives 422.
- Startup `seed_admin(session)` (a plain function, called at startup, unit-tested directly). `ADMIN_PASSWORD` must be at least 8 characters, else startup fails. Steps, idempotent:
  1. Demote every user with role `admin` other than `ADMIN_EMAIL` to `guest` (removes earlier self-registered admins; logged).
  2. No user with `ADMIN_EMAIL`: create it as `admin` (bcrypt). No `user.created` event (analytics does not need the seed).
  3. User with `ADMIN_EMAIL` exists with role `admin`: set its password hash to `ADMIN_PASSWORD` if it differs (so `-Rotate` takes effect).
  4. User with `ADMIN_EMAIL` exists with another role: log an error and do not promote it; no admin exists until the operator fixes the email.
- `JWT_SECRET` has no fallback: the service fails to start without it. `user/conftest.py` sets a test value so `test_metrics.py` and CI (`pytest` in `service.yml`) still import `main`. Provided by Secret `user-secrets` (`JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`), added to `user.yaml` via `secretKeyRef`. If other services verify JWTs (`grep -rn JWT_SECRET` at implementation), they read the same Secret.

### gateway (`gateway/default.conf` and `k8s/apps/gateway/gateway-configmap.yaml`, kept identical)

- Internal location `/api/auth/verify-admin` rewrites to `/users/admin-check`, forwards `Cookie` and `Authorization`, no body.
- Upstreams use `set $var` plus `resolver kube-dns.kube-system.svc.cluster.local valid=10s;` (variables, so nginx resolves at request time). The gateway must start and serve `/api/*` when `grafana`, `adminer` or `redpanda-console` do not exist. With variables `proxy_pass` does not rewrite the URI, so each block uses `rewrite ... break` as existing blocks do.
- Truth for the cluster is `k8s/apps/gateway/gateway-configmap.yaml` (it is mounted over the image config). `gateway/default.conf` already differs from it; this change adds the same new blocks to both and does not reconcile the old differences.
- `location /admin/<tool>/` blocks (the UI route `/admin/tools` is untouched), each with `auth_request /api/auth/verify-admin;`, plain 401/403 (no redirect), `Host $host` and `X-Forwarded-*`, WebSocket upgrade headers, `proxy_read_timeout 300s`, `client_max_body_size 0`, `proxy_buffering off` (SSE, uploads), and `proxy_set_header Cookie` with `raas_token` removed so the JWT never reaches a tool. `location = /admin/<tool>` returns 301 to the slash form:
  - `/admin/grafana/` to the Grafana Service, prefix NOT stripped (`serve_from_sub_path`). `auth_request_set $auth_user $upstream_http_x_auth_user;` then `proxy_set_header X-WEBAUTH-USER $auth_user;`, which overwrites any client-sent value.
  - `/admin/minio/` to the MinIO console (9001), `/admin/adminer/` to `adminer:8080`, `/admin/kafka/` to `redpanda-console:8080`.
- Base paths per tool:
  - Grafana: `GF_SERVER_ROOT_URL=%(protocol)s://%(domain)s/admin/grafana/`, `GF_SERVER_SERVE_FROM_SUB_PATH=true`, `GF_AUTH_PROXY_ENABLED=true`, `GF_AUTH_PROXY_HEADER_NAME=X-WEBAUTH-USER`, `GF_AUTH_PROXY_AUTO_SIGN_UP=true`, `GF_USERS_AUTO_ASSIGN_ORG_ROLE=Admin`.
  - MinIO console: gateway strips the prefix, `MINIO_BROWSER_REDIRECT_URL` points at `/admin/minio`.
  - Redpanda Console: `SERVER_BASEPATH=/admin/kafka` (exact keys for the pinned v3.3.0 verified at implementation).
  - Adminer: no base-path setting; the gateway strips the prefix and relative links are checked in the deployment check.
  - Every tool must work under its sub-path (user decision: no dropping). If a setting is not enough, make it work in the gateway (`sub_filter` rewrite of absolute links, `proxy_redirect`, cookie path) and keep it documented in the gateway block.
- Grafana's own login stays enabled and its Service is ClusterIP only. Accepted risk: any pod in the cluster can still send `X-WEBAUTH-USER` to `grafana:3000` (no NetworkPolicy, and Docker Desktop does not enforce one); the same holds for Console. Fine for the local single-tenant cluster; a NetworkPolicy restricting these Services to the gateway is a required follow-up before a cloud deploy. The proxied identity is an Org Admin login `admin@raas.local`, not Grafana server admin; the `grafana-admin` credentials stay the server-admin path. Grafana's readiness probe path is re-checked under the sub-path and adjusted if it breaks.
- Users logged in before the deploy have no cookie: they must log in again. Logout is a fire-and-forget POST; a failed call leaves the cookie valid until its 24h expiry (no revocation, same as the JWT today).
- Supported origin: the Traefik ingress on `localhost:80` (and gateway NodePort 30000). The UI NodePort 30080 does not serve `/admin/<tool>/`.
- MinIO and Adminer keep their own login (user decision): the gateway gate keeps non-admins out but a second password remains. Redpanda Console has no login, so the gate is its only protection.
- The Traefik ingress already forwards `/` to the gateway; unchanged.

### cluster (`k8s/infra/`)

- `kubectl apply -k` does not prune: the script/README run explicit `kubectl -n raas delete` for the old NodePort Services and `redpanda-console-auth` Deployment/ConfigMap/Secret. Delete `admin-tools/nodeports.yaml` and the `redpanda-console-auth` Deployment, ConfigMap and Secret handling (manifest and `scripts/admin-secrets.ps1`). The `adminer` Service becomes ClusterIP. Grafana and MinIO console get ClusterIP Services if the existing ones lack the port; no NodePort remains for the four tools.
- `scripts/admin-secrets.ps1` also generates `JWT_SECRET` (random; reused from `.passwords` unless `-Rotate`, and a `.passwords` lacking the new keys gets them added without touching existing ones), `ADMIN_EMAIL` (`admin@raas.local`) and `ADMIN_PASSWORD`, writes Secret `user-secrets`, restarts `user-deployment`, and lists the admin login in `.passwords`. Rotating `JWT_SECRET` logs everyone out; accepted.
- Memory: no new pods; `redpanda-console-auth` (16Mi request, 32Mi limit) is removed. Before/after `kubectl top` recorded.

### ui (`ui/`)

- `config/infra-tools.ts`: urls become `/admin/grafana/`, `/admin/minio/`, `/admin/adminer/`, `/admin/kafka/`.
- `AuthService.logout()` also calls `POST /api/users/logout` (fire and forget) so the cookie is cleared.
- Route, guard and page unchanged.

## Testing Decisions

Levels needed: unit tests (user, ui), deployment check, manual.

Seams: the user service HTTP API through FastAPI `TestClient` (existing pattern in `test_metrics.py`: stub `db_utils` and `kafka_client`; DB replaced with in-memory SQLite via `app.dependency_overrides[get_db]`), and the gateway over HTTP in the deployment check. UI: `INFRA_TOOLS` and `AuthService.logout` specs. No gateway unit tests (nginx has no test framework here; YAGNI).

- Unit, test-first (`pytest` in `user/`): login sets the cookie with HttpOnly and SameSite; `admin-check` is 200 with `X-Auth-User` for admin via cookie, 403 for host and guest, 401 with no token and with a bad token; Bearer still works; register with `role=admin` returns 422; seed creates the admin once and is a no-op on the second run; startup without `JWT_SECRET` fails (import in a subprocess); seed demotes other admins, creates, syncs the password, and refuses to promote an existing non-admin with that email.
- Unit (`npm run test -- --watch=false`): `INFRA_TOOLS` urls all start with `/admin/`; `logout()` posts to `/api/users/logout`.
- Manifests/gateway: `kubectl apply --dry-run=server -k k8s/infra/admin-tools`; `grep -rn nodePort k8s/infra` shows none for the four tools.
- Deployment (after `scripts/admin-secrets.ps1`, apply, `redeploy.ps1` for `user`, `gateway`, `ui`), with curl against `localhost`:
  - the gateway starts and `/api/listings` answers while the `adminer` Service is deleted (then restored);
  - each `/admin/<tool>/` with no cookie gives 401, with a host cookie 403, with the admin cookie 200 plus one tool asset under the prefix returning 200 with the right content type (sub-path sentinel); a tool failing this is fixed in the gateway before moving on;
  - after the deploy, a self-registered pre-existing admin is `guest` and gets 403;
  - the `raas_token` cookie is absent from requests the tool receives (gateway debug log);
  - `/admin/grafana/api/user` as admin returns the admin email (signed in without password);
  - a non-admin client sending `X-WEBAUTH-USER` gets 401/403, and the admin sending a different `X-WEBAUTH-USER` still appears in Grafana as `admin@raas.local`;
  - old NodePorts 30901, 30090, 30808, 30300 refuse connections and `kubectl -n raas get svc,deploy` shows the old objects gone;
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

- Earlier self-registered admins are demoted to `guest` by the startup seed (one-way; they can be re-promoted only by editing the DB).
- The JWT secret fix goes beyond the issue text, but without it the admin gate is bypassable with a forged token, so it is in scope.
- CSRF: SameSite=Lax blocks cross-site POSTs to the tools; their own CSRF protections apply.
- `X-Auth-User` is a response header of an internal subrequest; clients never see it.

### Adversarial review outcome

Accepted: leftover self-registered admins (demoted at startup); seed collision and password rotation; gateway crashloop on missing upstream (resolver + variables); `nginx -t` AC dropped; `JWT_SECRET` vs existing test and CI; Grafana prefix, Host and probe; cookie stripped upstream; `/admin/<tool>` slash redirect; `auto_error=False` and route order; body size and buffering; default.conf vs configmap drift stated; no pruning on `kubectl apply -k`; spoofed header AC; sub-path sentinel AC; supported origin; re-login after deploy; Grafana trust risk stated as accepted.
Rejected: `auth_request` caching (YAGNI, single admin user, revisit if the user service shows load); login rate limiting (random 24-char password, separate concern); deny-all `/admin/*` (would break the UI route `/admin/tools`); reconciling old default.conf/configmap drift (out of scope).
