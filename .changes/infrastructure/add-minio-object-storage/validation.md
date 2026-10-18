# Validation: add-minio-object-storage

| Level | What was run | Result |
|---|---|---|
| Unit / integration | `go vet ./...` and `go test ./...` in `media/` (new `storage_test.go`: red first, then green) | pass |
| Syntax | YAML parse of `k8s/infra/minio.yaml` | 5 documents parse |
| Deployment | `kubectl apply --dry-run=server`, apply infra, redeploy media, pods Ready, upload via gateway, object in MinIO | NOT RUN: Docker Desktop and the Kubernetes API (127.0.0.1:6443) were down |
| Manual | port-forward console (9001), log in with Secret credentials; port-forward 9000 and open an uploaded image URL | pending the user |

## Not verified
- Server-side dry run, live deploy of MinIO and media, and the end-to-end upload: cluster unreachable. Run `kubectl apply -R -f k8s/infra`, `kubectl apply -k k8s/overlays/local`, `powershell -File scripts/redeploy.ps1 -Service media`, then POST an image to `/api/media/upload`.
- Image tag `RELEASE.2025-04-22T22-12-26Z` not pulled (Docker down).
- Persistence across MinIO pod deletion and `kubectl get pv minio-pv` Bound.
- Bucket anonymous-read policy against real MinIO (tested only against a fake S3 that accepts any PUT).
