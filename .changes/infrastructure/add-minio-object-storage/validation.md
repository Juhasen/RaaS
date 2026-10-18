# Validation: add-minio-object-storage

Run on the live Docker Desktop cluster (context `docker-desktop`, namespace `raas`), 2026-10-18.

| Level | What was run | Result |
|---|---|---|
| Unit / integration | `go vet ./...`, `go test ./...` in `media/` (storage_test.go, plus `TestNewMediaService_DoesNotBlockOnUnreachableS3` red then green) | pass (6 passed) |
| Apply | `kubectl apply -f k8s/infra/minio.yaml` | pv, pvc, secret, deployment, service created |
| PV binding | `kubectl get pv minio-pv` | `Bound raas/minio-pvc`, 5Gi RWO |
| Pod | `rollout status deploy/minio` | rolled out, pod 1/1 Running, log `Version: RELEASE.2026-09-22T19-25-18Z` |
| Health | port-forward 9000, `curl /minio/health/ready` | 200 |
| Console | port-forward 9001, `GET /` | 200; `POST /api/v1/login` with Secret creds 204, wrong password 401 |
| Bucket + policy | media log `S3 bucket raas-media-bucket ready` against real MinIO; anonymous GET of object 200, anonymous bucket listing 403 | pass |
| Upload | `POST /media/upload` (port-forward to media pod, 1x1 PNG) | 201, url `http://localhost:9000/raas-media-bucket/val-1/<id>.png`; Kafka event published |
| Stored URL | anonymous `curl` of that URL via 9000 port-forward | 200 `image/png`, bytes identical to upload |
| Persistence | `kubectl delete pod -l app=minio`, re-forward, GET same URL | 200, identical bytes |
| Media cold start | scale minio to 0, restart media, wait ~50s, then scale minio to 1 | media 1/1, 0 restarts; bucket became ready after MinIO came up |

## Bugs found and fixed
1. Image `quay.io/minio/minio:RELEASE.2025-04-22T22-12-26Z` does not exist (`ImagePullBackOff`, `docker manifest inspect`: no such manifest); neither do `latest` or other RELEASE tags on quay.io or docker.io. Switched to `cgr.dev/chainguard/minio` pinned by digest (`RELEASE.2026-09-22T19-25-18Z`).
2. That image runs as uid 65532 and the root-owned hostPath gave `unable to create (/data/.minio.sys/tmp) file access denied`. Added pod `securityContext.runAsUser: 0` (fsGroup does not apply to hostPath).
3. `bucketReady` blocked `NewMediaService` for up to ~80s when MinIO was unreachable, so the liveness probe (about 35s) put media into CrashLoopBackOff. Now run as a goroutine.

## Not verified
- Upload through the gateway: tested by port-forwarding straight to the media pod (gateway auth not exercised).
- Browser rendering of the console UI (only the login API and the login page HTTP status were checked).
- Uploads sent in the first seconds after media starts, before the async bucket creation finishes, fail; not tested.
- Note: media from this branch replaces the #42 media image, so `/metrics` on media is absent while it runs.
