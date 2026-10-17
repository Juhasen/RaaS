#!/bin/bash
set -e

# Get repo root relative to scripts/debian directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
REPO_ROOT="$(dirname "$(dirname "$DIR")")"
cd "$REPO_ROOT"

# Usage: deploy-ghcr.sh [tag]
#   tag  image tag to deploy for every service, e.g. sha-abc1234 (default: latest).
#        CI publishes a complete sha-<short> set for every commit on main.
# Non-interactive. If the 'regcred' pull secret does not exist yet, set GHCR_PAT
# to a GitHub token with read:packages so the script can create it.
TAG="${1:-latest}"

# All images live in the project owner's GHCR namespace.
gh_username="juhasen"

echo "Deploying ghcr.io/${gh_username}/<service>:${TAG}"

if kubectl get namespace >/dev/null 2>&1; then
    KUBECTL_CMD=(kubectl)
elif command -v sudo >/dev/null 2>&1; then
    if sudo -n kubectl get namespace >/dev/null 2>&1; then
        KUBECTL_CMD=(sudo -n kubectl)
    else
        echo "kubectl is not accessible to the current user and passwordless sudo is not available."
        exit 1
    fi
else
    echo "kubectl is not usable and sudo is not available."
    exit 1
fi

kc() {
    "${KUBECTL_CMD[@]}" "$@"
}

echo "Applying namespace..."
kc apply -f k8s/infra/namespace.yaml

# Image pull secret for GHCR: create it once from $GHCR_PAT
if ! kc get secret regcred -n raas >/dev/null 2>&1; then
    if [ -z "${GHCR_PAT:-}" ]; then
        echo "Image pull secret 'regcred' not found in namespace 'raas'. Re-run with GHCR_PAT set to a GitHub token with read:packages."
        exit 1
    fi
    kc create secret docker-registry regcred       --docker-server=ghcr.io       --docker-username="${gh_username}"       --docker-password="${GHCR_PAT}"       -n raas
    echo "Secret 'regcred' created."
else
    echo "Using existing 'regcred' image pull secret."
fi

echo "Applying media environment secret..."
if [ -f "media/.env" ]; then
    kc create secret generic media-env --from-file=.env=media/.env -n raas --dry-run=client -o yaml | kc apply -f -
else
    echo "Warning: media/.env not found! Creating media-env secret from media/.env.example template instead."
    kc create secret generic media-env --from-file=.env=media/.env.example -n raas --dry-run=client -o yaml | kc apply -f -
fi


echo "Applying payment environment secret..."
if [ -f "payment/.env" ]; then
    kc create secret generic payment-env --from-env-file=payment/.env -n raas --dry-run=client -o yaml | kc apply -f -
else
    echo "Warning: payment/.env not found! Creating payment-env secret from payment/.env.example template instead."
    kc create secret generic payment-env --from-env-file=payment/.env.example -n raas --dry-run=client -o yaml | kc apply -f -
fi

echo "Applying notification environment secret..."
if [ -f "notification/.env" ]; then
    kc create secret generic notification-env --from-env-file=notification/.env -n raas --dry-run=client -o yaml | kc apply -f -
else
    echo "Warning: notification/.env not found! Creating notification-env secret from notification/.env.example template instead."
    kc create secret generic notification-env --from-env-file=notification/.env.example -n raas --dry-run=client -o yaml | kc apply -f -
fi


echo "Applying infrastructure configmap..."
kc apply -f k8s/infra/configmap.yaml

echo "Applying databases and message brokers..."
kc apply -f k8s/infra/postgres.yaml
kc apply -f k8s/infra/mongodb.yaml
kc apply -f k8s/infra/redis.yaml
kc apply -f k8s/infra/kafka.yaml

echo "Waiting for infrastructure to become ready..."
kc wait --for=condition=ready pod -l app=postgres -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=mongodb -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=redis -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=kafka -n raas --timeout=120s

echo "Initializing Postgres databases for microservices..."
pgPod=$(kc get pods -n raas -l app=postgres -o jsonpath="{.items[0].metadata.name}")
dbs=("payment" "review" "favorites" "user_db" "notification" "analytics")
for db in "${dbs[@]}"; do
    kc exec -n raas "$pgPod" -- psql -U raas_user -d raas_db -c "CREATE DATABASE $db;" 2>/dev/null || true
done

echo "Applying applications (Kustomize ghcr overlay, tag ${TAG})..."
# Pin every image to $TAG with a temporary overlay on top of the ghcr overlay.
RENDER_DIR=$(mktemp -d)
trap 'rm -rf "$RENDER_DIR"' EXIT
cp -R k8s "$RENDER_DIR/k8s"
{
    echo "apiVersion: kustomize.config.k8s.io/v1beta1"
    echo "kind: Kustomization"
    echo "resources:"
    echo "  - k8s/overlays/ghcr"
    echo "images:"
    for svc in booking listing media favorites payment review analytics notification user ui gateway; do
        echo "  - name: ghcr.io/${gh_username}/${svc}"
        echo "    newTag: ${TAG}"
    done
} > "$RENDER_DIR/kustomization.yaml"
kc kustomize "$RENDER_DIR" | kc apply -f -

deployments=(
    "booking-deployment"
    "listing-deployment"
    "media-deployment"
    "favorites-deployment"
    "payment-deployment"
    "review-deployment"
    "analytics-deployment"
    "notification-deployment"
    "user-deployment"
    "ui-deployment"
    "gateway"
)
# A new tag changes the pod spec and rolls out by itself; 'latest' needs a restart to re-pull.
if [ "$TAG" = "latest" ]; then
    echo "Tag is 'latest': forcing rollout restart so the newest images are pulled..."
    for deployment in "${deployments[@]}"; do
        kc rollout restart deployment/$deployment -n raas
    done
fi

for deployment in "${deployments[@]}"; do
    kc rollout status deployment/$deployment -n raas --timeout=120s
done

echo "Waiting for all applications to become ready..."
kc wait --for=condition=ready pod -l app=listing-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=media-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=booking-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=payment-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=review-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=favorites-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=user-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=notification-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=analytics-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=ui-service -n raas --timeout=120s
kc wait --for=condition=ready pod -l app=gateway -n raas --timeout=120s

echo "All services are running!"
echo "Access the app at: http://<your-vm-public-ip>"
echo "============================================="
