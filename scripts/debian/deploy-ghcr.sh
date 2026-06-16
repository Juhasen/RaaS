#!/bin/bash
set -e

# Get repo root relative to scripts/debian directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
REPO_ROOT="$(dirname "$(dirname "$DIR")")"
cd "$REPO_ROOT"

echo "============================================="
echo "   Deploying RaaS from GHCR to local K3s"
echo "============================================="
echo ""

# Ask for GitHub username
read -p "Enter your GitHub username: " gh_username
gh_username=$(echo "$gh_username" | tr '[:upper:]' '[:lower:]')

echo "Applying namespace..."
kubectl apply -f k8s/infra/namespace.yaml

# Check if image pull secret exists, if not create it
if ! kubectl get secret regcred -n raas >/dev/null 2>&1; then
    echo "Image pull secret 'regcred' not found in namespace 'raas'!"
    echo "Since you chose to use private packages, please provide your GitHub Personal Access Token (PAT) to create it."
    read -sp "GitHub Personal Access Token (PAT): " pat_token
    echo ""
    kubectl create secret docker-registry regcred \
      --docker-server=ghcr.io \
      --docker-username="${gh_username}" \
      --docker-password="${pat_token}" \
      -n raas
    echo "Secret 'regcred' created successfully!"
else
    echo "Using existing 'regcred' image pull secret."
fi

echo "Rewriting image registry source, pull policy, and imagePullSecrets in manifests..."
# Find all deployment YAMLs and replace "image: raas/" with "image: ghcr.io/username/"
# and "imagePullPolicy: Never" with "imagePullPolicy: Always"
find k8s/apps/ -type f -name "*.yaml" -exec sed -i "s|image: raas/|image: ghcr.io/${gh_username}/|g" {} +
find k8s/apps/ -type f -name "*.yaml" -exec sed -i "s|imagePullPolicy: Never|imagePullPolicy: Always|g" {} +

# Inject imagePullSecrets block right before containers: if not already injected
find k8s/apps/ -type f -name "*.yaml" -exec grep -q "regcred" {} \; -o -exec sed -i "s|      containers:|      imagePullSecrets:\n        - name: regcred\n      containers:|g" {} \;


echo "Applying media environment secret..."
if [ -f "media/.env" ]; then
    kubectl create secret generic media-env --from-file=.env=media/.env -n raas --dry-run=client -o yaml | kubectl apply -f -
else
    echo "Warning: media/.env not found! Creating media-env secret from media/.env.example template instead."
    kubectl create secret generic media-env --from-file=.env=media/.env.example -n raas --dry-run=client -o yaml | kubectl apply -f -
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

echo "Applying Gateway ConfigMap, Deployment, and Ingress..."
kubectl apply -f k8s/apps/gateway/

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
echo "Access the app at: http://<your-vm-public-ip>"
echo "============================================="
