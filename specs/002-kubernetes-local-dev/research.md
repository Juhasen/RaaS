# Phase 0: Research & Clarifications

## Technical Unknowns Resolved

1. **How will backing services (Databases, Kafka) be deployed locally?**
   - **Decision**: Create lightweight Kubernetes manifests for infrastructural dependencies directly mimicking production structures, but tuned for local resources (minimal replicas, nodePorts).
   - **Rationale**: Avoids the "works on my machine" discrepancy between Docker Compose and Production Kubernetes networks.
   - **Alternatives considered**: Hybrid approach: Docker Compose for infrastructure and Kubernetes for applications. Rejected due to complex networking requirements bridging the Docker bridge network with the Docker Desktop K8s network.

2. **How to handle persistent data for local databases in K8s?**
   - **Decision**: Use standard PersistentVolumeClaims (PVCs) requesting standard storage (`hostPath` or default provisioners).
   - **Rationale**: Retains data between Pod restarts locally so that mock accounts/listings aren't lost immediately.
   - **Alternatives considered**: Ephemeral `emptyDir`. Rejected due to pain of losing mock data constantly.

3. **How do developers access the services from their host machine?**
   - **Decision**: Use `NodePort` service types for local debugging alongside clear documentation for `kubectl port-forward`.
   - **Rationale**: `NodePort` is immediate upon Pod run for HTTP hitting `localhost:PORT`.
   - **Alternatives considered**: Ingress controllers (NGINX), but installing an Ingress adds unnecessary bulk for simple local track execution.

4. **Resource Limitations (Docker Desktop RAM limit)**
   - **Decision**: Recommend allocating > 8GB RAM for the K8s node if spinning up the entire platform, but provide logical subdivisions (`k8s/apps/go`) so individuals only spin up what they need.
   - **Rationale**: Adheres to SC-001 while retaining usability for constrained laptops.
