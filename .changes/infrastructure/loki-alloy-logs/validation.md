# Validation: loki-alloy-logs

| Level | What was run | Result |
|---|---|---|
| Unit / integration | none | n/a, manifests only, no application code changed |
| Manifests | `kubectl kustomize` and `kubectl apply --dry-run=server -k` for `k8s/infra/logging` and `k8s/infra/monitoring` | pass, both render and validate |
| Deployment | `kubectl apply -k k8s/infra/logging`, `kubectl apply -k k8s/infra/monitoring`, rollout status | loki, alloy, grafana Ready, 0 restarts after about 2 minutes |
| Loki ready | port-forward, `GET /ready` | `ready` |
| All pods collected | Loki `label/app/values` | 23 apps (all `*-service`, gateway, kafka, minio, postgres, redis, mongodb, exporters, prometheus, grafana); `loki` and `alloy` absent |
| Alloy errors | Alloy logs | only connection refused while Loki was starting, no 403 or tail errors |
| Recent lines | `query_range` for `{app="listing-service"}`, and for a freshly started `user-service` pod | lines returned; new pod lines appeared within seconds |
| Survives pod restart | deleted `user-service` pod, queried the old pod name | old pod logs still returned |
| Chunks in MinIO | ephemeral pod running `mc ls -r minio/raas-loki` | chunk objects present about 90s after deploy |
| Retention config | Loki `/config` | `retention_enabled: true`, `retention_period: 1d` |
| Grafana data source | `GET /api/datasources/uid/loki/health` | "Data source successfully connected." |
| Grafana dashboard | `GET /api/search?query=Logs` | "RaaS Logs" listed in folder RaaS |
| Manual | open "RaaS Logs" in Grafana (`kubectl -n raas port-forward svc/grafana 3000`, admin/admin), pick an app, check the three panels | pending the user |

## Not verified

- MinIO "each line once": MinIO wrote no log lines in the window, so duplicates could not be counted. Evidence is indirect: Alloy opened exactly one log stream for the `minio` container (rechecked after the relabel fix).
- After the review fix, Alloy was redeployed and rechecked only for stream count (24, none for loki or alloy) and errors; the earlier query checks were not repeated. Alloy logs 400 "entry too far behind" for old lines after a restart; expected, Loki drops them.
- Retention deletion: data older than 24h cannot be produced in one session; only the loaded config was checked.
- Index files in MinIO were not asserted (tsdb upload lag, per spec).
- Dashboard panels were not rendered; variable queries and panel expressions are untested until the manual check.
