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

## media finding (open question from the issue)

`media` uses Cloudflare R2 through the S3 SDK, with the endpoint hard-coded to `https://<R2_ACCOUNT_ID>.r2.cloudflarestorage.com` (`media/service/media_service.go`), and falls back to local `./uploads` when R2 creds are absent. Pointing it at MinIO needs an endpoint config option plus path-style addressing. That is a code change in a Go service, so it is out of scope here and filed as a follow-up.

## Testing Decisions

- Dry run: `kubectl apply --dry-run=server -R -f k8s/infra`.
- Deployment: pod Ready; `curl` of `/minio/health/ready` through a port-forward returns 200; console port-forward serves the login page.
- Persistence: create a bucket/object, delete the pod, object still there.
- Manual (human): log into the console with the Secret's credentials.
- No unit tests (manifests only).

## Out of Scope

- Wiring `media` to MinIO (follow-up issue).
- Loki integration, ingress for the console, TLS, multi-node/distributed mode, bucket provisioning, non-default credentials management.

## Further Notes

- Questions normally put to the user via grilling were answered with defaults (hostPath PV, ClusterIP + port-forward, no ingress) because this run is non-interactive; the user can overturn them at review.
- Parallel issue #42 (Prometheus/Grafana) touches other files; the single new file avoids merge conflicts.
