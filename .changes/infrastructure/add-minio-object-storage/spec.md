# Add MinIO object storage

Issue: #41. Source: `.changes/proposals/observability.md` section 3.

## Problem Statement

The cluster has no S3-compatible object store. Loki (later) wants one for chunks, and `media` uploads should eventually land in one.

## Solution

One file, `k8s/infra/minio.yaml`: PV, PVC, Secret (root credentials), Deployment (single node, `/data`), Service. Applied with the rest of `k8s/infra/` (`kubectl apply -R -f k8s/infra`); that directory has no kustomization, so no kustomize edits are needed.

## User Stories

- As an operator, I want MinIO running in namespace `raas`, so that other components can use S3.
- As an operator, I want the console reachable via `kubectl port-forward`, so that I can inspect buckets without an ingress.
- As an operator, I want data on a PVC, so that objects survive pod restarts.
- As an operator, I want root credentials in a Secret, so that consumers can reference them via `secretKeyRef`.

## Implementation Decisions

- Follow `postgres.yaml`/`mongodb.yaml`: hostPath PV `/mnt/data/minio` 5Gi, PVC `minio-pvc`, Deployment `minio` (1 replica, `Recreate` strategy because RWO hostPath).
- Image `quay.io/minio/minio` pinned to a release tag; `args: server /data --console-address :9001`.
- Secret `minio-credentials` with `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` in `stringData`. Dev-only values, same standing as the inline Postgres password in this repo. Passed with `envFrom`.
- Service `minio`: ports 9000 (api) and 9001 (console), ClusterIP. Console via `kubectl -n raas port-forward svc/minio 9001:9001`.
- Readiness/liveness on `/minio/health/ready` and `/minio/health/live` (port 9000).
- Resource requests/limits modest.

## media wiring (user decision: in scope, no follow-up issue)

`media` used Cloudflare R2 with a hard-coded endpoint and fell back to local `./uploads`. Now:

- `Config` drops `R2_*`, adds `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` (default `raas-media-bucket`), `AWS_REGION` (default `us-east-1`), `S3_PUBLIC_URL`.
- `media/service/storage.go`: `newS3Client` (path-style, static credentials; nil if endpoint or keys missing), `ensureBucket` (HEAD, create if missing, anonymous GetObject policy so browsers can load stored URLs), `bucketReady` (up to 10 tries, 3s apart, so media can start before MinIO).
- Stored URL is `S3_PUBLIC_URL/<key>`, default `S3_ENDPOINT/<bucket>/<key>`. In-cluster `minio:9000` is not browser-reachable, so the manifest sets `S3_PUBLIC_URL=http://localhost:9000/raas-media-bucket` and the operator runs `kubectl -n raas port-forward svc/minio 9000:9000`. No ingress, consistent with Out of Scope.
- Local `./uploads` fallback is kept for running without S3 settings (dev, unit tests).
- `k8s/apps/go/media.yaml`: env `S3_ENDPOINT=http://minio:9000`, `S3_PUBLIC_URL`, and credentials from Secret `minio-credentials` via `secretKeyRef`.
- Existing media records keep their old R2 URLs; no migration.
- `scripts/debian/setup-env.sh` still prompts for R2 values; left untouched (unrelated host script), stale prompts are harmless.

## Testing Decisions

- Dry run: `kubectl apply --dry-run=server -R -f k8s/infra`.
- Deployment: pod Ready; `curl` of `/minio/health/ready` through a port-forward returns 200; console port-forward serves the login page.
- Binding: `kubectl get pv minio-pv` shows Bound to `raas/minio-pvc`.
- Persistence: create a bucket/object, delete the pod, object still there.
- Manual (human): log into the console with the Secret's credentials.
- Go (test-first, `media/service/storage_test.go`): env parsing of S3 settings; client is nil without endpoint/keys; `ensureBucket` against an `httptest` fake S3 issues HEAD, create, policy requests with the bucket in the path (path-style).
- Deployment of media: upload an image via the gateway, object appears in the console, URL loads through the 9000 port-forward.

## Out of Scope

- Migrating existing R2 media records or objects.
- Loki integration, ingress for the console, TLS, multi-node/distributed mode, bucket provisioning, non-default credentials management.

## Further Notes

- Questions normally put to the user via grilling were answered with defaults (hostPath PV, ClusterIP + port-forward, no ingress) because this run is non-interactive; the user can overturn them at review.
- Parallel issue #42 (Prometheus/Grafana) touches other files; the single new file avoids merge conflicts.

### Adversarial review outcome

Accepted: PV/PVC pinned with `storageClassName: ""` + `volumeName` and `type: DirectoryOrCreate` so the hostPath PV really binds; Binding check added. Deviations from postgres (Recreate strategy, probes) are deliberate. Image tag `RELEASE.2025-04-22T22-12-26Z` is pinned because upstream stopped publishing community images; it was not pullable during this run (Docker down).
Noted: `namespace.yaml` sorts after `minio.yaml` in `apply -R`; on a fresh cluster apply the namespace first or apply twice (same as every other infra file). `media` originally kept using R2 (superseded: wired in this change).
Rejected: startup probe, extra probe tuning, CPU limit, sizing justification (YAGNI for a local dev store).
