#!/bin/bash

echo "Starting local Kubernetes environment for RaaS..."

echo "Applying namespace..."
kubectl apply -f k8s/infra/namespace.yaml

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

echo "Applying Go applications..."
kubectl apply -f k8s/apps/go/

echo "Applying Java applications..."
kubectl apply -f k8s/apps/java/

echo "Applying Python applications..."
kubectl apply -f k8s/apps/python/

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

echo "All services are running!"
