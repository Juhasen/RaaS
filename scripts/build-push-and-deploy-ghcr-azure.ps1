param (
    [string]$GithubUsername = "juhasen",

    [Parameter(Mandatory = $true)]
    [string]$VmHost,

    [Parameter(Mandatory = $true)]
    [string]$VmUser,

    [string]$RemoteRepoPath = "~/RaaS",

    [int]$SshPort = 22,

    [string]$IdentityFile = "C:\Users\Krystian\dev\RaaS_key.pem"
)

$ErrorActionPreference = 'Stop'

$repoRoot = Resolve-Path "$PSScriptRoot\.."
Push-Location $repoRoot

try {
    Write-Host "============================================="
    Write-Host "   Building & Pushing RaaS Images to GHCR    "
    Write-Host "============================================="
    Write-Host "Target username: $($GithubUsername.ToLower())"
    Write-Host ""

    & "$repoRoot\scripts\build-and-push-ghcr.ps1" -GithubUsername $GithubUsername

    Write-Host ""
    Write-Host "============================================="
    Write-Host "   Deploying GHCR images on Azure VM        "
    Write-Host "============================================="
    Write-Host "VM: ${VmUser}@${VmHost}:$SshPort"
    Write-Host "Remote repo path: $RemoteRepoPath"
    Write-Host ""

    if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) {
        throw "ssh was not found in PATH. Install OpenSSH client before using this script."
    }

    if ($IdentityFile -and -not (Test-Path -LiteralPath $IdentityFile)) {
        throw "SSH identity file not found: $IdentityFile"
    }

    if ($RemoteRepoPath -like '~/*') {
        $RemoteRepoPath = "/home/$VmUser/" + $RemoteRepoPath.Substring(2)
    }

    $remoteCommand = "cd '$RemoteRepoPath' && git pull && bash scripts/debian/deploy-ghcr.sh '$($GithubUsername.ToLower())'"
    $sshArgs = @('-tt', '-p', $SshPort)
    if ($IdentityFile) {
        $sshArgs += @('-i', $IdentityFile)
    }
    $sshArgs += @("$VmUser@$VmHost", $remoteCommand)

    & ssh @sshArgs
    if ($LASTEXITCODE -ne 0) {
        throw "Remote deploy on Azure VM failed."
    }

    Write-Host ""
    Write-Host "Deployment completed successfully."
}
finally {
    Pop-Location
}