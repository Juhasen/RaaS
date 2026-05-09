Write-Host "Starting local Kubernetes environment for RaaS..."

Write-Host "Applying namespace..."
kubectl apply -f k8s/infra/namespace.yaml

Write-Host "Applying infrastructure configmap..."
kubectl apply -f k8s/infra/configmap.yaml

Write-Host "Applying databases and message brokers..."
kubectl apply -f k8s/infra/postgres.yaml
kubectl apply -f k8s/infra/mongodb.yaml
kubectl apply -f k8s/infra/redis.yaml
kubectl apply -f k8s/infra/kafka.yaml

Write-Host "Waiting for infrastructure to become ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=mongodb -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=kafka -n raas --timeout=120s

Write-Host "Initializing Postgres databases for microservices..."
$pgPod = (kubectl get pods -n raas -l app=postgres -o jsonpath="{.items[0].metadata.name}")
$dbs = "payment", "review", "favorites", "user_db", "notification", "analytics"
foreach ($db in $dbs) {
    # We suppress errors because CREATE DATABASE will error if the DB already exists, which is fine
    kubectl exec -n raas $pgPod -- psql -U raas_user -d raas_db -c "CREATE DATABASE $db;" 2>$null
}

Write-Host "Applying Go applications..."
kubectl apply -f k8s/apps/go/

Write-Host "Applying Java applications..."
kubectl apply -f k8s/apps/java/

Write-Host "Applying Python applications..."
kubectl apply -f k8s/apps/python/

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

Write-Host "All services are running!"
