<#
.SYNOPSIS
  Deploy aura-authz (authorization evaluate API).

.DESCRIPTION
  No Azure peer dependency. Deploy after worker+supervisor if you follow the full stack order,
  or any time before BFF.

    Use -PackageVersion for the GHCR tag; -ImageTag is an alias.

.EXAMPLE
  .\deploy-azure-authz.ps1 -UniqueSuffix 'jdoe01' -ResourceGroupName 'aura-rg' -PackageVersion 'abc1234'

.EXAMPLE
  .\deploy-azure-authz.ps1 -UniqueSuffix 'jdoe01' -PackageVersion 'latest' -Local
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$UniqueSuffix,
    [string]$ResourceGroupName = 'aura-rg',
    [string]$Location = 'eastus',
    [string]$SharedPlanName = 'aura-free-plan',
    [string]$AppServiceSku = 'FREE',
    [string]$GhcrOwner = 'sushanth262',
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

$names = Get-AuraStandardAppNames -UniqueSuffix $UniqueSuffix
$tag = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag $ImageTag
$image = "ghcr.io/$GhcrOwner/aura-authz`:$tag"

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
    Deploy-AuraK8sServiceFromDockerDesktopTemplate -TemplateFileName 'authz.yaml' -Replacements @{
        '{{NAMESPACE}}'          = $ns
        '{{AUTHZ_IMAGE}}'        = $image
        '{{IMAGE_PULL_SECRETS}}' = $pullBlock
    }
    Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-authz'
    $httpsRoot = 'http://127.0.0.1:30881'
    Write-Host ''
    Write-Host '=== aura-authz deployed (Kubernetes) ===' -ForegroundColor Green
    Write-Host "  Namespace : $ns"
    Write-Host "  Host URL  : $httpsRoot"
    Write-Host ''
    if (-not $SkipHealthWait) {
        Wait-AuraHealthEndpoint -HttpsRoot $httpsRoot -Label 'aura-authz'
    }
    if ($SummaryOutPath) {
        @{ authz_https = $httpsRoot; kubernetes_namespace = $ns } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
    }
    return
}

Assert-AzLoggedIn
if ($SubscriptionId) {
    az account set --subscription $SubscriptionId --only-show-errors
    if ($LASTEXITCODE -ne 0) { throw 'az account set failed' }
}

$creds = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken

$extra = @{
    POLICY_VERSION = 'stub-v1'
}

$fqdn = Deploy-AuraLinuxContainerWebApp `
    -ResourceGroup $ResourceGroupName `
    -Location $Location `
    -PlanName $SharedPlanName `
    -AppName $names.Authz `
    -ContainerImage $image `
    -GhcrUser $creds.User `
    -GhcrToken $creds.Token `
    -WebSitesPort 8081 `
    -ExtraAppSettings $extra `
    -Sku $AppServiceSku

$httpsRoot = "https://$fqdn"
Write-Host ""
Write-Host "=== aura-authz deployed ===" -ForegroundColor Green
Write-Host "  PackageVersion (image tag): $tag"
Write-Host "  Web App name : $($names.Authz)"
Write-Host "  Evaluate API : $httpsRoot/v1/evaluate"
Write-Host ""

if (-not $SkipHealthWait) {
    Wait-AuraHealthEndpoint -HttpsRoot $httpsRoot -Label 'aura-authz'
}

if ($SummaryOutPath) {
    @{ authz_https = $httpsRoot; authz_app = $names.Authz } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
}
