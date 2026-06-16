param (
    [Parameter(Mandatory=$true)]
    [string]$GithubUsername
)

$repoRoot = Resolve-Path "$PSScriptRoot\.."
Push-Location $repoRoot

# GHCR requires lowercase usernames
$username = $GithubUsername.ToLower()

$services = @{
    "booking"      = "booking-service"
    "listing"      = "listing-service"
    "media"        = "media-service"
    "favorites"    = "favorites-service"
    "payment"      = "payment-service"
    "review"       = "review-service"
    "analytics"    = "analytics-service"
    "notification" = "notification-service"
    "user"         = "user-service"
    "ui"           = "ui-service"
    "gateway"      = "gateway"
}

Write-Host "============================================="
Write-Host "   Building & Pushing RaaS Images to GHCR    "
Write-Host "============================================="
Write-Host "Target username: $username"
Write-Host "Please ensure you have authenticated via 'docker login ghcr.io' first."
Write-Host ""

foreach ($dir in $services.Keys) {
    $imageName = $services[$dir]
    $tag = "ghcr.io/$username/${imageName}:latest"
    
    Write-Host "---------------------------------------------"
    Write-Host "Building image: $imageName"
    Write-Host "Tag: $tag"
    Write-Host "---------------------------------------------"
    
    # Build
    docker build -t $tag ./$dir
    
    # Push
    Write-Host "Pushing $tag to GHCR..."
    docker push $tag
}

Write-Host ""
Write-Host "============================================="
Write-Host "All images built and pushed successfully!"
Write-Host "============================================="

Pop-Location

