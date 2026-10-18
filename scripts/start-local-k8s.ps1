# Deploys RaaS to the local cluster (images must exist: see build-and-start.ps1).
# Optional layers: -Monitoring (Prometheus, Grafana), -Logging (Loki, Alloy), -AdminTools (Adminer, Redpanda Console).
param (
    [switch]$Monitoring,
    [switch]$Logging,
    [switch]$AdminTools
)

# Continue: expected native stderr (CREATE DATABASE on an existing database) must not abort the script.
$ErrorActionPreference = 'Continue'
Set-Location (Resolve-Path "$PSScriptRoot\..")

Write-Host "Starting local Kubernetes environment for RaaS..."

kubectl cluster-info *> $null
if ($LASTEXITCODE -ne 0) { throw "kubectl cannot reach a cluster. Start Kubernetes in Docker Desktop first." }

Write-Host "Applying namespace..."
kubectl apply -f k8s/infra/namespace.yaml

Write-Host "Generating credentials (.passwords) and Secrets..."
& "$PSScriptRoot\admin-secrets.ps1"

Write-Host "Applying media environment secret..."
if (Test-Path "media/.env") {
    kubectl create secret generic media-env --from-file=.env=media/.env -n raas --dry-run=client -o yaml | kubectl apply -f -
} else {
    Write-Warning "media/.env not found! Creating media-env secret from media/.env.example template instead. Please configure it later."
    kubectl create secret generic media-env --from-file=.env=media/.env.example -n raas --dry-run=client -o yaml | kubectl apply -f -
}

Write-Host "Applying payment environment secret..."
if (Test-Path "payment/.env") {
    kubectl create secret generic payment-env --from-env-file=payment/.env -n raas --dry-run=client -o yaml | kubectl apply -f -
} else {
    Write-Warning "payment/.env not found! Creating payment-env secret from payment/.env.example template instead. Please configure it later."
    kubectl create secret generic payment-env --from-env-file=payment/.env.example -n raas --dry-run=client -o yaml | kubectl apply -f -
}


Write-Host "Applying notification environment secret..."
if (Test-Path "notification/.env") {
    kubectl create secret generic notification-env --from-env-file=notification/.env -n raas --dry-run=client -o yaml | kubectl apply -f -
} else {
    Write-Warning "notification/.env not found! Creating notification-env secret from notification/.env.example template instead. Please configure it later."
    kubectl create secret generic notification-env --from-env-file=notification/.env.example -n raas --dry-run=client -o yaml | kubectl apply -f -
}

Write-Host "Applying infrastructure configmap..."
kubectl apply -f k8s/infra/configmap.yaml

Write-Host "Applying databases and message brokers..."
kubectl apply -f k8s/infra/postgres.yaml
kubectl apply -f k8s/infra/mongodb.yaml
kubectl apply -f k8s/infra/redis.yaml
kubectl apply -f k8s/infra/kafka.yaml
kubectl apply -f k8s/infra/minio.yaml

Write-Host "Waiting for infrastructure to become ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=mongodb -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=kafka -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=minio -n raas --timeout=120s

Write-Host "Initializing Postgres databases for microservices..."
$pgPod = (kubectl get pods -n raas -l app=postgres -o jsonpath="{.items[0].metadata.name}")
$dbs = "payment", "review", "favorites", "user_db", "notification", "analytics"
foreach ($db in $dbs) {
    # We suppress errors because CREATE DATABASE will error if the DB already exists, which is fine
    kubectl exec -n raas $pgPod -- psql -U raas_user -d raas_db -c "CREATE DATABASE $db;" 2>$null
}

Write-Host "Applying applications (Kustomize local overlay)..."
kubectl apply -k k8s/overlays/local

Write-Host "Waiting for all applications to become ready..."
kubectl wait --for=condition=ready pod -l app=listing-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=media-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=booking-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=payment-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=review-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=favorites-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=user-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=notification-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=analytics-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=ui-service -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=gateway -n raas --timeout=120s

if ($Monitoring) { kubectl apply -k k8s/infra/monitoring }
if ($Logging) { kubectl apply -k k8s/infra/logging }
if ($AdminTools) { kubectl apply -k k8s/infra/admin-tools }

Write-Host "All services are running!"
Write-Host ""
Write-Host "App:          http://localhost:30000"
Write-Host "Admin login:  RAAS_ADMIN_EMAIL / RAAS_ADMIN_PASSWORD in .passwords"
if ($AdminTools) { Write-Host "Admin tools:  http://localhost:30000/admin/tools (admin only)" }
