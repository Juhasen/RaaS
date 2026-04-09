# Feature Specification: Local Kubernetes Development Environment

**Feature Branch**: `002-kubernetes-local-dev`
**Created**: 2026-04-10
**Status**: Draft  
**Input**: User description: "Create a local testing environment that uses Kubernetes manifests (mirroring production) instead of Docker Compose to better validate how K8s works. Tasks will be split between Go, Java, and Python developers."

## Clarifications

### Session 2026-04-10

- Q: What local Kubernetes provider should be assumed? → A: Assume Docker Desktop with Kubernetes enabled, Minikube, or Kind.
- Q: How will backing services (Databases, Kafka) be deployed locally? → A: Create lightweight Kubernetes manifests for infrastructural dependencies alongside the application services.
- Q: How will developers access the services from their host machine? → A: Use Kubernetes `NodePort` services or document `kubectl port-forward` usage to expose services to localhost.

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Full Stack Local Emulation (Priority: P1)

As a developer, I want to spin up the entire RaaS microservice ecosystem locally using Kubernetes manifests, so that I can perform end-to-end testing in an environment that effectively mirrors production.

**Why this priority**: Testing cross-service interactions and Kafka event propagation in an environment identical to production prevents "works on my machine" deployment issues.

**Independent Test**: Can be fully tested by applying the infrastructure and all service K8s manifests locally (`kubectl apply -f k8s/infra/`, `kubectl apply -f k8s/apps/`) and verifying all pods reach a `Running` state and communicate successfully.

**Acceptance Scenarios**:

1. **Given** a local Kubernetes cluster, **When** the developer applies the infrastructure manifests, **Then** databases and Kafka pods start successfully and become ready.
2. **Given** the infrastructure is running, **When** a developer applies the application manifests, **Then** all 9 microservices boot, connect to the infrastructure, and achieve a healthy state without crashlooping.

---

### User Story 2 - Domain-Specific Local Development (Priority: P2)

As a language-specific developer (Go, Java, or Python), I want to be able to deploy only the infrastructure and my specific track's services to the local Kubernetes cluster, so that I can conserve local system resources.

**Why this priority**: Running the entire ecosystem consumes significant RAM/CPU. Developers often only need their specific track running alongside the database layer.

**Independent Test**: Can be fully tested by applying only the manifests relevant to a specific track (e.g., using directory separation like `k8s/apps/go-services/`) and verifying the isolated pods run without errors.

**Acceptance Scenarios**:

1. **Given** a developer is working only on the Go services, **When** they apply the infrastructure and Go specific manifests, **Then** only the Go services and necessary data infrastructure begin, leaving Java and Python services offline.

### Edge Cases

- What happens if the developer's local Kubernetes cluster lacks sufficient resources (e.g., Docker Desktop RAM limit)? (Needs documentation on minimum limits).
- How do we handle persistent data for local databases in K8s? (Use simple PersistentVolumeClaims (PVCs) bound to hostPath or standard local provisioner).

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST provide Kubernetes manifests (`Deployment`, `Service`, `PVC`) for all backing services locally (PostgreSQL, MongoDB, Redis, Kafka using KRaft mode instead of Zookeeper).
- **FR-002**: System MUST organize the K8s manifests logically to separate the Go, Java, and Python developer tracks (e.g., via distinct folders like `k8s/go/`) so work can be applied independently.
- **FR-003**: System MUST configure local environment variables (e.g., ConfigMaps/Secrets) with DB URIs and Kafka brokers that map to internal Kubernetes DNS names (e.g., `kafka.default.svc.cluster.local`).
- **FR-004**: System MUST expose external access to the microservices to the host machine (via `NodePort` or documented `port-forward` commands) to allow local cURL/Postman testing.
- **FR-005**: Provide a developer guide explaining the commands to deploy to the local cluster and build local images.

### Key Entities

- **Kubernetes Manifests**: The declarative YAML files for Deployments, Services, and PVCs replacing Docker Compose YAMLs.
- **Local Infrastructure Pods**: The local K8s instantiations of Postgres, Mongo, Redis, and Kafka.

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: The entire local stack can be applied and reach `Ready` state in under 5 minutes on a standard developer machine.
- **SC-002**: 100% of the microservices report a healthy status via Kubernetes readiness/liveness probes utilizing the existing `/health` endpoints.
- **SC-003**: Developers can successfully route a test HTTP request from their host machine (`localhost:PORT`) to any running microservice pod.

## Assumptions

- Developers have a local Kubernetes environment running (e.g., Docker Desktop, Minikube, Kind) and `kubectl` configured.
- Local infrastructure pods will use ephemeral or simple local persistent volumes; production-grade storage classes are not needed.
