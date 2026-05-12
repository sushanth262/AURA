<#
.SYNOPSIS
  Deploy aura-frontend to Azure Static Web Apps or Docker Desktop Kubernetes (-Local).

.DESCRIPTION
  Azure (default):
    Pulls ghcr.io/{GhcrOwner}/aura-frontend:{PackageVersion}, extracts the static bundle,
    and publishes to the Azure Static Web App defined by -StaticSiteResourceId.
    Uses -UseAzureCliForToken to fetch the SWA deployment token automatically, or reads
    AZURE_STATIC_WEB_APPS_API_TOKEN from the environment.

  -Local:
    Deploys the nginx-based aura-frontend container to Docker Desktop Kubernetes namespace aura-{UniqueSuffix}.
    Serves on NodePort 30900 -> http://127.0.0.1:30900

  The SPA must be built with EXPO_PUBLIC_API_BASE_URL / EXPO_PUBLIC_WS_BASE_URL aimed at your BFF.
  Rebuild/push via Terraform if the BFF URL changes.

.EXAMPLE
  # Azure Static Web App (auto-fetch token via az CLI)
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-frontend.ps1 -UniqueSuffix 'jdoe01' -PackageVersion 'local-ui-fix-7' -UseAzureCliForToken

.EXAMPLE
  # Docker Desktop Kubernetes
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-frontend.ps1 -UniqueSuffix 'jdoe01' -PackageVersion 'latest' -Local
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$UniqueSuffix,

    [string]$GhcrOwner = 'sushanth262',
    [string]$PackageVersion,
    [string]$ImageTag,
    [string]$GhcrUser,
    [string]$GhcrToken,

    [switch]$SkipHealthWait,
    [string]$SummaryOutPath,

    # Azure SWA parameters
    [string]$StaticSiteResourceId = '/subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourceGroups/aura/providers/Microsoft.Web/staticSites/Aura',
    [switch]$UseAzureCliForToken,
    [string]$BackendStackSummaryJson,
    [switch]$StrictBffBundleMatch,
    [switch]$SkipBffConnectivityCheck,

    [switch]$Local,
    [switch]$SkipGhcrImagePullSecret
)

$ErrorActionPreference = 'Stop'
. "$PSScriptRoot/deploy-azure-backend-common.ps1"

$tag = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag $ImageTag
$image = "ghcr.io/$GhcrOwner/aura-frontend`:$tag"

# ── Local: Kubernetes (Docker Desktop) ─────────────────────────────────────
if ($Local) {
    Assert-KubectlCli
    Warn-KubectlDockerDesktopContext
    $ns = Get-AuraKubernetesNamespaceFromSuffix -UniqueSuffix $UniqueSuffix
    Ensure-AuraKubernetesNamespace -Namespace $ns

    if (-not $SkipGhcrImagePullSecret) {
        $credsL = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken
        Ensure-AuraK8sGhcrPullSecret -Namespace $ns -DockerUsername $credsL.User -DockerPassword $credsL.Token
    }
    $pullBlock = Resolve-AuraK8sImagePullSecretsYamlBlock -SkipGhcrImagePullSecret:$SkipGhcrImagePullSecret

    Deploy-AuraK8sServiceFromDockerDesktopTemplate -TemplateFileName 'frontend.yaml' -Replacements @{
        '{{NAMESPACE}}'          = $ns
        '{{FRONTEND_IMAGE}}'     = $image
        '{{IMAGE_PULL_SECRETS}}' = $pullBlock
    }

    Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-frontend'

    $uiRoot = 'http://127.0.0.1:30900'
    Write-Host ''
    Write-Host '=== aura-frontend deployed (Kubernetes) ===' -ForegroundColor Green
    Write-Host "  Namespace : $ns"
    Write-Host "  Image     : $image"
    Write-Host "  Open UI   : $uiRoot"
    Write-Host ''

    if (-not $SkipHealthWait) {
        Wait-AuraHealthEndpoint -HttpsRoot $uiRoot -Label 'aura-frontend'
    }

    if ($SummaryOutPath) {
        @{ frontend_https = $uiRoot; kubernetes_namespace = $ns } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
    }
    return
}

# ── Azure: Static Web App publish ──────────────────────────────────────────
Write-Host ''
Write-Host '=== Publishing aura-frontend to Azure Static Web App ===' -ForegroundColor Cyan
Write-Host "  Container image : $image"
Write-Host "  SWA resource    : $StaticSiteResourceId"
Write-Host ''

$swaParams = @{
    ContainerImage         = $image
    StaticSiteResourceId   = $StaticSiteResourceId
}
if ($UseAzureCliForToken)    { $swaParams.UseAzureCliForToken    = $true }
if ($StrictBffBundleMatch)   { $swaParams.StrictBffBundleMatch   = $true }
if ($SkipBffConnectivityCheck) { $swaParams.SkipBffConnectivityCheck = $true }
if (-not [string]::IsNullOrWhiteSpace($BackendStackSummaryJson)) {
    $swaParams.BackendStackSummaryJson = $BackendStackSummaryJson
}

Publish-AuraAzureStaticWebApp @swaParams

$ctx = Resolve-AuraStaticSiteFromResourceId -ResourceId $StaticSiteResourceId
$siteName = $ctx.StaticSiteName
$swaHostname = az staticwebapp show --name $siteName --resource-group $ctx.ResourceGroupName --query defaultHostname -o tsv --only-show-errors 2>$null
if ([string]::IsNullOrWhiteSpace($swaHostname)) {
    $swaHostname = 'calm-stone-06e76f010.7.azurestaticapps.net'
}

$uiRoot = "https://$($swaHostname.Trim())"
Write-Host ''
Write-Host '=== aura-frontend deployed (Azure Static Web App) ===' -ForegroundColor Green
Write-Host "  Image   : $image"
Write-Host "  Open UI : $uiRoot"
Write-Host ''

if ($SummaryOutPath) {
    @{ frontend_https = $uiRoot; static_site_name = $siteName } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
}
