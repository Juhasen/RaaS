#!/bin/bash
set -e

# Get repo root relative to scripts/debian directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
REPO_ROOT="$(dirname "$(dirname "$DIR")")"
cd "$REPO_ROOT"

echo "Building Docker images for RaaS microservices..."

# Go Services
echo "Building Go Service: booking..."
docker build -t raas/booking-service:latest ./booking

echo "Building Go Service: listing..."
docker build -t raas/listing-service:latest ./listing

echo "Building Go Service: media..."
docker build -t raas/media-service:latest ./media

# Java Services
echo "Building Java Service: favorites..."
docker build -t raas/favorites-service:latest ./favorites

echo "Building Java Service: payment..."
docker build -t raas/payment-service:latest ./payment

echo "Building Java Service: review..."
docker build -t raas/review-service:latest ./review

# Python Services
echo "Building Python Service: analytics..."
docker build -t raas/analytics-service:latest ./analytics

echo "Building Python Service: notification..."
docker build -t raas/notification-service:latest ./notification

echo "Building Python Service: user..."
docker build -t raas/user-service:latest ./user

# UI
echo "Building UI Service: ui..."
docker build -t raas/ui-service:latest ./ui

# Gateway
echo "Building Gateway Service: gateway..."
docker build -t raas/gateway:latest ./gateway

echo "Images built successfully!"

echo "Starting Kubernetes deployments..."
sleep 2

# Call the local k8s startup script
chmod +x scripts/debian/start-local-k8s.sh
./scripts/debian/start-local-k8s.sh
