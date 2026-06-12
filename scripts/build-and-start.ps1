$repoRoot = Resolve-Path "$PSScriptRoot\.."
Push-Location $repoRoot

Write-Host "Building Docker images for RaaS microservices..."

# Go Services
Write-Host "Building Go Service: booking..."
docker build -t raas/booking-service:latest ./booking

Write-Host "Building Go Service: listing..."
docker build -t raas/listing-service:latest ./listing

Write-Host "Building Go Service: media..."
docker build -t raas/media-service:latest ./media

# Java Services
Write-Host "Building Java Service: favorites..."
docker build -t raas/favorites-service:latest ./favorites

Write-Host "Building Java Service: payment..."
docker build -t raas/payment-service:latest ./payment

Write-Host "Building Java Service: review..."
docker build -t raas/review-service:latest ./review

# Python Services
Write-Host "Building Python Service: analytics..."
docker build -t raas/analytics-service:latest ./analytics

Write-Host "Building Python Service: notification..."
docker build -t raas/notification-service:latest ./notification

Write-Host "Building Python Service: user..."
docker build -t raas/user-service:latest ./user

# Gateway
Write-Host "Building Gateway Service: gateway..."
docker build -t raas/gateway:latest ./gateway

Write-Host "Images built successfully!"

Write-Host "Starting Kubernetes deployments..."

# Wait a second before launching the other script
Start-Sleep -Seconds 2

# Call the local k8s startup script
.\scripts\start-local-k8s.ps1
