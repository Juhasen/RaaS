#!/bin/bash
set -e

echo "Starting local Kubernetes environment for RaaS..."

echo "Applying namespace..."
kubectl apply -f k8s/infra/namespace.yaml

echo "Applying media environment secret..."
if [ -f "media/.env" ]; then
    kubectl create secret generic media-env --from-file=.env=media/.env -n raas --dry-run=client -o yaml | kubectl apply -f -
else
    echo "Warning: media/.env not found! Creating media-env secret from media/.env.example template instead."
    kubectl create secret generic media-env --from-file=.env=media/.env.example -n raas --dry-run=client -o yaml | kubectl apply -f -
fi

echo "Applying payment environment secret..."
if [ -f "payment/.env" ]; then
    kubectl create secret generic payment-env --from-env-file=payment/.env -n raas --dry-run=client -o yaml | kubectl apply -f -
else
    echo "Warning: payment/.env not found! Creating payment-env secret from payment/.env.example template instead."
    kubectl create secret generic payment-env --from-env-file=payment/.env.example -n raas --dry-run=client -o yaml | kubectl apply -f -
fi

echo "Applying notification environment secret..."
if [ -f "notification/.env" ]; then
    kubectl create secret generic notification-env --from-env-file=notification/.env -n raas --dry-run=client -o yaml | kubectl apply -f -
else
    echo "Warning: notification/.env not found! Creating notification-env secret from notification/.env.example template instead."
    kubectl create secret generic notification-env --from-env-file=notification/.env.example -n raas --dry-run=client -o yaml | kubectl apply -f -
fi

echo "Applying infrastructure configmap..."
kubectl apply -f k8s/infra/configmap.yaml

echo "Applying databases and message brokers..."
kubectl apply -f k8s/infra/postgres.yaml
kubectl apply -f k8s/infra/mongodb.yaml
kubectl apply -f k8s/infra/redis.yaml
kubectl apply -f k8s/infra/kafka.yaml

echo "Waiting for infrastructure to become ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=mongodb -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis -n raas --timeout=120s
kubectl wait --for=condition=ready pod -l app=kafka -n raas --timeout=120s

echo "Initializing Postgres databases for microservices..."
pgPod=$(kubectl get pods -n raas -l app=postgres -o jsonpath="{.items[0].metadata.name}")
dbs=("payment" "review" "favorites" "user_db" "notification" "analytics")
for db in "${dbs[@]}"; do
    # Suppress errors if DB already exists
    kubectl exec -n raas "$pgPod" -- psql -U raas_user -d raas_db -c "CREATE DATABASE $db;" 2>/dev/null || true
done

echo "Applying Go applications..."
kubectl apply -f k8s/apps/go/

echo "Applying Java applications..."
kubectl apply -f k8s/apps/java/

echo "Applying Python applications..."
kubectl apply -f k8s/apps/python/

echo "Applying UI Deployment..."
kubectl apply -f k8s/apps/ui/

echo "Applying Gateway ConfigMap and Deployment (includes Ingress)..."
kubectl apply -f k8s/apps/gateway/

echo "Restarting gateway to load the updated config..."
kubectl rollout restart deployment/gateway -n raas
kubectl rollout status deployment/gateway -n raas --timeout=120s

echo "Waiting for all applications to become ready..."
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

echo "All services are running!"
