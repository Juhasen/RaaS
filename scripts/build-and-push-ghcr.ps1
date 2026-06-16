param (
    [Parameter(Mandatory=$true)]
    [string]$GithubUsername
)

$ErrorActionPreference = 'Stop'

$repoRoot = Resolve-Path "$PSScriptRoot\.."
Push-Location $repoRoot

# GHCR requires lowercase usernames
$username = $GithubUsername.ToLower()

function Test-ServiceChanges {
    param (
        [Parameter(Mandatory = $true)]
        [string]$ServicePath
    )

    $gitStatus = & git status --porcelain -- $ServicePath
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to inspect git changes for $ServicePath."
    }

    return -not [string]::IsNullOrWhiteSpace($gitStatus)
}

$services = [ordered]@{
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

    if (-not (Test-ServiceChanges -ServicePath $dir)) {
        Write-Host "---------------------------------------------"
        Write-Host "Skipping image: $imageName"
        Write-Host "Reason: no git changes detected in ./$dir"
        Write-Host "---------------------------------------------"
        continue
    }
    
    Write-Host "---------------------------------------------"
    Write-Host "Building image: $imageName"
    Write-Host "Tag: $tag"
    Write-Host "---------------------------------------------"
    
    # Build
    & docker build -t $tag ./$dir
    if ($LASTEXITCODE -ne 0) {
        throw "Docker build failed for $imageName."
    }
    
    # Push
    Write-Host "Pushing $tag to GHCR..."
    & docker push $tag
    if ($LASTEXITCODE -ne 0) {
        throw "GHCR denied the push for $tag. Confirm you are logged in to ghcr.io with a token that has write:packages permission, and that the package namespace matches the account that owns the image."
    }
}

Write-Host ""
Write-Host "============================================="
Write-Host "GHCR build/push complete!"
Write-Host "============================================="

Pop-Location

