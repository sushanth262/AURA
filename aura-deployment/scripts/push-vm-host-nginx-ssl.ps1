#Requires -Version 5.1
<#
.SYNOPSIS
  Upload aura-deployment/nginx-vm-proxy-ssl.conf to the Azure VM and reload nginx.

.DESCRIPTION
  Tries host nginx first (/usr/sbin/nginx). If not present, finds a Docker container that
  publishes port 443, docker cp the config in, and nginx -s reload inside the container.

  Default path inside the VM/container is /etc/nginx/conf.d/aura-rca-cloudapp.conf.

.PARAMETER OutFile
  Full path for the vhost file on the VM or inside the nginx container.
#>
param(
  [string]$ResourceGroup = "auravm",
  [string]$VmName = "aura",
  [string]$OutFile = "/etc/nginx/conf.d/aura-rca-cloudapp.conf"
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

# Verbatim here-string: double single quotes '' emit one ' in the bash script.
$bash = @'
set -euo pipefail
sudo mkdir -p /etc/nginx/conf.d
OUT="__OUTFILE__"
TMP=/tmp/aura-rca-cloudapp.conf
echo "__CONFB64__" | base64 -d > "$TMP"
chmod 644 "$TMP"

try_host() {
  if [ ! -x /usr/sbin/nginx ] && ! command -v nginx >/dev/null 2>&1; then
    return 1
  fi
  sudo cp "$TMP" "$OUT"
  sudo chmod 644 "$OUT"
  if [ -x /usr/sbin/nginx ]; then
    sudo /usr/sbin/nginx -t
    sudo /usr/sbin/nginx -s reload 2>/dev/null || sudo systemctl reload nginx
  else
    sudo nginx -t
    sudo nginx -s reload 2>/dev/null || sudo systemctl reload nginx
  fi
  rm -f "$TMP"
  echo NGINX_OK_HOST
  return 0
}

if try_host; then
  exit 0
fi

NGINX_C=""
for n in $(docker ps --format ''{{.Names}}'' 2>/dev/null); do
  if docker port "$n" 2>/dev/null | grep -qE '(:443->|0\.0\.0\.0:443|\[::\]:443)'; then
    NGINX_C="$n"
    break
  fi
done
if [ -z "$NGINX_C" ]; then
  echo "No host nginx and no Docker container publishing 443 found." >&2
  rm -f "$TMP"
  exit 1
fi

docker cp "$TMP" "${NGINX_C}:$OUT"
rm -f "$TMP"
docker exec "$NGINX_C" nginx -t
docker exec "$NGINX_C" nginx -s reload
echo "NGINX_OK_DOCKER_${NGINX_C}"
'@

$bash = $bash.Replace("__OUTFILE__", $OutFile)
$bash = $bash.Replace("__CONFB64__", $confB64)
$bash = $bash -replace "`r`n", "`n" -replace "`r", "`n"
$remoteB64 = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($bash))

Write-Host "=== push SSL nginx to $VmName ($ResourceGroup) ===" -ForegroundColor Cyan
Write-Host "    local:  $confPath"
Write-Host "    remote: $OutFile"

az vm run-command invoke `
  --resource-group $ResourceGroup `
  --name $VmName `
  --command-id RunShellScript `
  --scripts "echo $remoteB64 | base64 -d | bash" `
  --only-show-errors -o json

if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "=== done ===" -ForegroundColor Green
