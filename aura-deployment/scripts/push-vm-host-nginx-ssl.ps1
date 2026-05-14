#Requires -Version 5.1
<#
.SYNOPSIS
  Upload aura-deployment/nginx-vm-proxy-ssl.conf to the Azure VM and reload host nginx.

.DESCRIPTION
  Base64-encodes the repo SSL vhost file, sends it via az vm run-command, writes to
  sites-available, symlinks into sites-enabled, runs nginx -t, reloads nginx.

  If nginx -t fails (duplicate listen/server_name), remove or rename the conflicting file
  under /etc/nginx/sites-enabled/ on the VM, then re-run.

.PARAMETER SitesAvailablePath
  Full path on the VM for this vhost file.

.PARAMETER SitesEnabledName
  Symlink filename under /etc/nginx/sites-enabled/ (not a full path).
#>
param(
  [string]$ResourceGroup = "auravm",
  [string]$VmName = "aura",
  [string]$SitesAvailablePath = "/etc/nginx/sites-available/aura-rca-cloudapp",
  [string]$SitesEnabledName = "aura-rca-cloudapp"
)

$ErrorActionPreference = "Stop"
if (-not (Get-Command az -ErrorAction SilentlyContinue)) {
  Write-Error "Azure CLI (az) not found."
}
az account show -o none 2>$null | Out-Null
if ($LASTEXITCODE -ne 0) {
  Write-Error "Not logged in to Azure CLI. Run: az login"
}

$auraDeployment = Resolve-Path (Join-Path $PSScriptRoot "..")
$confPath = Join-Path $auraDeployment "nginx-vm-proxy-ssl.conf"
if (-not (Test-Path $confPath)) {
  Write-Error "Config not found: $confPath"
}

$text = [System.IO.File]::ReadAllText($confPath)
$text = $text -replace "`r`n", "`n" -replace "`r", "`n"
$confB64 = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($text))

$enabledFull = "/etc/nginx/sites-enabled/$SitesEnabledName"

$remote = @"
set -euo pipefail
echo '$confB64' | base64 -d | sudo tee $SitesAvailablePath > /dev/null
sudo chmod 644 $SitesAvailablePath
sudo ln -sf $SitesAvailablePath $enabledFull
sudo nginx -t
sudo systemctl reload nginx
echo NGINX_OK
"@
$remote = $remote -replace "`r`n", "`n" -replace "`r", "`n"
$remoteB64 = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($remote))

Write-Host "=== push SSL nginx to $VmName ($ResourceGroup) ===" -ForegroundColor Cyan
Write-Host "    local:  $confPath"
Write-Host "    remote: $SitesAvailablePath -> $enabledFull"

az vm run-command invoke `
  --resource-group $ResourceGroup `
  --name $VmName `
  --command-id RunShellScript `
  --scripts "echo $remoteB64 | base64 -d | bash" `
  --only-show-errors -o json

if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "=== done ===" -ForegroundColor Green
