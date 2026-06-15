# Kubernetes Deployment Cheat Sheet

This service is containerized and deployed to the local Kubernetes environment.

## Rebuilding the Image
To rebuild the Docker image from the repository root:
```bash
docker build -t raas/notification-service:latest ./notification
```

## Reapplying the Rollout
To restart the deployment and force Kubernetes to run pods with the new local image:
```bash
kubectl rollout restart deployment notification-deployment -n raas
```

To watch the status of the rollout:
```bash
kubectl rollout status deployment notification-deployment -n raas
```
