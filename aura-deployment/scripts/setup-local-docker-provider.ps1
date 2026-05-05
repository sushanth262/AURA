# Stages kreuzwerker/docker for "terraform init -plugin-dir" (avoids Windows filesystem_mirror
# bug where only the .zip is installed and Terraform tries to fork/exec the zip).
param(
    [string]$SourceExe = "$env:USERPROFILE\OneDrive\Documents\terraform-provider-docker_3.9.0_windows_amd64\terraform-provider-docker_v3.9.0.exe",
    [string]$ProviderVersion = "3.9.0"
)

$ErrorActionPreference = "Stop"
$deploymentRoot = Split-Path $PSScriptRoot -Parent
$destDir = Join-Path $deploymentRoot ".terraform-plugins\registry.terraform.io\kreuzwerker\docker\$ProviderVersion\windows_amd64"
$destExe = Join-Path $destDir "terraform-provider-docker_v$ProviderVersion.exe"

if (-not (Test-Path -LiteralPath $SourceExe)) {
    Write-Error "Provider exe not found: $SourceExe`nDownload/unpack release v$ProviderVersion from https://github.com/kreuzwerker/terraform-provider-docker/releases"
}

New-Item -ItemType Directory -Force -Path $destDir | Out-Null
Copy-Item -LiteralPath $SourceExe -Destination $destExe -Force
Write-Host "Staged: $destExe"
Write-Host "From aura-deployment, run:"
Write-Host '  terraform init -plugin-dir=".terraform-plugins"'
