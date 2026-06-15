# Kubernetes Deployment Cheat Sheet

This service is containerized and deployed to the local Kubernetes environment.

## Creating the Environment Secret
Before running the service in Kubernetes, you must create a secret named `media-env` with the `.env` file content:
```bash
kubectl create secret generic media-env --from-file=.env=media/.env -n raas
```

## Rebuilding the Image
To rebuild the Docker image from the repository root:
```bash
docker build -t raas/media-service:latest ./media
```

## Reapplying the Rollout
To restart the deployment and force Kubernetes to run pods with the new local image:
```bash
kubectl rollout restart deployment media-deployment -n raas
```

To watch the status of the rollout:
```bash
kubectl rollout status deployment media-deployment -n raas
```
