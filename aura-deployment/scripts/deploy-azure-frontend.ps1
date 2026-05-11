<#
.SYNOPSIS
  Deploy aura-frontend (nginx static bundle) to Docker Desktop Kubernetes (-Local only in this repo).

.DESCRIPTION
  Pulls ghcr.io/{GhcrOwner}/aura-frontend:{PackageVersion} into namespace aura-{UniqueSuffix}.
  Serves on NodePort 30900 -> http://127.0.0.1:30900

  The SPA must be built with EXPO_PUBLIC_API_BASE_URL / EXPO_PUBLIC_WS_BASE_URL aimed at your BFF
  (local K8s: http://127.0.0.1:30880/v1 and ws://127.0.0.1:30880). Rebuild/push via Terraform if needed.

.EXAMPLE
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

    [switch]$Local,
    [switch]$SkipGhcrImagePullSecret
)

$ErrorActionPreference = 'Stop'
. "$PSScriptRoot/deploy-azure-backend-common.ps1"

if (-not $Local) {
    throw 'This script currently supports only -Local (Kubernetes). Use deploy-azure-full-release.ps1 -PublishStaticWebApp for Azure Static Web Apps.'
}

$tag = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag $ImageTag
$image = "ghcr.io/$GhcrOwner/aura-frontend`:$tag"

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
