<#
.SYNOPSIS
  Full pipeline: deploy four backends (README order) + write combined release manifest; optionally publish Static Web App like frontend-only flow.

.DESCRIPTION
  1. deploy-azure-worker.ps1 (PackageVersion = GHCR tag)
  2. deploy-azure-supervisor.ps1
  3. deploy-azure-authz.ps1
  4. deploy-azure-bff-api.ps1
  5. Writes aura-release-{suffix}-{package}.json (Azure: az webapp show; -Local: localhost NodePorts + frontend_https).
  With -Local (default): deploys aura-frontend after BFF (NodePort 30900). Use -SkipLocalFrontend for APIs only.

  Prerequisites:
    - Azure: az login; terraform apply pushed GHCR images at PackageVersion
    - -Local: kubectl (Docker Desktop Kubernetes); same GHCR images & pull secret (or -SkipGhcrImagePullSecret for public)
    - For -PublishStaticWebApp: aura-frontend:{PackageVersion} on GHCR; bundle EXPO_PUBLIC_* must match BFF
    - Not supported: -Local with -PublishStaticWebApp (use Azure SWA path only)

.EXAMPLE
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-full-release.ps1 `
    -PackageVersion $(git rev-parse --short HEAD) `
    -UniqueSuffix 'jdoe01' `
    -ResourceGroupName 'aura-rg' `
    -CorsAllowedOrigins 'https://YOUR_STATIC_SITE.azurestaticapps.net,http://localhost:19006'

.EXAMPLE
  Same plus frontend publish (pull aura-frontend:{PackageVersion} -> SWA):
  .\deploy-azure-full-release.ps1 -PackageVersion 'abc1234' -UniqueSuffix 'jdoe01' `
    -PublishStaticWebApp -UseAzureCliForToken -StrictBffBundleMatch

.EXAMPLE
  Docker Desktop Kubernetes (backends + UI). -CorsAllowedOrigins is optional; UI origin http://127.0.0.1:30900 is added when not already listed.
  .\deploy-azure-full-release.ps1 -PackageVersion 'latest' -UniqueSuffix 'jdoe01' -Local `
    -CorsAllowedOrigins 'http://127.0.0.1:30900,http://localhost:19006'
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$PackageVersion,
    [Parameter(Mandatory)][string]$UniqueSuffix,

    [string]$ResourceGroupName = 'aura-rg',
    [string]$Location = 'eastus',
    [string]$SharedPlanName = 'aura-free-plan',
    [string]$AppServiceSku = 'FREE',
    [string]$GhcrOwner = 'sushanth262',
    [string]$GhcrUser,
    [string]$GhcrToken,
    [string]$SubscriptionId,

    [string]$AuthDevJwtSecret = 'aura-dev-secret-change-me-use-32chars-minimum!!',
    [string]$InternalSharedSecret = 'aura-internal-demo',
    [string]$CorsAllowedOrigins = 'http://localhost:8081,http://localhost:19006',

    [switch]$SkipDependencyChecks,
    [switch]$SkipHealthWaitBackend,

    [string]$ReleaseManifestPath,

    [switch]$PublishStaticWebApp,
    [switch]$UseAzureCliForToken,
    [string]$ContainerImage,
    [string]$StaticSiteResourceId,

    [switch]$StrictBffBundleMatch,
    [switch]$SkipBffConnectivityCheck,

    [switch]$Local,
    [switch]$SkipGhcrImagePullSecret,

    # With -Local: deploy aura-frontend (nginx on NodePort 30900) after BFF. Omit -SkipLocalFrontend to skip UI only.
    [switch]$SkipLocalFrontend
)

$ErrorActionPreference = 'Stop'
$here = $PSScriptRoot
. "$here/deploy-azure-backend-common.ps1"

if ($Local -and $PublishStaticWebApp) {
    throw '-PublishStaticWebApp is only supported for Azure. Omit -Local or skip Static Web Apps publish.'
}

if ($Local) {
    Assert-KubectlCli
    Warn-KubectlDockerDesktopContext
}
else {
    Assert-AzLoggedIn
    if ($SubscriptionId) {
        az account set --subscription $SubscriptionId --only-show-errors | Out-Null
        if ($LASTEXITCODE -ne 0) { throw 'az account set failed' }
    }
}

$stem = Get-AuraSafeReleaseFileStem -PackageVersion $PackageVersion
if ([string]::IsNullOrWhiteSpace($ReleaseManifestPath)) {
    $ReleaseManifestPath = Join-Path $PWD ('aura-release-{0}-{1}.json' -f $UniqueSuffix, $stem)
}

$common = @{
    UniqueSuffix             = $UniqueSuffix
    ResourceGroupName        = $ResourceGroupName
    Location                 = $Location
    SharedPlanName           = $SharedPlanName
    AppServiceSku            = $AppServiceSku
    GhcrOwner                = $GhcrOwner
    PackageVersion           = $PackageVersion
    GhcrUser                 = $GhcrUser
    GhcrToken                = $GhcrToken
    SubscriptionId           = $SubscriptionId
    Local                    = $Local
    SkipGhcrImagePullSecret = $SkipGhcrImagePullSecret
}

$corsForBff = $CorsAllowedOrigins
if ($Local -and -not $SkipLocalFrontend) {
    if ($corsForBff -notmatch '30900') {
        $uiOrigin = 'http://127.0.0.1:30900'
        if ([string]::IsNullOrWhiteSpace($corsForBff)) {
            $corsForBff = $uiOrigin
        }
        else {
            $corsForBff = ($corsForBff.Trim().TrimEnd(',')) + ',' + $uiOrigin
        }
    }
}

Write-Host ''
Write-Host '=== Aura full release: backends in README dependency order ===' -ForegroundColor Cyan
Write-Host "    PackageVersion (GHCR tag): $PackageVersion"
if ($Local) {
    Write-Host '    Target: Docker Desktop Kubernetes (-Local)' -ForegroundColor Yellow
    if (-not $SkipLocalFrontend) { Write-Host '    Includes aura-frontend -> http://127.0.0.1:30900' -ForegroundColor Yellow }
}
Write-Host ''

Write-Host '--- 1/4 worker ---' -ForegroundColor Cyan
& "$here/deploy-azure-worker.ps1" @common -SkipHealthWait:$SkipHealthWaitBackend

Write-Host '--- 2/4 supervisor ---' -ForegroundColor Cyan
& "$here/deploy-azure-supervisor.ps1" @common `
    -AuthDevJwtSecret $AuthDevJwtSecret `
    -InternalSharedSecret $InternalSharedSecret `
    -SkipDependencyChecks:$SkipDependencyChecks `
    -SkipHealthWait:$SkipHealthWaitBackend

Write-Host '--- 3/4 authz ---' -ForegroundColor Cyan
& "$here/deploy-azure-authz.ps1" @common -SkipHealthWait:$SkipHealthWaitBackend

Write-Host '--- 4/4 bff-api ---' -ForegroundColor Cyan
& "$here/deploy-azure-bff-api.ps1" @common `
    -AuthDevJwtSecret $AuthDevJwtSecret `
    -InternalSharedSecret $InternalSharedSecret `
    -CorsAllowedOrigins $corsForBff `
    -SkipDependencyChecks:$SkipDependencyChecks `
    -SkipHealthWait:$SkipHealthWaitBackend

if ($Local -and -not $SkipLocalFrontend) {
    Write-Host '--- aura-frontend (local UI) ---' -ForegroundColor Cyan
    $fe = @{
        UniqueSuffix             = $UniqueSuffix
        GhcrOwner                = $GhcrOwner
        PackageVersion           = $PackageVersion
        GhcrUser                 = $GhcrUser
        GhcrToken                = $GhcrToken
        SkipGhcrImagePullSecret = $SkipGhcrImagePullSecret
        Local                    = $true
    }
    & "$here/deploy-azure-frontend.ps1" @fe -SkipHealthWait:$SkipHealthWaitBackend
}

Write-Host '--- combined manifest ---' -ForegroundColor Cyan
if ($Local) {
    $k8sns = Get-AuraKubernetesNamespaceFromSuffix -UniqueSuffix $UniqueSuffix
    Export-AuraLocalKubernetesEndpointsManifest `
        -KubernetesNamespace $k8sns `
        -UniqueSuffix $UniqueSuffix `
        -OutputPath $ReleaseManifestPath `
        -PackageVersion $PackageVersion `
        -GhcrOwner $GhcrOwner
}
else {
    Export-AuraAzureBackendEndpointsManifest `
        -UniqueSuffix $UniqueSuffix `
        -ResourceGroupName $ResourceGroupName `
        -OutputPath $ReleaseManifestPath `
        -SubscriptionId $SubscriptionId `
        -PackageVersion $PackageVersion `
        -GhcrOwner $GhcrOwner
}

Write-Host ''
Write-Host "Release manifest: $ReleaseManifestPath" -ForegroundColor Green

if ($PublishStaticWebApp) {
    $img = $ContainerImage
    if ([string]::IsNullOrWhiteSpace($img)) {
        $tagResolved = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag ''
        $img = "ghcr.io/$GhcrOwner/aura-frontend`:$tagResolved"
    }

    Write-Host ''
    Write-Host "Publishing Static Web App from $img (frontend-only flow)..." -ForegroundColor Cyan

    $swa = @{
        ContainerImage          = $img
        BackendStackSummaryJson = $ReleaseManifestPath
    }
    if ($StrictBffBundleMatch) { $swa.StrictBffBundleMatch = $true }
    if ($SkipBffConnectivityCheck) { $swa.SkipBffConnectivityCheck = $true }
    if ($UseAzureCliForToken) { $swa.UseAzureCliForToken = $true }
    if (-not [string]::IsNullOrWhiteSpace($StaticSiteResourceId)) {
        $swa.StaticSiteResourceId = $StaticSiteResourceId
    }

    Publish-AuraAzureStaticWebApp @swa
}
