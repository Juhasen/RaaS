# Quickstart: Local Kubernetes Development

This guide documents how developers can spin up the RaaS local Kubernetes infrastructure to emulate production behavior directly on their machines.

## Prerequisites

1. Ensure you have a local Kubernetes provider running. Supported providers include:
   - **Docker Desktop** (with kubeadm configured)
2. Verify `kubectl` is installed and connected:

   ```bash
   kubectl get nodes
   ```

## Memory Consumption Limits

Running the full microservices stack locally requires the following system configurations to ensure cluster stability:

- **Full Stack Emulation**: 8GB RAM minimum, 16GB RAM recommended.
- **Single Domain Track**: 4GB RAM minimum.
- **CPU**: 4 virtual CPUs recommended.
- **Note**: The microservices have predefined resource requests (`memory: "128Mi"`) and limits (`memory: "512Mi"`). If nodes lack capacity, you may observe pods in `Pending` or `OOMKilled` states.

## Deploying the Full Stack

If your machine has sufficient resources (minimum 8-16GB allocated to K8s) and you wish to run the entire backend locally, you can apply all components.

1. **Spin up Infrastructure first:**
   Apply the databases, Redis, and Kafka.

   ```bash
   kubectl apply -f k8s/infra/
   ```

   _Wait for all pods to be `Running` before proceeding._

   ```bash
   kubectl get pods -w
   ```

2. **Spin up Applications:**
   Apply all language tracks.
   ```bash
   kubectl apply -f k8s/apps/go/
   kubectl apply -f k8s/apps/java/
   kubectl apply -f k8s/apps/python/
   ```

## Deploying a Single Developer Track (e.g., Go)

If you only need to work on the Go microservices and want to save RAM/CPU, run the isolated stack.

1. **Spin up Infrastructure:**

   ```bash
   kubectl apply -f k8s/infra/
   ```

2. **Spin up ONLY the Go Services:**
   ```bash
   kubectl apply -f k8s/apps/go/
   ```
   _This will boot `listing-service`, `booking-service`, and `media-service`, while keeping Java and Python off._

## Accessing Services Externally

Access services via `NodePort` mapping or `kubectl port-forward`:

```bash
# Example: Port-forward the Listing Service to test endpoints via Postman or cURL
kubectl port-forward svc/listing-service 8080:8080

curl http://localhost:8080/health
```

## Teardown

To shut down the entire environment and delete the pods. Warning: Unless your PVCs are explicitly managed, mock database states might be reset.

```bash
kubectl delete -f k8s/apps/go/
kubectl delete -f k8s/apps/java/
kubectl delete -f k8s/apps/python/
kubectl delete -f k8s/infra/
```
