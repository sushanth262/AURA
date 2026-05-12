<#
.SYNOPSIS
  Deploy aura-worker to Azure App Service (Linux container, FREE SKU by default).

.DESCRIPTION
  Creates or updates a Web App that pulls the GHCR image for aura-worker.
    Sets WEBSITES_PORT=8083 and WORKER_ENABLED_SOURCES.

    Deploy order for linked stack: (1) worker — this script — then supervisor, authz, BFF.

    Use -PackageVersion for the GHCR image tag (same tag Terraform pushed); -ImageTag is an alias.

  -Local
    Applies aura-deployment/k8s/docker-desktop/worker.yaml to the current kubectl context (Docker Desktop Kubernetes).
    Exposes NodePort 30883 → http://127.0.0.1:30883 . Uses namespace aura-{UniqueSuffix}.

.EXAMPLE
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-worker.ps1 -UniqueSuffix 'jdoe01' -ResourceGroupName 'aura-rg' -Location 'eastus' -PackageVersion 'abc1234'

.EXAMPLE
  kubectl config use-context docker-desktop
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-worker.ps1 -UniqueSuffix 'jdoe01' -PackageVersion 'latest' -Local
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$UniqueSuffix,
    [string]$ResourceGroupName = 'aura',
    [string]$Location = 'centralus',
    [string]$SharedPlanName = 'aura-free-plan',
    [string]$AppServiceSku = 'B1',
    [string]$GhcrOwner = 'sushanth262',

    # GHCR image tag (@PackageVersion); overrides -ImageTag if both set to same semantics.
    [string]$PackageVersion,
    [string]$ImageTag,
    [string]$GhcrUser,
    [string]$GhcrToken,
    [string]$SubscriptionId,
    [switch]$SkipHealthWait,
    [string]$SummaryOutPath,

    [switch]$Local,
    [switch]$SkipGhcrImagePullSecret
)

$ErrorActionPreference = 'Stop'
. "$PSScriptRoot/deploy-azure-backend-common.ps1"

$tag = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag $ImageTag
$image = "ghcr.io/$GhcrOwner/aura-worker`:$tag"

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
    Deploy-AuraK8sServiceFromDockerDesktopTemplate -TemplateFileName 'worker.yaml' -Replacements @{
        '{{NAMESPACE}}'          = $ns
        '{{WORKER_IMAGE}}'       = $image
        '{{IMAGE_PULL_SECRETS}}' = $pullBlock
    }
    Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-worker'
    $hostRoot = 'http://127.0.0.1:30883'
    Write-Host ''
    Write-Host '=== aura-worker deployed (Kubernetes) ===' -ForegroundColor Green
    Write-Host "  Namespace    : $ns"
    Write-Host "  Image        : $image"
    Write-Host "  Host URL     : $hostRoot"
    Write-Host ''
    if (-not $SkipHealthWait) {
        Wait-AuraHealthEndpoint -HttpsRoot $hostRoot -Label 'aura-worker'
    }
    if ($SummaryOutPath) {
        @{ worker_https = $hostRoot; kubernetes_namespace = $ns } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
    }
    return
}

Assert-AzLoggedIn
if ($SubscriptionId) {
    az account set --subscription $SubscriptionId --only-show-errors
    if ($LASTEXITCODE -ne 0) { throw 'az account set failed' }
}

$creds = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken
$names = Get-AuraStandardAppNames -UniqueSuffix $UniqueSuffix

$extra = @{
    WORKER_ENABLED_SOURCES = 'grafana,github,jira'
}

$fqdn = Deploy-AuraLinuxContainerWebApp `
    -ResourceGroup $ResourceGroupName `
    -Location $Location `
    -PlanName $SharedPlanName `
    -AppName $names.Worker `
    -ContainerImage $image `
    -GhcrUser $creds.User `
    -GhcrToken $creds.Token `
    -WebSitesPort 8083 `
    -ExtraAppSettings $extra `
    -Sku $AppServiceSku

$httpsRoot = "https://$fqdn"
Write-Host ""
Write-Host "=== aura-worker deployed ===" -ForegroundColor Green
Write-Host "  PackageVersion (image tag): $tag"
Write-Host "  Web App name : $($names.Worker)"
Write-Host "  Public URL   : $httpsRoot"
Write-Host "  Health       : $httpsRoot/healthz"
Write-Host "  Sample mock  : $httpsRoot/v1/sources/grafana?scenario_key=inc2847_api_gateway"
Write-Host ""

if (-not $SkipHealthWait) {
    Wait-AuraHealthEndpoint -HttpsRoot $httpsRoot -Label 'aura-worker'
}

if ($SummaryOutPath) {
    @{ worker_https = $httpsRoot; worker_app = $names.Worker } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
}
