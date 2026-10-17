param (
    [Parameter(Mandatory = $true)]
    [ValidateSet("booking", "listing", "media", "favorites", "payment", "review", "analytics", "notification", "user", "ui", "gateway")]
    [string]$Service
)

$ErrorActionPreference = 'Stop'

$repoRoot = Resolve-Path "$PSScriptRoot\.."
Push-Location $repoRoot

try {
    # Image is raas/<dir>; deployment is <dir>-deployment, except gateway.
    $image = "raas/${Service}:latest"
    $deployment = if ($Service -eq "gateway") { "gateway" } else { "${Service}-deployment" }

    docker build -t $image "./$Service"
    if ($LASTEXITCODE -ne 0) { throw "Docker build failed for $Service." }

    kubectl rollout restart deployment $deployment -n raas
    if ($LASTEXITCODE -ne 0) { throw "Rollout restart failed for $deployment." }

    kubectl rollout status deployment $deployment -n raas
    if ($LASTEXITCODE -ne 0) { throw "Rollout of $deployment did not become ready." }
}
finally {
    Pop-Location
}
