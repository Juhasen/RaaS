# Deploys already-published GHCR images to the Azure VM over SSH.
# CI (build-push.yml) builds and pushes the images; this script only rolls them out.
#
# Example: .\scripts\deploy-azure.ps1 -VmHost 1.2.3.4 -VmUser azureuser
# Without -Tag the VM deploys its freshly pulled commit: each service gets the tag CI built for it.
# Set GHCR_PAT on the first deploy if the VM has no 'regcred' pull secret yet.
param (
    [Parameter(Mandatory = $true)]
    [string]$VmHost,

    [Parameter(Mandatory = $true)]
    [string]$VmUser,

    # Optional: one image tag for every service (e.g. latest). Default: per-service sha tags.
    [string]$Tag = "",

    [string]$RemoteRepoPath = "~/RaaS",

    [int]$SshPort = 22,

    # Private key for the VM; defaults to $env:RAAS_SSH_KEY, or ssh's own key lookup.
    [string]$IdentityFile = $env:RAAS_SSH_KEY
)

$ErrorActionPreference = 'Stop'

if ($Tag -and $Tag -notmatch '^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$') {
    throw "Invalid image tag: $Tag"
}

if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) {
    throw "ssh was not found in PATH. Install OpenSSH client before using this script."
}

if ($IdentityFile -and -not (Test-Path -LiteralPath $IdentityFile)) {
    throw "SSH identity file not found: $IdentityFile"
}

if ($RemoteRepoPath -like '~/*') {
    $RemoteRepoPath = "/home/$VmUser/" + $RemoteRepoPath.Substring(2)
}

Write-Host "Deploying (tag: $(if ($Tag) { $Tag } else { 'per-service sha' })) to ${VmUser}@${VmHost}:$SshPort ($RemoteRepoPath)"

$envPrefix = if ($env:GHCR_PAT) { "GHCR_PAT='$($env:GHCR_PAT)' " } else { "" }
$remoteCommand = "cd '$RemoteRepoPath' && git pull --ff-only && ${envPrefix}bash scripts/debian/deploy-ghcr.sh $(if ($Tag) { "'$Tag'" })"

$sshArgs = @('-p', $SshPort)
if ($IdentityFile) {
    $sshArgs += @('-i', $IdentityFile)
}
$sshArgs += @("$VmUser@$VmHost", $remoteCommand)

& ssh @sshArgs
if ($LASTEXITCODE -ne 0) {
    throw "Remote deploy on Azure VM failed."
}

Write-Host "Deployment completed successfully."
