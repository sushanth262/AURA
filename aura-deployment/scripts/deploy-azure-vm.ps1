<#
.SYNOPSIS
  Deploy all Aura services to an Azure Ubuntu VM via Docker Compose.

.DESCRIPTION
  1. Looks up the VM's public IP (or creates one if missing).
  2. Opens NSG ports 80 (frontend) and 8080 (BFF API).
  3. SSHs into the VM to install Docker (if needed) and login to GHCR.
  4. Uploads and runs a Docker Compose file that pulls pre-built GHCR images.
  5. Frontend on port 80, BFF API on port 8080.

  The VM must already exist with SSH key auth (azureuser).
  Works with the free-tier Standard_B2ats_v2 VM.

  Prerequisites:
    - az login
    - VM created with SSH key auth (default azureuser)
    - GHCR images pushed at PackageVersion
    - $env:GITHUB_TOKEN set (or pass -GhcrToken)

.EXAMPLE
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-vm.ps1 -PackageVersion 'local-ui-fix-8' -UniqueSuffix 'jdoe01'

.EXAMPLE
  .\deploy-azure-vm.ps1 -PackageVersion 'latest' -UniqueSuffix 'jdoe01' -VmPublicIp '20.1.2.3'
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$PackageVersion,
    [Parameter(Mandatory)][string]$UniqueSuffix,

    [string]$ResourceGroupName = 'auravm',
    [string]$VmName = 'aura',
    [string]$SshUser = 'azureuser',
    [string]$VmPublicIp,

    [string]$GhcrOwner = 'sushanth262',
    [string]$GhcrUser,
    [string]$GhcrToken,
    [string]$SubscriptionId,

    [string]$AuthDevJwtSecret = 'aura-dev-secret-change-me-use-32chars-minimum!!',
    [string]$InternalSharedSecret = 'aura-internal-demo',
    [string]$CorsAllowedOrigins,

    [switch]$SkipNsgRules,
    [switch]$SkipDockerInstall,
    [switch]$SkipFrontend,

    [string]$ReleaseManifestPath
)

$ErrorActionPreference = 'Stop'
$here = $PSScriptRoot
. "$here/deploy-azure-backend-common.ps1"

Assert-AzLoggedIn

if ($SubscriptionId) {
    az account set --subscription $SubscriptionId --only-show-errors | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'az account set failed' }
}

$tag = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag $null
$stem = Get-AuraSafeReleaseFileStem -PackageVersion $PackageVersion
$suffix = $UniqueSuffix.ToLower().Trim() -replace '[^a-z0-9\-]', ''
$creds = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken

if ([string]::IsNullOrWhiteSpace($ReleaseManifestPath)) {
    $ReleaseManifestPath = Join-Path $PWD ('aura-release-{0}-{1}-vm.json' -f $suffix, $stem)
}

$images = @{
    Worker     = "ghcr.io/$GhcrOwner/aura-worker`:$tag"
    Supervisor = "ghcr.io/$GhcrOwner/aura-supervisor`:$tag"
    Authz      = "ghcr.io/$GhcrOwner/aura-authz`:$tag"
    Bff        = "ghcr.io/$GhcrOwner/aura-bff-api`:$tag"
    Frontend   = "ghcr.io/$GhcrOwner/aura-frontend`:$tag"
}

Write-Host ''
Write-Host '=== Aura VM deployment ===' -ForegroundColor Cyan
Write-Host "    PackageVersion (GHCR tag): $tag"
Write-Host "    VM: $VmName (resource group: $ResourceGroupName)"
Write-Host ''

# ---- Step 1: Get VM public IP ----
Write-Host '--- 1/5 resolving VM public IP ---' -ForegroundColor Cyan

if ([string]::IsNullOrWhiteSpace($VmPublicIp)) {
    $ipResult = Invoke-AzWithRetry -Label "resolve VM public IP ($VmName)" -ScriptBlock {
        az vm show --name $VmName --resource-group $ResourceGroupName --show-details --query "publicIps" -o tsv --only-show-errors
    }
    $VmPublicIp = ($ipResult | Where-Object { $_ -match '^\d+\.\d+\.\d+\.\d+$' } | Select-Object -First 1)
    if ([string]::IsNullOrWhiteSpace($VmPublicIp)) {
        throw "Could not resolve public IP for VM '$VmName'. Pass -VmPublicIp explicitly or assign a public IP to the VM."
    }
}
Write-Host "  VM public IP: $VmPublicIp" -ForegroundColor Green

# ---- Step 2: Open NSG ports ----
Write-Host '--- 2/5 NSG rules ---' -ForegroundColor Cyan

if (-not $SkipNsgRules) {
    $nicId = Invoke-AzWithRetry -Label 'get VM NIC ID' -ScriptBlock {
        az vm show --name $VmName --resource-group $ResourceGroupName --query "networkProfile.networkInterfaces[0].id" -o tsv --only-show-errors
    }
    $nicId = ($nicId | Where-Object { $_ -match '/networkInterfaces/' } | Select-Object -First 1)

    $nsgId = Invoke-AzWithRetry -Label 'get NSG from NIC' -ScriptBlock {
        az network nic show --ids $nicId --query "networkSecurityGroup.id" -o tsv --only-show-errors
    }
    $nsgId = ($nsgId | Where-Object { $_ -match '/networkSecurityGroups/' } | Select-Object -First 1)

    if ([string]::IsNullOrWhiteSpace($nsgId)) {
        $subnetId = Invoke-AzWithRetry -Label 'get subnet from NIC' -ScriptBlock {
            az network nic show --ids $nicId --query "ipConfigurations[0].subnet.id" -o tsv --only-show-errors
        }
        $subnetId = ($subnetId | Where-Object { $_ -match '/subnets/' } | Select-Object -First 1)
        if (-not [string]::IsNullOrWhiteSpace($subnetId)) {
            $oldPref = $ErrorActionPreference; $ErrorActionPreference = 'SilentlyContinue'
            $nsgId = az network vnet subnet show --ids $subnetId --query "networkSecurityGroup.id" -o tsv --only-show-errors 2>&1
            $ErrorActionPreference = $oldPref
            $nsgId = ($nsgId | Where-Object { $_ -match '/networkSecurityGroups/' } | Select-Object -First 1)
        }
    }

    if ([string]::IsNullOrWhiteSpace($nsgId)) {
        Write-Host "  No NSG found attached to VM NIC or subnet. Skipping port rules." -ForegroundColor Yellow
    }
    else {
        $nsgName = ($nsgId -split '/')[-1]
        $nsgRg = ($nsgId -split '/')[4]

        $rules = @(
            @{ Name = 'AllowHTTP';    Port = '80';   Priority = 1010 }
            @{ Name = 'AllowBFF';     Port = '8080'; Priority = 1020 }
        )

        foreach ($rule in $rules) {
            $oldPref = $ErrorActionPreference
            $ErrorActionPreference = 'SilentlyContinue'
            az network nsg rule show --nsg-name $nsgName --resource-group $nsgRg --name $rule.Name -o none --only-show-errors 2>&1 | Out-Null
            $ruleExists = ($LASTEXITCODE -eq 0)
            $ErrorActionPreference = $oldPref

            if (-not $ruleExists) {
                Write-Host "  Creating NSG rule: $($rule.Name) (port $($rule.Port))..."
                az network nsg rule create `
                    --nsg-name $nsgName `
                    --resource-group $nsgRg `
                    --name $rule.Name `
                    --priority $rule.Priority `
                    --direction Inbound `
                    --access Allow `
                    --protocol Tcp `
                    --destination-port-ranges $rule.Port `
                    --only-show-errors -o none
            }
            else {
                Write-Host "  NSG rule '$($rule.Name)' already exists." -ForegroundColor Green
            }
        }
    }
}
else {
    Write-Host '  Skipped (-SkipNsgRules).'
}

# ---- Step 3: Prepare docker-compose.yml ----
Write-Host '--- 3/5 preparing docker-compose ---' -ForegroundColor Cyan

$corsOrigins = $CorsAllowedOrigins
if ([string]::IsNullOrWhiteSpace($corsOrigins)) {
    $vmDnsLabel = 'aura-rca'
    $vmRegion   = 'northcentralus'
    $vmFqdn     = "$vmDnsLabel.$vmRegion.cloudapp.azure.com"
    $corsOrigins = "http://$VmPublicIp,http://$vmFqdn,https://$vmFqdn,https://calm-stone-06e76f010.7.azurestaticapps.net,http://localhost:8081,http://localhost:19006"
}

$composeSrc = Join-Path (Join-Path $here '..') 'docker-compose.vm.yml'
$composeSrc = (Resolve-Path -LiteralPath $composeSrc -ErrorAction Stop).Path
$composeRaw = Get-Content -LiteralPath $composeSrc -Raw -Encoding UTF8

$composeYaml = $composeRaw `
    -replace '{{WORKER_IMAGE}}', $images.Worker `
    -replace '{{SUPERVISOR_IMAGE}}', $images.Supervisor `
    -replace '{{AUTHZ_IMAGE}}', $images.Authz `
    -replace '{{BFF_IMAGE}}', $images.Bff `
    -replace '{{FRONTEND_IMAGE}}', $images.Frontend `
    -replace '{{AUTH_DEV_JWT_SECRET}}', $AuthDevJwtSecret `
    -replace '{{INTERNAL_SHARED_SECRET}}', $InternalSharedSecret `
    -replace '{{CORS_ALLOWED_ORIGINS}}', $corsOrigins

$tmpCompose = Join-Path $env:TEMP 'aura-docker-compose-vm.yml'
Set-Content -LiteralPath $tmpCompose -Value $composeYaml -Encoding UTF8
Write-Host "  Compose file prepared ($tmpCompose)." -ForegroundColor Green

# ---- Step 4: SSH setup + deploy ----
Write-Host '--- 4/5 deploying to VM ---' -ForegroundColor Cyan

# Use 'az ssh vm' for AAD-based SSH (no local key needed for Azure-generated keys).
$azSshCommon = @('--resource-group', $ResourceGroupName, '--vm-name', $VmName, '--only-show-errors')

Write-Host "  Uploading docker-compose.yml to VM..."
# az ssh vm doesn't have scp built in; use run-command to write the file.
$composeB64 = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($composeYaml))
$uploadScript = "echo '$composeB64' | base64 -d > /home/$SshUser/docker-compose.yml && chown ${SshUser}:${SshUser} /home/$SshUser/docker-compose.yml && echo 'docker-compose.yml uploaded'"
Invoke-AzWithRetry -Label 'upload docker-compose.yml' -ScriptBlock {
    az vm run-command invoke @azSshCommon --command-id RunShellScript --scripts $uploadScript -o tsv --query "value[0].message"
}
Write-Host '  docker-compose.yml uploaded.' -ForegroundColor Green

if (-not $SkipDockerInstall) {
    Write-Host '  Installing Docker on VM (if needed, this may take 1-2 minutes)...'
    $installScript = @"
set -e
if ! command -v docker >/dev/null 2>&1; then
  echo '>>> Installing Docker...'
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc
  ARCH=`$(dpkg --print-architecture)
  CODENAME=`$(. /etc/os-release; echo `$VERSION_CODENAME)
  echo "deb [arch=`$ARCH signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu `$CODENAME stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin
  usermod -aG docker $SshUser
  echo '>>> Docker installed.'
else
  echo '>>> Docker already installed.'
fi
docker --version
docker compose version
"@
    Invoke-AzWithRetry -Label 'install Docker' -MaxRetries 3 -DelaySec 20 -ScriptBlock {
        az vm run-command invoke @azSshCommon --command-id RunShellScript --scripts $installScript -o tsv --query "value[0].message"
    }
    Write-Host '  Docker ready.' -ForegroundColor Green
}

Write-Host '  Logging into GHCR on VM...'
$ghcrLoginScript = "echo '$($creds.Token)' | docker login ghcr.io -u '$($creds.User)' --password-stdin"
Invoke-AzWithRetry -Label 'GHCR login' -ScriptBlock {
    az vm run-command invoke @azSshCommon --command-id RunShellScript --scripts $ghcrLoginScript -o tsv --query "value[0].message"
}
Write-Host '  GHCR login OK.' -ForegroundColor Green

Write-Host '  Pulling images and starting services (this may take a few minutes)...'
$deployScript = "set -e; cd /home/$SshUser; docker compose pull; docker compose up -d --remove-orphans; sleep 10; docker compose ps"
Invoke-AzWithRetry -Label 'docker compose up' -MaxRetries 3 -DelaySec 20 -ScriptBlock {
    az vm run-command invoke @azSshCommon --command-id RunShellScript --scripts $deployScript -o tsv --query "value[0].message"
}

Write-Host '  Services deployed.' -ForegroundColor Green

# ---- Step 5: Release manifest ----
Write-Host '--- 5/5 release manifest ---' -ForegroundColor Cyan

$manifest = @{
    deployment_target = 'vm'
    vm_name           = $VmName
    resource_group    = $ResourceGroupName
    public_ip         = $VmPublicIp
    image_tag         = $tag
    ghcr_owner        = $GhcrOwner
    endpoints         = @{
        bff_public = "http://${VmPublicIp}:8080"
        bff_api    = "http://${VmPublicIp}:8080/v1"
        bff_ws     = "ws://${VmPublicIp}:8080"
        frontend   = "http://$VmPublicIp"
    }
    services = @{
        worker     = @{ image = $images.Worker;     port = 'internal' }
        supervisor = @{ image = $images.Supervisor;  port = 'internal' }
        authz      = @{ image = $images.Authz;       port = 'internal' }
        bff_api    = @{ image = $images.Bff;          port = '8080' }
        frontend   = @{ image = $images.Frontend;     port = '80' }
    }
    timestamp = (Get-Date -Format 'o')
}

$manifest | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $ReleaseManifestPath -Encoding UTF8

Write-Host ''
Write-Host '========================================' -ForegroundColor Green
Write-Host ' Aura VM deployment complete!' -ForegroundColor Green
Write-Host '========================================' -ForegroundColor Green
Write-Host ''
Write-Host "  VM          : $VmName ($VmPublicIp)" -ForegroundColor White
Write-Host "  Frontend    : http://$VmPublicIp" -ForegroundColor Cyan
Write-Host "  BFF API     : http://${VmPublicIp}:8080/v1" -ForegroundColor Cyan
Write-Host "  BFF WS      : ws://${VmPublicIp}:8080" -ForegroundColor Cyan
Write-Host ''
Write-Host "  Release manifest: $ReleaseManifestPath" -ForegroundColor White
Write-Host ''
Write-Host '  SSH commands:' -ForegroundColor DarkGray
Write-Host "    ssh $sshTarget" -ForegroundColor DarkGray
Write-Host "    ssh $sshTarget docker compose ps" -ForegroundColor DarkGray
Write-Host "    ssh $sshTarget docker compose logs -f aura-bff-api" -ForegroundColor DarkGray
Write-Host ''
