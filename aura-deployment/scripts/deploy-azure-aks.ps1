<#
.SYNOPSIS
  Create a Free-tier AKS cluster and deploy all Aura services.

.DESCRIPTION
  1. Creates (or reuses) a Free-tier AKS cluster with 1 Standard_B2s node.
  2. Deploys all five Aura microservices (worker, supervisor, authz, bff-api, frontend).
  3. BFF and frontend are exposed via LoadBalancer services with public IPs.
  4. Backend services (worker, supervisor, authz) stay internal (ClusterIP).
  5. Writes a release manifest JSON with the public endpoints.

  Prerequisites:
    - az login
    - GHCR images pushed at PackageVersion (e.g. via terraform apply)
    - $env:GITHUB_TOKEN set (or pass -GhcrToken)

.EXAMPLE
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-aks.ps1 -PackageVersion 'local-ui-fix-8' -UniqueSuffix 'jdoe01'

.EXAMPLE
  .\deploy-azure-aks.ps1 -PackageVersion 'latest' -UniqueSuffix 'jdoe01' -Location 'eastus'
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$PackageVersion,
    [Parameter(Mandatory)][string]$UniqueSuffix,

    [string]$ResourceGroupName = 'aura',
    [string]$Location = 'centralus',
    [string]$ClusterName = 'aura-aks',
    [int]$NodeCount = 1,
    [string]$NodeVmSize = 'Standard_D2s_v3',

    [string]$GhcrOwner = 'sushanth262',
    [string]$GhcrUser,
    [string]$GhcrToken,
    [string]$SubscriptionId,

    [string]$AuthDevJwtSecret = 'aura-dev-secret-change-me-use-32chars-minimum!!',
    [string]$InternalSharedSecret = 'aura-internal-demo',
    [string]$CorsAllowedOrigins,

    [switch]$SkipGhcrImagePullSecret,
    [switch]$SkipFrontend,
    [switch]$SkipClusterCreate,

    [string]$ReleaseManifestPath
)

$ErrorActionPreference = 'Stop'
$here = $PSScriptRoot
. "$here/deploy-azure-backend-common.ps1"

Assert-AzLoggedIn
Assert-KubectlCli

if ($SubscriptionId) {
    az account set --subscription $SubscriptionId --only-show-errors | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'az account set failed' }
}

$ns = Get-AuraKubernetesNamespaceFromSuffix -UniqueSuffix $UniqueSuffix
$tag = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag $null
$stem = Get-AuraSafeReleaseFileStem -PackageVersion $PackageVersion

if ([string]::IsNullOrWhiteSpace($ReleaseManifestPath)) {
    $ReleaseManifestPath = Join-Path $PWD ('aura-release-{0}-{1}-aks.json' -f $UniqueSuffix, $stem)
}

$images = @{
    Worker     = "ghcr.io/$GhcrOwner/aura-worker`:$tag"
    Supervisor = "ghcr.io/$GhcrOwner/aura-supervisor`:$tag"
    Authz      = "ghcr.io/$GhcrOwner/aura-authz`:$tag"
    Bff        = "ghcr.io/$GhcrOwner/aura-bff-api`:$tag"
    Frontend   = "ghcr.io/$GhcrOwner/aura-frontend`:$tag"
}

Write-Host ''
Write-Host '=== Aura AKS deployment ===' -ForegroundColor Cyan
Write-Host "    PackageVersion (GHCR tag): $tag"
Write-Host "    AKS cluster: $ClusterName ($Location, Free tier, ${NodeCount}x $NodeVmSize)"
Write-Host "    Namespace: $ns"
Write-Host ''

# ---- Step 0: Ensure Microsoft.ContainerService provider is registered ----
$regState = az provider show --namespace Microsoft.ContainerService --query "registrationState" -o tsv --only-show-errors 2>$null
if ($regState -ne 'Registered') {
    Write-Host 'Registering Microsoft.ContainerService resource provider...' -ForegroundColor Yellow
    az provider register --namespace Microsoft.ContainerService --only-show-errors 2>$null
    $waited = 0
    while ($waited -lt 180) {
        Start-Sleep -Seconds 10
        $waited += 10
        $regState = az provider show --namespace Microsoft.ContainerService --query "registrationState" -o tsv --only-show-errors 2>$null
        if ($regState -eq 'Registered') { break }
        Write-Host "  Waiting for registration... ($regState, ${waited}s)" -ForegroundColor Yellow
    }
    if ($regState -ne 'Registered') { throw 'Microsoft.ContainerService provider failed to register within 3 minutes.' }
    Write-Host 'Microsoft.ContainerService registered.' -ForegroundColor Green
}

# ---- Step 1: Resource group ----
Write-Host '--- 1/7 resource group ---' -ForegroundColor Cyan
Ensure-ResourceGroup -Name $ResourceGroupName -Location $Location

# ---- Step 2: AKS cluster ----
Write-Host '--- 2/7 AKS cluster ---' -ForegroundColor Cyan
if (-not $SkipClusterCreate) {
    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = 'SilentlyContinue'
    $clusterJson = az aks show --name $ClusterName --resource-group $ResourceGroupName -o json --only-show-errors 2>&1
    $clusterExists = ($LASTEXITCODE -eq 0)
    $ErrorActionPreference = $oldPref

    if ($clusterExists) {
        Write-Host "AKS cluster '$ClusterName' already exists. Reusing." -ForegroundColor Green
    }
    else {
        Write-Host "Creating Free-tier AKS cluster '$ClusterName' ($NodeVmSize x $NodeCount)..."
        Write-Host "  This may take 3-5 minutes..." -ForegroundColor Yellow
        az aks create `
            --resource-group $ResourceGroupName `
            --name $ClusterName `
            --location $Location `
            --node-count $NodeCount `
            --node-vm-size $NodeVmSize `
            --tier free `
            --generate-ssh-keys `
            --enable-managed-identity `
            --network-plugin azure `
            --only-show-errors `
            -o none
        if ($LASTEXITCODE -ne 0) { throw "az aks create failed for '$ClusterName'" }
        Write-Host "AKS cluster '$ClusterName' created." -ForegroundColor Green
    }
}
else {
    Write-Host 'Skipping cluster creation (-SkipClusterCreate).'
}

# ---- Step 3: Get AKS credentials ----
Write-Host '--- 3/7 kubectl credentials ---' -ForegroundColor Cyan
az aks get-credentials `
    --resource-group $ResourceGroupName `
    --name $ClusterName `
    --overwrite-existing `
    --only-show-errors | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'az aks get-credentials failed' }
Write-Host "kubectl context set to AKS cluster '$ClusterName'." -ForegroundColor Green

# ---- Step 4: Namespace + pull secret ----
Write-Host '--- 4/7 namespace + pull secret ---' -ForegroundColor Cyan
Ensure-AuraKubernetesNamespace -Namespace $ns
if (-not $SkipGhcrImagePullSecret) {
    $creds = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken
    Ensure-AuraK8sGhcrPullSecret -Namespace $ns -DockerUsername $creds.User -DockerPassword $creds.Token
}
$pullBlock = Resolve-AuraK8sImagePullSecretsYamlBlock -SkipGhcrImagePullSecret:$SkipGhcrImagePullSecret

# ---- Step 5: Deploy backend services ----
Write-Host '--- 5/7 deploying backend services ---' -ForegroundColor Cyan

$aksManifestDir = Join-Path (Join-Path (Join-Path $here '..') 'k8s') 'aks'
$aksManifestDir = (Resolve-Path -LiteralPath $aksManifestDir -ErrorAction Stop).Path

function Deploy-AksManifest {
    param(
        [Parameter(Mandatory)][string]$FileName,
        [Parameter(Mandatory)][hashtable]$Replacements
    )
    $path = Join-Path $aksManifestDir $FileName
    if (-not (Test-Path -LiteralPath $path)) { throw "AKS manifest missing: $path" }
    $raw = Get-Content -LiteralPath $path -Raw -Encoding UTF8
    $yaml = Expand-AuraK8sManifestPlaceholders -YamlContent $raw -Replacements $Replacements
    Invoke-AuraKubectlApplyYamlString -YamlContent $yaml
}

$commonReplacements = @{
    '{{NAMESPACE}}'          = $ns
    '{{IMAGE_PULL_SECRETS}}' = $pullBlock
}

Write-Host '  worker...'
Deploy-AksManifest -FileName 'worker.yaml' -Replacements ($commonReplacements + @{
    '{{WORKER_IMAGE}}' = $images.Worker
})
Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-worker'
Write-Host '  worker deployed.' -ForegroundColor Green

Write-Host '  supervisor...'
Deploy-AksManifest -FileName 'supervisor.yaml' -Replacements ($commonReplacements + @{
    '{{SUPERVISOR_IMAGE}}'      = $images.Supervisor
    '{{AUTH_DEV_JWT_SECRET}}'    = $AuthDevJwtSecret
    '{{INTERNAL_SHARED_SECRET}}' = $InternalSharedSecret
})
Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-supervisor'
Write-Host '  supervisor deployed.' -ForegroundColor Green

Write-Host '  authz...'
Deploy-AksManifest -FileName 'authz.yaml' -Replacements ($commonReplacements + @{
    '{{AUTHZ_IMAGE}}' = $images.Authz
})
Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-authz'
Write-Host '  authz deployed.' -ForegroundColor Green

# ---- Step 6: Deploy BFF + frontend (LoadBalancer) ----
Write-Host '--- 6/7 deploying BFF + frontend ---' -ForegroundColor Cyan

$corsOrigins = $CorsAllowedOrigins
if ([string]::IsNullOrWhiteSpace($corsOrigins)) {
    $corsOrigins = 'https://calm-stone-06e76f010.7.azurestaticapps.net,http://localhost:8081,http://localhost:19006'
}

Write-Host '  bff-api...'
Deploy-AksManifest -FileName 'bff.yaml' -Replacements ($commonReplacements + @{
    '{{BFF_IMAGE}}'              = $images.Bff
    '{{AUTH_DEV_JWT_SECRET}}'    = $AuthDevJwtSecret
    '{{AUTHZ_URL}}'              = 'http://aura-authz:8081'
    '{{SUPERVISOR_URL}}'         = 'http://aura-supervisor:8082'
    '{{INTERNAL_SHARED_SECRET}}' = $InternalSharedSecret
    '{{CORS_ALLOWED_ORIGINS}}'   = $corsOrigins
})
Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-bff-api'
Write-Host '  bff-api deployed.' -ForegroundColor Green

if (-not $SkipFrontend) {
    Write-Host '  frontend...'
    Deploy-AksManifest -FileName 'frontend.yaml' -Replacements ($commonReplacements + @{
        '{{FRONTEND_IMAGE}}' = $images.Frontend
    })
    Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-frontend'
    Write-Host '  frontend deployed.' -ForegroundColor Green
}

# ---- Step 7: Wait for public IPs ----
Write-Host '--- 7/7 waiting for public IPs ---' -ForegroundColor Cyan

function Wait-LoadBalancerIP {
    param(
        [Parameter(Mandatory)][string]$ServiceName,
        [Parameter(Mandatory)][string]$Namespace,
        [int]$TimeoutSec = 300
    )
    $start = Get-Date
    while ($true) {
        $ip = kubectl get svc $ServiceName -n $Namespace -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>$null
        if (-not [string]::IsNullOrWhiteSpace($ip) -and $ip -ne '<pending>') {
            return $ip
        }
        $elapsed = ((Get-Date) - $start).TotalSeconds
        if ($elapsed -gt $TimeoutSec) {
            throw "Timed out waiting for LoadBalancer IP on service '$ServiceName' after ${TimeoutSec}s"
        }
        Write-Host "  Waiting for $ServiceName external IP... (${([int]$elapsed)}s)" -ForegroundColor Yellow
        Start-Sleep -Seconds 10
    }
}

$bffIp = Wait-LoadBalancerIP -ServiceName 'aura-bff-api' -Namespace $ns
Write-Host "  BFF public IP: $bffIp" -ForegroundColor Green

$frontendIp = $null
if (-not $SkipFrontend) {
    $frontendIp = Wait-LoadBalancerIP -ServiceName 'aura-frontend' -Namespace $ns
    Write-Host "  Frontend public IP: $frontendIp" -ForegroundColor Green

    # Update CORS to include the frontend LB IP
    $feOrigin = "http://$frontendIp"
    if ($corsOrigins -notmatch [regex]::Escape($frontendIp)) {
        $updatedCors = $corsOrigins + ',' + $feOrigin
        Write-Host "  Updating BFF CORS to include frontend IP ($feOrigin)..."
        Deploy-AksManifest -FileName 'bff.yaml' -Replacements ($commonReplacements + @{
            '{{BFF_IMAGE}}'              = $images.Bff
            '{{AUTH_DEV_JWT_SECRET}}'    = $AuthDevJwtSecret
            '{{AUTHZ_URL}}'              = 'http://aura-authz:8081'
            '{{SUPERVISOR_URL}}'         = 'http://aura-supervisor:8082'
            '{{INTERNAL_SHARED_SECRET}}' = $InternalSharedSecret
            '{{CORS_ALLOWED_ORIGINS}}'   = $updatedCors
        })
        kubectl rollout restart deployment/aura-bff-api -n $ns | Out-Null
    }
}

# ---- Release manifest ----
$endpoints = @{
    worker_https     = "http://$bffIp"
    supervisor_https = "http://$bffIp"
    authz_https      = "http://$bffIp"
    bff_https        = "http://$bffIp"
}
if ($frontendIp) {
    $endpoints.frontend_https = "http://$frontendIp"
}

$manifest = @{
    deployment_target = 'aks'
    cluster_name      = $ClusterName
    namespace         = $ns
    image_tag         = $tag
    ghcr_owner        = $GhcrOwner
    endpoints         = @{
        bff_public      = "http://${bffIp}"
        bff_api         = "http://${bffIp}/v1"
        bff_ws          = "ws://${bffIp}"
    }
    services = @{
        worker     = @{ image = $images.Worker;     type = 'ClusterIP'; port = 8083 }
        supervisor = @{ image = $images.Supervisor;  type = 'ClusterIP'; port = 8082 }
        authz      = @{ image = $images.Authz;       type = 'ClusterIP'; port = 8081 }
        bff_api    = @{ image = $images.Bff;          type = 'LoadBalancer'; port = 80; external_ip = $bffIp }
    }
    timestamp = (Get-Date -Format 'o')
}
if ($frontendIp) {
    $manifest.endpoints.frontend = "http://$frontendIp"
    $manifest.services.frontend = @{ image = $images.Frontend; type = 'LoadBalancer'; port = 80; external_ip = $frontendIp }
}

$manifest | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $ReleaseManifestPath -Encoding UTF8

Write-Host ''
Write-Host '========================================' -ForegroundColor Green
Write-Host ' Aura AKS deployment complete!' -ForegroundColor Green
Write-Host '========================================' -ForegroundColor Green
Write-Host ''
Write-Host "  AKS cluster : $ClusterName ($Location, Free tier)" -ForegroundColor White
Write-Host "  Namespace   : $ns" -ForegroundColor White
Write-Host "  BFF API     : http://$bffIp/v1" -ForegroundColor Cyan
Write-Host "  BFF WS      : ws://$bffIp" -ForegroundColor Cyan
if ($frontendIp) {
    Write-Host "  Frontend    : http://$frontendIp" -ForegroundColor Cyan
}
Write-Host ''
Write-Host "  Release manifest: $ReleaseManifestPath" -ForegroundColor White
Write-Host ''
Write-Host '  kubectl commands:' -ForegroundColor DarkGray
Write-Host "    kubectl get pods -n $ns" -ForegroundColor DarkGray
Write-Host "    kubectl get svc -n $ns" -ForegroundColor DarkGray
Write-Host "    kubectl logs -n $ns deployment/aura-bff-api --tail=50" -ForegroundColor DarkGray
Write-Host ''
