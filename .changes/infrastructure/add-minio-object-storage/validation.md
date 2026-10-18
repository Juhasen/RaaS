# Validation: add-minio-object-storage

| Level | What was run | Result |
|---|---|---|
| Unit / integration | n/a | manifests only |
| Syntax | YAML parse of `k8s/infra/minio.yaml` | 5 documents parse |
| Deployment | `kubectl apply --dry-run=server`, apply, pod Ready, health curl, persistence check | NOT RUN: Docker Desktop / Kubernetes API (127.0.0.1:6443) was down |
| Manual | port-forward console, log in with Secret credentials | pending the user |

## Not verified
- Server-side dry run and live deploy: cluster unreachable. Run `kubectl apply -R -f k8s/infra`, `kubectl -n raas rollout status deploy/minio`, `kubectl -n raas port-forward svc/minio 9001:9001`.
- Image tag `RELEASE.2025-04-22T22-12-26Z` could not be pulled to confirm it exists.
- Persistence across pod deletion.
