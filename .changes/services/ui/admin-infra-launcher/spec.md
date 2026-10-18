# Admin infra launcher

Issue: #50. Spans `ui` and `k8s/infra/admin-tools/`; filed under `ui` because the UI is the entry point.

## Problem Statement

Operators open MinIO, Postgres, Kafka and Grafana by port-forwarding by hand. Postgres and Kafka have no web UI deployed at all. There is no single place to reach them, and nothing restricts who sees the links.

## Solution

An admin-only page `/admin/tools` in the Angular UI with one button per tool; each opens the tool's own web UI in a new tab. No tool data is rendered inside RaaS. In the cluster, a new `k8s/infra/admin-tools/` kustomization adds the missing UIs (Adminer, Redpanda Console behind an nginx basic-auth proxy) and NodePort Services for all four tools.

## User Stories

- As an admin, I want a page with a button per infrastructure tool, so that I can open any tool UI in one click.
- As a guest, host or logged-out visitor, I do not see the nav link and cannot reach the page, so that the launcher is not advertised.
- As an operator, I want Adminer and Redpanda Console deployed cheaply, so that Postgres and Kafka have a UI without blowing the 16 GB host budget.
- As an operator, I want Redpanda Console behind basic auth, so that Kafka topics are not browsable anonymously (the Console OSS build has no login).
- As an operator, I want tool URLs in one frontend config object, so that moving to cloud addresses changes one file.
- As an operator, I want one command to deploy everything, so that nothing is configured by hand.

## Implementation Decisions

### Cluster (`k8s/infra/admin-tools/`)

- Own `kustomization.yaml`, applied with `kubectl apply -k k8s/infra/admin-tools`, like `monitoring/` and `logging/`. `k8s/overlays/local` only contains `../../apps`; infra is never applied through it, so the issue's "wire into the local overlay" and "apply -k k8s/overlays/local" are replaced by this command. One README line documents it. No existing manifest is edited.
- **Adminer**: image `adminer:4` (pin the exact 4.x tag at implementation), container port 8080, 1 replica, no PVC. `ADMINER_DEFAULT_SERVER=postgres` (prefills the host only). Login: system PostgreSQL, user `raas_user`, password `raas_password`, db `raas_db` (from `postgres.yaml`). Requests 32Mi, limit 96Mi. Readiness probe `httpGet /` on 8080.
- **Redpanda Console**: image `docker.redpanda.com/redpandadata/console`, exact tag pinned at implementation after checking that the tag reads brokers from `KAFKA_BROKERS` (else use that tag's config key, via ConfigMap) and works against `apache/kafka:4.0.1` (KRaft). Port 8080. Service `redpanda-console` is ClusterIP only. Requests 64Mi, limit 256Mi. Readiness probe `httpGet /admin/health` (or the tag's health path).
- **Auth proxy** `redpanda-console-auth`: separate Deployment, `nginx:stable-alpine` (same image as the gateway), ConfigMap with `auth_basic` + `proxy_pass http://redpanda-console:8080`, Secret holding an htpasswd file (dev default `admin`/`admin`, bcrypt-hashed). Requests 16Mi, limit 32Mi. `proxy_http_version 1.1`, `Upgrade`/`Connection` headers and `proxy_read_timeout 300s` so WebSocket/SSE live views work. The ConfigMap is mounted at `/etc/nginx/conf.d` (replaces `default.conf`) with `listen 80`. The htpasswd hash is generated with `openssl passwd -apr1 admin` (apr1/MD5 is supported by nginx on Alpine; no bcrypt dependency) and committed as the literal value. The nginx here is separate from `gateway`: a plain browser link cannot send the JWT `Authorization` header the gateway's `auth_request` needs, Console would need a base path, and a config bug must not touch the main entry point.
- **NodePorts** (separate Services selecting existing pods, so `minio.yaml` and `grafana.yaml` stay untouched): MinIO console 30901 (selector `app: minio`, port 9001), Grafana 30300 (`app: grafana`, 3000), Adminer 30090 (30080 is the `ui` Service), Redpanda Console auth proxy 30808. All nodePorts in `k8s/` must be unique (`grep -rn nodePort k8s`).
- All new objects set `namespace: raas` explicitly; the kustomization uses no `commonLabels` (it would rewrite the NodePort selectors).
- Every new Deployment sets requests and `limits.memory`; single replica, no persistence, no sidecars.
- Depends on: base infra (`postgres`, `kafka`, `minio`) and `monitoring/` (Grafana) applied first for their buttons to work.

### UI (`ui/`)

- `AuthService`: add `userLoaded` (a promise resolved on both success and error of the `/api/users/me` call, immediately when there is no token, and immediately on the server where `window` is undefined). `login()` already sets the user directly; the guard reads `currentUser()` after the await, never a snapshot taken before. `loadCurrentUser` is async, so a hard refresh on `/admin/tools` would otherwise see no user.
- `adminGuard` (functional `CanActivateFn`): awaits `userLoaded`; role `admin` allows, anything else redirects to `/`.
- Route `admin/tools` lazy-loaded (`loadComponent`) under the shell, `canActivate: [adminGuard]`. `app.routes.server.ts`: add `admin/tools` as `RenderMode.Client` before the `**` entry (first match wins). The SSR shell renders without a user, so the Admin link appears after hydration; accepted.
- `config/infra-tools.ts`: typed `INFRA_TOOLS` array (`id`, `name`, `description`, `url`; `@for` tracks `id`), URLs `http://localhost:<nodeport>`. This assumes a local cluster whose NodePorts are published on `localhost` (Docker Desktop); cloud addresses are a later change to this one file.
- `AdminToolsComponent`: standalone, OnPush, `@for` over `INFRA_TOOLS`, each an `<a target="_blank" rel="noopener">` styled as a button.
- Shell header: "Admin" link to `/admin/tools`, shown only when `currentUser()?.role === 'admin'`.
- Follow `guidelines/TYPESCRIPT.md` and `guidelines/ANGULAR.md`.

## Testing Decisions

Seams: `adminGuard` (with a stubbed `AuthService`) and `AdminToolsComponent`. Both are new, plus the shell header (no spec exists yet) and `AuthService.userLoaded`. Runner is `ng test` (vitest): use `vi.fn` and `TestBed.runInInjectionContext`, not Jasmine idioms; the loading case uses a deferred promise in the stub.

- Unit (`npm run test -- --watch=false`, test-first):
  - Guard: admin allowed; host, guest, logged-out redirect to `/`; user still loading at call time resolves correctly after `userLoaded`.
  - Component: renders one link per `INFRA_TOOLS` entry with the right `href`, `target="_blank"`, `rel="noopener noreferrer"`.
  - Shell: Admin link present for admin, absent for host/guest/logged-out.
- Manifests: `kubectl kustomize k8s/infra/admin-tools` renders; `kubectl apply --dry-run=server -k k8s/infra/admin-tools` passes.
- Deployment: `kubectl apply -k k8s/infra/admin-tools`; pods Ready, 0 restarts; `kubectl top pod` recorded per new pod (needs metrics-server; fall back to Prometheus `container_memory_working_set_bytes`).
  - `curl -i localhost:30808` and `localhost:30808/api/topics` return 401 without credentials, 200 with `admin:admin`; with credentials the topics list renders (Console reaches Kafka).
  - `curl -I`: Adminer 30090 returns 200, MinIO console 30901 returns 200, Grafana 30300 returns 302 to `/login`.
  - `kubectl -n raas get endpoints` shows non-empty endpoints for every new NodePort Service.
  - All nodePorts unique.
  - 0 restarts after 2 minutes.
- Build: `npm run build` (SSR) and the CI lint/test commands from `.github/workflows/ci-ui.yml` pass.
- Redeploy `ui` with `scripts/redeploy.ps1 -Service ui`.
- Manual (human only): log in as an admin (register one via the UI with role admin, see Further Notes), see Admin link, click each of the four buttons and log in to each tool; log in as host/guest and confirm the link is absent and `/admin/tools` redirects.

## Out of Scope

- Embedding tool UIs (iframes, reverse proxy), SSO, ingress/TLS, production exposure, secrets management for the proxy password.
- Changing existing tool credentials (MinIO, Grafana, Postgres keep their dev values).
- Fronting Adminer, Grafana or MinIO with the proxy; they have their own login.
- Prometheus, Loki, MongoDB, Redis buttons (add to `INFRA_TOOLS` later).
- Server-side enforcement: the route guard only hides the page; tools are protected by their own credentials.

## Further Notes

- **Known hole, not fixed here:** `user/schemas.py` lets anyone self-register with `role=admin` (`pattern ^(guest|host|admin)$`). The guard therefore hides the page from honest users but does not stop a determined visitor, who still faces each tool's own login. Fixing registration is a `user` change; to be decided by the user (separate issue suggested).
- Do not deploy `admin-tools` to a cloud cluster as is: default passwords, NodePort exposure, `localhost` URLs. It is a local gate; cloud exposure is a separate decision.
- Total new memory: requests 112Mi (32+64+16), limits 384Mi (96+256+32). Measured values go in `validation.md` and the PR.
- A button whose tool is down shows the browser's error page; accepted, no health indicator.
- No explicit CPU requests; matches the infra files (Grafana has none).

- Security: the proxy's `admin`/`admin` is a dev default, like Grafana's; rotate before any cloud deploy (user decision: Console must have auth because this will be deployed to cloud, not only local). For cloud, only `INFRA_TOOLS` URLs and the exposure method change.
- Grafana already exists in `k8s/infra/monitoring`, so the issue's "waits for observability proposal" no longer applies.
- On a fresh cluster, apply base infra and monitoring before `admin-tools`.

### Adversarial review outcome

Accepted: Adminer port clash (30080 is `ui`, moved to 30090); pinned images and ports; Console/Kafka 4 compatibility check and topics-list AC; concrete WebSocket nginx config and apr1 hash; `userLoaded` on error/server/login; route order in `app.routes.server.ts`; exact curl codes, endpoints, unique nodePorts, build/lint ACs; namespace and no `commonLabels`; probes; higher limits (Adminer 96Mi, Console 256Mi); `noreferrer`; `id` for tracking; self-registration hole recorded; do-not-deploy-to-cloud note.
Rejected: deriving the host from `window.location.hostname` (YAGNI, cloud address is a config change; SSR guard cost); replacing the Console with another UI or dropping its button (user decided Console with auth); CPU requests; test for the server route list; accessibility/empty-state ACs beyond semantic links (links with text satisfy AXE basics; no data states exist).
