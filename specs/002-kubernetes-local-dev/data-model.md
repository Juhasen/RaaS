# Phase 1: Local Kubernetes Architecture Model

This documents the abstract layout of the Kubernetes resources to fulfill the feature specification, rather than a code-level entity diagram.

## Infrastructure Pods (Databases & Messaging)

These represent the mock external resources.

1. **PostgreSQL**
   - **Pod**: RaaS user/transactional backing DB.
   - **Service Name**: `postgres`
   - **Type**: `ClusterIP` (Internal Access Only)
   - **Port**: `5432`
   - **PVC**: Persistent volume mapping `1Gi`.

2. **MongoDB**
   - **Pod**: Listing and non-schema data DB.
   - **Service Name**: `mongodb`
   - **Type**: `ClusterIP`
   - **Port**: `27017`
   - **PVC**: Persistent volume mapping `1Gi`.

3. **Redis**
   - **Pod**: Key/value caching and Redsync locking server.
   - **Service Name**: `redis`
   - **Type**: `ClusterIP`
   - **Port**: `6379`
   - **PVC**: Shared ephemeral volume is acceptable.

4. **Kafka (KRaft Mode)**
   - **Pods**: A Kafka broker pod running in KRaft mode (no Zookeeper).
   - **Service Name**: `kafka`
   - **Type**: `ClusterIP`
   - **Port**: `9092` (Kafka).

## Application Tracks

These represent the language-specific development scopes.

### Go Services (`k8s/apps/go`)

1. **`listing-service`**: Listens for HTTP, uses `mongodb`, emits/consumes `kafka`. Exposes `NodePort: 30001`.
2. **`booking-service`**: Uses `postgres` and `redis` distributed locks. Exposes `NodePort: 30002`.
3. **`media-service`**: Simulates S3 uploads locally with internal storage endpoints. Exposes `NodePort: 30003`.

### Java Services (`k8s/apps/java`)

1. **`payment-service`**: Exposes `NodePort`.
2. **`review-service`**: Exposes `NodePort`.
3. **`favorites-service`**: Exposes `NodePort`.

### Python Services (`k8s/apps/python`)

1. **`user-service`**: Exposes `NodePort`.
2. **`notification-service`**: Exposes `NodePort`.
3. **`analytics-service`**: Exposes `NodePort`.

## Validation Rules & Cross-Service Constraints

- **Namespaces**: Everything will default to `default` namespace for initial simplicity, relying on standard cluster DNS resolving (e.g., `postgres.default.svc.cluster.local`) mapped to the environment variable configurations of all microservices mapping to the deployment `.env` configmaps.
