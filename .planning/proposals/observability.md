# Observability and storage additions plan

Course: Virtualization and Cloud Architecture. Goal: more containers in the k8s deployment, each with a clear purpose, plus a story of "observable, autoscaling cloud system".

Status: planned, not started.

## Scope (decided)

1. Prometheus + Grafana (metrics and dashboards)
2. Loki + log collector (central logs)
3. MinIO (S3-compatible object storage)

## 1. Prometheus + Grafana

- Prometheus: scrapes metrics from services and exporters.
- Grafana: dashboards, with Prometheus and Loki as data sources.
- Cluster metrics: `kube-state-metrics`, `node-exporter`.
- Service metrics: Spring Boot Actuator (`/actuator/prometheus`) for Java services, `/metrics` endpoint for Go and Python services (needs a small change per service).
- Infra exporters: `kafka-exporter`, `redis-exporter`, `postgres-exporter`, `mongodb-exporter`.
- Manifests in `k8s/infra/monitoring/`, wired into the existing kustomize layout.
- Open question: plain manifests or `kube-prometheus-stack` Helm chart rendered through kustomize. Plain manifests teach more, Helm is faster.

## 2. Loki + collector

- Loki stores logs. Grafana queries them.
- Collector choice (to decide after reading):
  - Promtail: simple, but in maintenance mode and being replaced.
  - Grafana Alloy: the current recommended agent. One agent can also collect metrics and traces later (OpenTelemetry-compatible).
  - Leaning to Alloy as a DaemonSet unless the config looks too heavy.
- Loki can use MinIO as its chunk store, which ties items 2 and 3 together.

## 3. MinIO

- S3-compatible object storage in the cluster.
- Uses:
  - Backing store for the `media` service (uploads).
  - Loki chunk storage (optional, nice demo).
- Needs: PVC, Secret for root credentials, Service, console exposed via port-forward or ingress.
- Open question: does `media` already use S3 or local disk? Check before wiring.

## Suggested order

1. MinIO (independent, small).
2. Prometheus + Grafana with cluster metrics only.
3. Add service and exporter scraping, build dashboards.
4. Loki + Alloy, add log panels to Grafana.
5. Point Loki at MinIO.

## Later ideas (not in scope now)

- Tempo + OpenTelemetry Collector for tracing.
- Alertmanager.
- Kafka UI, pgAdmin, Mongo Express.
- Ingress controller + cert-manager, HPA + metrics-server, k6 load generator.
- Argo CD, network policies, service mesh.

## To read first

- Grafana Alloy vs Promtail.
- Loki deployment modes (monolithic is enough here).
- `kube-prometheus-stack` vs hand-written manifests.
