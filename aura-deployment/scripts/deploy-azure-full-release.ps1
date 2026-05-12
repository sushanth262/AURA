<#
.SYNOPSIS
  Full pipeline: deploy four backends + frontend, then write a combined release manifest.

.DESCRIPTION
  Deploys in dependency order:
    1. deploy-azure-worker.ps1
    2. deploy-azure-supervisor.ps1
    3. deploy-azure-authz.ps1
    4. deploy-azure-bff-api.ps1
    5. deploy-azure-frontend.ps1 (Azure SWA or local Kubernetes)
    6. Writes aura-release-{suffix}-{package}.json

  Azure App Service (default):
    Backends deploy to Azure App Service (Linux containers on the shared Free plan).
    Frontend publishes to Azure Static Web App (calm-stone-06e76f010.7.azurestaticapps.net).
    Use -SkipFrontend to deploy backends only.

  -AKS:
    Creates (or reuses) a Free-tier AKS cluster and deploys all services to it.
    BFF and frontend get public LoadBalancer IPs. Internal services use ClusterIP.
    Delegates to deploy-azure-aks.ps1.

  -ContainerApps:
    Deploys all services to Azure Container Apps (free-tier Consumption plan).
    BFF and frontend get external HTTPS ingress. Internal services use internal ingress.
    Delegates to deploy-azure-containerapps.ps1.

  -VM:
    Deploys all services to an existing Azure VM via Docker Compose over SSH.
    Frontend on port 80, BFF on port 8080. Delegates to deploy-azure-vm.ps1.

  -Local:
    Backends + frontend deploy to Docker Desktop Kubernetes (NodePorts 30880-30883, 30900).

  Prerequisites:
    - Azure: az login; terraform apply pushed GHCR images at PackageVersion
    - -Local: kubectl (Docker Desktop Kubernetes); same GHCR images & pull secret

.EXAMPLE
  # Full Azure release (backends + SWA frontend, auto-fetch SWA token via az CLI)
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-full-release.ps1 -PackageVersion 'local-ui-fix-7' -UniqueSuffix 'jdoe01' -UseAzureCliForToken

.EXAMPLE
  # AKS deployment (Free tier cluster, all services)
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-full-release.ps1 -PackageVersion 'local-ui-fix-8' -UniqueSuffix 'jdoe01' -AKS

.EXAMPLE
  # Container Apps deployment (free tier, westus)
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-full-release.ps1 -PackageVersion 'local-ui-fix-8' -UniqueSuffix 'jdoe01' -ContainerApps

.EXAMPLE
  # Azure backends only, skip frontend
  .\deploy-azure-full-release.ps1 -PackageVersion 'abc1234' -UniqueSuffix 'jdoe01' -SkipFrontend

.EXAMPLE
  # Docker Desktop Kubernetes (backends + UI on NodePort 30900)
  .\deploy-azure-full-release.ps1 -PackageVersion 'latest' -UniqueSuffix 'jdoe01' -Local
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$PackageVersion,
    [Parameter(Mandatory)][string]$UniqueSuffix,

    [string]$ResourceGroupName = 'aura',
    [string]$Location = 'centralus',
    [string]$SharedPlanName = 'aura-free-plan',
    [string]$AppServiceSku = 'B1',
    [string]$GhcrOwner = 'sushanth262',
    [string]$GhcrUser,
    [string]$GhcrToken,
    [string]$SubscriptionId,

    [string]$AuthDevJwtSecret = 'aura-dev-secret-change-me-use-32chars-minimum!!',
    [string]$InternalSharedSecret = 'aura-internal-demo',
    [string]$CorsAllowedOrigins = 'https://calm-stone-06e76f010.7.azurestaticapps.net,http://localhost:8081,http://localhost:19006',

    [switch]$SkipDependencyChecks,
    [switch]$SkipHealthWaitBackend,

    [string]$ReleaseManifestPath,

    # Azure SWA parameters (ignored when -Local)
    [string]$StaticSiteResourceId = '/subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourceGroups/aura/providers/Microsoft.Web/staticSites/Aura',
    [switch]$UseAzureCliForToken,
    [switch]$StrictBffBundleMatch,
    [switch]$SkipBffConnectivityCheck,

    [switch]$Local,
    [switch]$SkipGhcrImagePullSecret,

    # Deploy to a Free-tier AKS cluster instead of App Service.
    [switch]$AKS,
    [string]$AksClusterName = 'aura-aks',
    [int]$AksNodeCount = 1,
    [string]$AksNodeVmSize = 'Standard_D2s_v3',
    [switch]$SkipAksClusterCreate,

    # Deploy to Azure Container Apps (free-tier Consumption plan).
    [switch]$ContainerApps,
    [string]$ContainerAppsEnvName = 'aura-cae',
    [switch]$SkipContainerAppsEnvCreate,

    # Deploy to an existing Azure VM via Docker Compose over SSH.
    [switch]$VM,
    [string]$VmResourceGroup = 'auravm',
    [string]$VmName = 'aura',
    [string]$VmSshUser = 'azureuser',
    [string]$VmPublicIp,
    [switch]$SkipVmNsgRules,
    [switch]$SkipVmDockerInstall,

    # Skip frontend deployment (Azure SWA or local K8s). Deploy backends only.
    [switch]$SkipFrontend,

    # Backward compat: -PublishStaticWebApp is now the default for Azure. Accepted but ignored.
    [switch]$PublishStaticWebApp,
    # Backward compat: renamed to -SkipFrontend.
    [switch]$SkipLocalFrontend
)

$ErrorActionPreference = 'Stop'
$here = $PSScriptRoot
. "$here/deploy-azure-backend-common.ps1"

if ($SkipLocalFrontend -and -not $SkipFrontend) {
    $SkipFrontend = $true
}

$targetSwitches = @($AKS, $ContainerApps, $VM, $Local) | Where-Object { $_ }
if ($targetSwitches.Count -gt 1) {
    throw 'Use only one of -AKS, -ContainerApps, -VM, or -Local. Choose one deployment target.'
}

# AKS path delegates entirely to deploy-azure-aks.ps1
if ($AKS) {
    Write-Host ''
    Write-Host '=== Aura full release (AKS) ===' -ForegroundColor Cyan
    $aksParams = @{
        PackageVersion       = $PackageVersion
        UniqueSuffix         = $UniqueSuffix
        ResourceGroupName    = $ResourceGroupName
        Location             = $Location
        ClusterName          = $AksClusterName
        NodeCount            = $AksNodeCount
        NodeVmSize           = $AksNodeVmSize
        GhcrOwner            = $GhcrOwner
        AuthDevJwtSecret     = $AuthDevJwtSecret
        InternalSharedSecret = $InternalSharedSecret
        CorsAllowedOrigins   = $CorsAllowedOrigins
    }
    if ($GhcrUser)               { $aksParams.GhcrUser = $GhcrUser }
    if ($GhcrToken)              { $aksParams.GhcrToken = $GhcrToken }
    if ($SubscriptionId)         { $aksParams.SubscriptionId = $SubscriptionId }
    if ($SkipGhcrImagePullSecret) { $aksParams.SkipGhcrImagePullSecret = $true }
    if ($SkipFrontend)           { $aksParams.SkipFrontend = $true }
    if ($SkipAksClusterCreate)   { $aksParams.SkipClusterCreate = $true }
    if ($ReleaseManifestPath)    { $aksParams.ReleaseManifestPath = $ReleaseManifestPath }

    & "$here/deploy-azure-aks.ps1" @aksParams
    return
}

# Container Apps path delegates to deploy-azure-containerapps.ps1
if ($ContainerApps) {
    Write-Host ''
    Write-Host '=== Aura full release (Container Apps) ===' -ForegroundColor Cyan
    $caLocation = if ($Location -eq 'centralus') { 'westus' } else { $Location }
    $caParams = @{
        PackageVersion       = $PackageVersion
        UniqueSuffix         = $UniqueSuffix
        ResourceGroupName    = $ResourceGroupName
        Location             = $caLocation
        EnvironmentName      = $ContainerAppsEnvName
        GhcrOwner            = $GhcrOwner
        AuthDevJwtSecret     = $AuthDevJwtSecret
        InternalSharedSecret = $InternalSharedSecret
        CorsAllowedOrigins   = $CorsAllowedOrigins
    }
    if ($GhcrUser)                    { $caParams.GhcrUser = $GhcrUser }
    if ($GhcrToken)                   { $caParams.GhcrToken = $GhcrToken }
    if ($SubscriptionId)              { $caParams.SubscriptionId = $SubscriptionId }
    if ($SkipGhcrImagePullSecret)     { $caParams.SkipGhcrImagePullSecret = $true }
    if ($SkipFrontend)                { $caParams.SkipFrontend = $true }
    if ($SkipContainerAppsEnvCreate)  { $caParams.SkipEnvironmentCreate = $true }
    if ($ReleaseManifestPath)         { $caParams.ReleaseManifestPath = $ReleaseManifestPath }

    & "$here/deploy-azure-containerapps.ps1" @caParams
    return
}

# VM path delegates to deploy-azure-vm.ps1
if ($VM) {
    Write-Host ''
    Write-Host '=== Aura full release (VM) ===' -ForegroundColor Cyan
    $vmParams = @{
        PackageVersion       = $PackageVersion
        UniqueSuffix         = $UniqueSuffix
        ResourceGroupName    = $VmResourceGroup
        VmName               = $VmName
        SshUser              = $VmSshUser
        GhcrOwner            = $GhcrOwner
        AuthDevJwtSecret     = $AuthDevJwtSecret
        InternalSharedSecret = $InternalSharedSecret
        CorsAllowedOrigins   = $CorsAllowedOrigins
    }
    if ($VmPublicIp)              { $vmParams.VmPublicIp = $VmPublicIp }
    if ($GhcrUser)                { $vmParams.GhcrUser = $GhcrUser }
    if ($GhcrToken)               { $vmParams.GhcrToken = $GhcrToken }
    if ($SubscriptionId)          { $vmParams.SubscriptionId = $SubscriptionId }
    if ($SkipVmNsgRules)          { $vmParams.SkipNsgRules = $true }
    if ($SkipVmDockerInstall)     { $vmParams.SkipDockerInstall = $true }
    if ($SkipFrontend)            { $vmParams.SkipFrontend = $true }
    if ($ReleaseManifestPath)     { $vmParams.ReleaseManifestPath = $ReleaseManifestPath }

    & "$here/deploy-azure-vm.ps1" @vmParams
    return
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
if ($Local -and -not $SkipFrontend) {
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

$targetLabel = if ($Local) { 'Docker Desktop Kubernetes (-Local)' } else { 'Azure App Service + Static Web App' }
Write-Host ''
Write-Host '=== Aura full release ===' -ForegroundColor Cyan
Write-Host "    PackageVersion (GHCR tag): $PackageVersion"
Write-Host "    Target: $targetLabel"
if (-not $SkipFrontend) {
    if ($Local) { Write-Host '    Frontend: http://127.0.0.1:30900 (Kubernetes)' -ForegroundColor Yellow }
    else        { Write-Host '    Frontend: Azure Static Web App (calm-stone-06e76f010.7.azurestaticapps.net)' -ForegroundColor Yellow }
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

if (-not $SkipFrontend) {
    Write-Host '--- aura-frontend ---' -ForegroundColor Cyan

    $fe = @{
        UniqueSuffix             = $UniqueSuffix
        GhcrOwner                = $GhcrOwner
        PackageVersion           = $PackageVersion
        GhcrUser                 = $GhcrUser
        GhcrToken                = $GhcrToken
        SkipGhcrImagePullSecret = $SkipGhcrImagePullSecret
    }

    if ($Local) {
        $fe.Local = $true
    }
    else {
        $fe.StaticSiteResourceId     = $StaticSiteResourceId
        $fe.UseAzureCliForToken      = $UseAzureCliForToken
        $fe.SkipBffConnectivityCheck = $SkipBffConnectivityCheck
        $fe.StrictBffBundleMatch     = $StrictBffBundleMatch
        $fe.BackendStackSummaryJson  = $ReleaseManifestPath
    }

    & "$here/deploy-azure-frontend.ps1" @fe
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
