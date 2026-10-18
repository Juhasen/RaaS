# Validation: prometheus-grafana-monitoring

Run on the live Docker Desktop cluster (context `docker-desktop`, namespace `raas`), 2026-10-18. All eleven images were rebuilt from this branch with `scripts/build-and-start.ps1`.

| Level | What was run | Result |
|---|---|---|
| Manifests | `kubectl kustomize k8s/infra/monitoring`, `kubectl apply --dry-run=server -k k8s/infra/monitoring` | renders, dry-run passes |
| Apply | `kubectl apply -k k8s/infra/monitoring` | prometheus, grafana, kube-state-metrics, node-exporter, kafka/redis/postgres/mongodb exporters Running |
| Targets | Prometheus `/api/v1/targets` | 16 of 16 `up`: 9 services (3 go, 3 java, 3 python), kube-state-metrics, node-exporter, kafka, redis, postgres, mongodb exporters, prometheus |
| Grafana provisioning | Grafana API `/api/datasources`, `/api/search?type=dash-db` | 1 data source (Prometheus, default), 3 dashboards in folder `RaaS`: RaaS Cluster, RaaS Infra, RaaS Services |
| Dashboard queries | every `expr` of the three dashboards run against Prometheus | all return `success`; see below for the two empty ones |
| Go tests | `go vet ./...`, `go test ./...` in `listing/`, `media/`, `booking/` | pass |
| Python test | `pytest` in `user/` inside `python:3.11-slim` with `pip install -r requirements.txt pytest` (same as CI) | pass after adding `httpx` to requirements; it failed in CI before (`No module named httpx`) |
| Java | images build in Docker (`mvn package`); `payment`, `review`, `favorites` targets `up`; `jvm_memory_used_bytes` and `http_server_requests_seconds_count` return series | pass (no Java test directory exists) |
| Media with the #41 image | media target | `down` while the MinIO branch image runs (it has no `/metrics`); `up` with this branch's image, 41 `media_*` series |

## Fixed during validation
1. `node-exporter` pod was rejected by Docker Desktop with `HostToContainer` on the `/` mount; the option was dropped (`ce8bf2b`).
2. The "Root disk used" panel was empty: with no `HostToContainer` mount there is no `mountpoint="/"` series. It now uses `/var/lib` (the Docker VM data disk) and is titled "Data disk used (Docker VM, /var/lib)".

## Dashboard panels with no data
- 5xx rate (Services): empty because no service returned a 5xx. Expected.
- Kafka consumer lag needs consumer-group traffic; the series exist (11 groups) but values depend on activity.

## Not verified
- Nobody opened the dashboards in a browser; queries were checked through the Prometheus API, not rendered in Grafana.
- Go panel for request duration matches `<service>_request_duration_seconds` by regex because echoprometheus prefixes the service name; checked to return series (4 hosts), not visually.
- Notification and analytics have no `/metrics` test; they are covered only by their live targets being `up`.
- Filesystem metrics describe the Docker Desktop VM, not the Windows host.
- Prometheus and Grafana use `emptyDir`; history and any UI edits are lost on pod restart.

## Code review (`code-review`, high)
Fixed:
- kube-state-metrics could list every Secret and ConfigMap cluster-wide. The ClusterRole is cut to nodes, pods, namespaces, deployments, replicasets, daemonsets and statefulsets, and the binary is limited with `--resources=`. Re-checked live: pod phase, restart and deployment series still return data.
- Declared kube-state-metrics port 8081, used by the readiness probe.
- Grafana memory: request 128Mi, limit 512Mi (was 256Mi limit only).

Accepted and left as is (documented in the spec):
- Exporters use the app DB user and inline credentials, like the existing infra manifests.
- Grafana `admin`/`admin` secret; ClusterIP only, reached by port-forward.
- Static scrape targets scrape one replica per Service; revisit with HPA.
- `httpx` in `user/requirements.txt`; moving it to a dev file needs a change to the shared CI workflow.
- The Go latency panel matches the metric name by regex.
- The disk panel mountpoint `/var/lib` is specific to Docker Desktop.
- No limits on Prometheus and the exporters.

The finding that the branch carries unrelated agent-workflow changes was a false positive: it diffed against a stale local `main`; the PR diff (`gh pr diff 46`) contains only monitoring and instrumentation files.
