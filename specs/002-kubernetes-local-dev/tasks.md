# Implementation Tasks: Local Kubernetes Development Environment

**Branch**: `002-kubernetes-local-dev`
**Spec**: [spec.md](spec.md)

## Phase 1: Setup

- [x] T001 Create namespace and folder structure for local Kubernetes (`k8s/infra`, `k8s/apps/go`, `k8s/apps/java`, `k8s/apps/python`)

## Phase 2: Foundational

- [x] T002 Implement PostgreSQL Deployment, Service (port `5432`), and PVC (`1Gi` hostPath) in `k8s/infra/postgres.yaml`
- [x] T003 [P] Implement MongoDB Deployment, Service (port `27017`), and PVC (`1Gi` hostPath) in `k8s/infra/mongodb.yaml`
- [x] T004 [P] Implement Redis Deployment and Service (port `6379`) in `k8s/infra/redis.yaml`
- [x] T005 Implement Kafka 4.0.1 (KRaft mode) Deployment and Service (port `9092`) in `k8s/infra/kafka.yaml`
- [x] T006 Create shared ConfigMap for infrastructure connection strings (URLs, Kafka broker) in `k8s/infra/configmap.yaml`

## Phase 3: Full Stack Local Emulation [US1]

**Goal**: Spin up the entire RaaS microservice ecosystem locally using K8s manifests
**Test Criteria**: Apply infra and apps, all 9 microservices boot, connect to infra, and achieve `Running` state without crashloops.

- [x] T007 [P] [US1] Move and refactor existing Go microservice manifests (listing, media, booking) into `k8s/apps/go/` and update mapped NodePorts (30001, 30002, 30003)
- [x] T008 [P] [US1] Create Deployment and NodePort Service manifests for Java services (payment, review, favorites) in `k8s/apps/java/` mapping to ports 30011, 30012, 30013
- [x] T009 [P] [US1] Create Deployment and NodePort Service manifests for Python services (user, notification, analytics) in `k8s/apps/python/` mapping to ports 30021, 30022, 30023
- [x] T010 [US1] Create unified bootstrap script `scripts/start-local-k8s.sh` to apply all manifests sequentially and wait for readiness

## Phase 4: Domain-Specific Local Development [US2]

**Goal**: Allow deploying only the infra and specific track's services
**Test Criteria**: Apply only the infrastructure and one application track (e.g., `kubectl apply -f k8s/apps/go/`), verifying isolated operation.

- [x] T011 [P] [US2] Update `README.md` to reference the local Kubernetes domain-specific execution workflow

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T012 Verify memory consumption limits of the local K8s cluster and document them in `specs/002-kubernetes-local-dev/quickstart.md`
- [x] T013 Consolidate environment variables into the ConfigMap to ensure no hardcoded localhost ports are present in the service manifests

## Dependencies

- **T001** must complete before all others.
- **T002-T005** are independent and can be executed in parallel.
- **T006** depends on **T002-T005** to finalize internal service URLs.
- **T007-T009** depend on **T006**.
- **US1** must complete before **US2**.

## Parallel Execution Examples

1. **Developer A** works on `k8s/infra/` manifests (T002-T006)
2. **Developer B** scaffolds the `k8s/apps/` manifests for the application tracks (T007-T009)

## Implementation Strategy

We will deliver the MVP by establishing the `infra` layer first, validating that the underlying stateful and messaging services are functional. Next, we will deliver the K8s manifests for the application tracks, ensuring each maps effectively to the configuration data stored in `configmap.yaml`, proving that US1 is met before fine-tuning documentation for US2.
