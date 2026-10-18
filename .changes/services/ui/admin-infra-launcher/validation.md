# Validation: admin-infra-launcher

| Level | What was run | Result |
|---|---|---|
| Unit | `npx ng test --watch=false` in `ui` | pass, 12 tests (guard 5, component 1, shell 4, app 2) |
| Build | `npx ng build` in `ui` (SSR, prerender) | pass |
| Manifests | `kubectl apply --dry-run=server -k k8s/infra/admin-tools`, dry-run of edited `minio.yaml` and `grafana.yaml` | pass |
| Secrets | `scripts/admin-secrets.ps1` twice | first run created `.passwords` and 3 Secrets and restarted minio, media, loki, grafana; second run: all Secrets `unchanged`, no restarts. `git check-ignore .passwords` prints `.passwords` |
| Deployment | `kubectl apply -k k8s/infra/admin-tools`, rollout status for the 3 new and 4 restarted deployments | all Ready, 0 restarts; all 4 new NodePort Services have endpoints; nodePorts unique |
| Endpoints | `curl` against localhost | Adminer 30090 200; MinIO console 30901 200; Grafana 30300 302 (to /login); Grafana API 200 with generated password, 401 with `admin:admin`; Console 30808: `/` and `/api/topics` 401 without credentials, 401 with `admin:admin`, 200 with the generated password; `/api/topics` lists the Kafka topics |
| Consumers | `kubectl logs` after the secret change | media: "S3 bucket raas-media-bucket ready"; loki Ready |
| Memory (measured in container) | cgroup `memory.current` / `usage_in_bytes` | adminer 12.8 Mi, redpanda-console-auth 14.5 Mi, redpanda-console 49.8 Mi (cgroup; `VmRSS` of the Go process reads 125 Mi). All below their limits (96 / 32 / 256 Mi) |
| UI redeploy | `scripts/redeploy.ps1 -Service ui` | rollout complete |
| Manual | log in as an admin, check the Admin link, click each of the 4 buttons and log in with the values from `.passwords`; log in as host/guest and check the link is absent and `/admin/tools` redirects to `/` | pending the user |

## Not verified

- Browser flow of the new page (guard redirect on hard refresh, link clicks): only unit-tested; manual row above.
- `kubectl top pod`: no metrics-server in the cluster; memory read from the container cgroup instead.
- Console behaviour over WebSocket/SSE live views through the proxy: config is in place, not exercised.
- The `.passwords` file lives in the worktree (`..\RaaS-admin-infra-launcher`); the cluster Secrets already hold the generated values. Running the script from another checkout creates a new `.passwords` and rotates them.

## After code review
Re-ran `scripts/admin-secrets.ps1` (all Secrets `unchanged`; with `.passwords` moved away it refuses with "is missing but the cluster already has credentials") and `ng test` (12 pass).

## Media URL fix (added during manual testing)
Uploaded images were stored as `http://localhost:9000/...`, unreachable without a port-forward. Now `S3_PUBLIC_URL=/media-files`, served by a new gateway route to the MinIO bucket (same origin, host independent).

| Level | What was run | Result |
|---|---|---|
| Deployment | applied gateway ConfigMap and `media.yaml`, restarted both | rollouts complete |
| Endpoint | `curl localhost:30000/media-files/<key>` | 200 `image/jpeg` |
| Data | `scripts/migrate-media-urls.ps1` (listing `raas.listings` 1 doc, `media_db.media` 3 docs); re-run prints nothing | listing now returns `/media-files/...` |
| Not verified | a fresh upload through the UI as a host, and the image rendering in the browser | pending the user |
