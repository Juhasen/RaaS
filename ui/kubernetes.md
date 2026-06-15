# Kubernetes Deployment Cheat Sheet

This service is containerized and deployed to the local Kubernetes environment.

## Rebuilding the Image
To rebuild the Docker image from the repository root:
```bash
docker build -t raas/ui-service:latest ./ui
```

## Reapplying the Rollout
To restart the deployment and force Kubernetes to run pods with the new local image:
```bash
kubectl rollout restart deployment ui-deployment -n raas
```

To watch the status of the rollout:
```bash
kubectl rollout status deployment ui-deployment -n raas
```
