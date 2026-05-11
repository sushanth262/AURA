<#
.SYNOPSIS
  Deploy aura-bff-api — depends on aura-authz and aura-supervisor (same -UniqueSuffix).

.DESCRIPTION
  Verifies authz and supervisor Web Apps exist and respond on /healthz (unless -SkipDependencyChecks).
  Sets AUTHZ_URL, SUPERVISOR_URL, INTERNAL_SHARED_SECRET, JWT secret (must match supervisor),
  and CORS_ALLOWED_ORIGINS.

    Use -PackageVersion for the GHCR tag; -ImageTag is an alias.

.EXAMPLE
  .\deploy-azure-bff-api.ps1 -UniqueSuffix 'jdoe01' -ResourceGroupName 'aura-rg' `
    -PackageVersion 'abc1234' `
    -CorsAllowedOrigins 'https://happy-rock-012345.azurestaticapps.net,http://localhost:19006'

.EXAMPLE
  # Local UI is http://127.0.0.1:30900 — include that origin (BFF NodePort 30880 is the API, not a browser Origin).
  .\deploy-azure-bff-api.ps1 -UniqueSuffix 'jdoe01' -PackageVersion 'latest' -Local `
    -CorsAllowedOrigins 'http://127.0.0.1:30900,http://localhost:19006'

  # With -Local, http://127.0.0.1:30900 is appended automatically if absent from -CorsAllowedOrigins.
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
    [string]$AuthzHttpsRoot,
    [string]$SupervisorHttpsRoot,
    [string]$AuthDevJwtSecret = 'aura-dev-secret-change-me-use-32chars-minimum!!',
    [string]$InternalSharedSecret = 'aura-internal-demo',
    [string]$CorsAllowedOrigins = 'http://localhost:8081,http://localhost:19006',
    [switch]$SkipDependencyChecks,
    [switch]$SkipHealthWait,
    [string]$SummaryOutPath,

    [switch]$Local,
    [switch]$SkipGhcrImagePullSecret
)

$ErrorActionPreference = 'Stop'
. "$PSScriptRoot/deploy-azure-backend-common.ps1"

$names = Get-AuraStandardAppNames -UniqueSuffix $UniqueSuffix
$tag = Resolve-AuraPackageImageTag -PackageVersion $PackageVersion -ImageTag $ImageTag
$image = "ghcr.io/$GhcrOwner/aura-bff-api`:$tag"

if ($Local) {
    # Browser Origin is the page URL (e.g. nginx UI on 30900), not the BFF NodePort.
    $corsLocalUi = 'http://127.0.0.1:30900'
    if ($CorsAllowedOrigins -notmatch '30900') {
        if ([string]::IsNullOrWhiteSpace($CorsAllowedOrigins)) {
            $CorsAllowedOrigins = $corsLocalUi
        }
        else {
            $CorsAllowedOrigins = ($CorsAllowedOrigins.Trim().TrimEnd(',')) + ',' + $corsLocalUi
        }
    }

    Assert-KubectlCli
    Warn-KubectlDockerDesktopContext
    $ns = Get-AuraKubernetesNamespaceFromSuffix -UniqueSuffix $UniqueSuffix
    Ensure-AuraKubernetesNamespace -Namespace $ns
    $authzHost = 'http://127.0.0.1:30881'
    $supervisorHost = 'http://127.0.0.1:30882'
    $null = Assert-HttpDependencyHealthyOrThrow -RootUrl $authzHost -DependencyLabel 'aura-authz' -SkipHealthCheck:$SkipDependencyChecks
    $null = Assert-HttpDependencyHealthyOrThrow -RootUrl $supervisorHost -DependencyLabel 'aura-supervisor' -SkipHealthCheck:$SkipDependencyChecks
    if (-not $SkipGhcrImagePullSecret) {
        $credsL = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken
        Ensure-AuraK8sGhcrPullSecret -Namespace $ns -DockerUsername $credsL.User -DockerPassword $credsL.Token
    }
    $pullBlock = Resolve-AuraK8sImagePullSecretsYamlBlock -SkipGhcrImagePullSecret:$SkipGhcrImagePullSecret
    $authzCluster = 'http://aura-authz:8081'
    $supervisorCluster = 'http://aura-supervisor:8082'
    Deploy-AuraK8sServiceFromDockerDesktopTemplate -TemplateFileName 'bff.yaml' -Replacements @{
        '{{NAMESPACE}}'              = $ns
        '{{BFF_IMAGE}}'              = $image
        '{{IMAGE_PULL_SECRETS}}'     = $pullBlock
        '{{AUTH_DEV_JWT_SECRET}}'    = (Escape-YamlSingleQuotedLiteral -Value $AuthDevJwtSecret)
        '{{INTERNAL_SHARED_SECRET}}' = (Escape-YamlSingleQuotedLiteral -Value $InternalSharedSecret)
        '{{AUTHZ_URL}}'              = $authzCluster
        '{{SUPERVISOR_URL}}'         = $supervisorCluster
        '{{CORS_ALLOWED_ORIGINS}}'   = (Escape-YamlSingleQuotedLiteral -Value $CorsAllowedOrigins)
    }
    Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-bff-api'
    $httpsRoot = 'http://127.0.0.1:30880'
    Write-Host ''
    Write-Host '=== aura-bff-api deployed (Kubernetes) ===' -ForegroundColor Green
    Write-Host "  Namespace       : $ns"
    Write-Host "  AUTHZ_URL       : $authzCluster"
    Write-Host "  SUPERVISOR_URL  : $supervisorCluster"
    Write-Host "  Host URL        : $httpsRoot"
    Write-Host ''
    Write-Host '  Expo hints:'
    Write-Host "    EXPO_PUBLIC_API_BASE_URL=$httpsRoot/v1"
    Write-Host "    EXPO_PUBLIC_WS_BASE_URL=ws://127.0.0.1:30880"
    Write-Host ''
    if (-not $SkipHealthWait) {
        Wait-AuraHealthEndpoint -HttpsRoot $httpsRoot -Label 'aura-bff-api'
    }
    if ($SummaryOutPath) {
        @{
            bff_https             = $httpsRoot
            kubernetes_namespace  = $ns
            authz_url_used        = $authzCluster
            supervisor_url_used   = $supervisorCluster
        } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
    }
    return
}

Assert-AzLoggedIn
if ($SubscriptionId) {
    az account set --subscription $SubscriptionId --only-show-errors
    if ($LASTEXITCODE -ne 0) { throw 'az account set failed' }
}

$creds = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken

if ([string]::IsNullOrWhiteSpace($AuthzHttpsRoot)) {
    $AuthzHttpsRoot = Assert-PeerHealthyOrThrow `
        -ResourceGroup $ResourceGroupName `
        -AppName $names.Authz `
        -DependencyLabel 'aura-authz' `
        -SkipHealthCheck:$SkipDependencyChecks
}
else {
    if (-not $SkipDependencyChecks -and -not (Test-AuraHealthEndpoint -HttpsRoot $AuthzHttpsRoot)) {
        throw "AuthZ not healthy at $AuthzHttpsRoot/healthz"
    }
}

if ([string]::IsNullOrWhiteSpace($SupervisorHttpsRoot)) {
    $SupervisorHttpsRoot = Assert-PeerHealthyOrThrow `
        -ResourceGroup $ResourceGroupName `
        -AppName $names.Supervisor `
        -DependencyLabel 'aura-supervisor' `
        -SkipHealthCheck:$SkipDependencyChecks
}
else {
    if (-not $SkipDependencyChecks -and -not (Test-AuraHealthEndpoint -HttpsRoot $SupervisorHttpsRoot)) {
        throw "Supervisor not healthy at $SupervisorHttpsRoot/healthz"
    }
}

$authzUrl = $AuthzHttpsRoot.TrimEnd('/')
$supUrl = $SupervisorHttpsRoot.TrimEnd('/')

$extra = @{
    AUTH_DEV_MOCK       = 'true'
    AUTH_DEV_JWT_SECRET = $AuthDevJwtSecret
    AUTHZ_URL           = $authzUrl
    SUPERVISOR_URL      = $supUrl
    INTERNAL_SHARED_SECRET = $InternalSharedSecret
    CORS_ALLOWED_ORIGINS   = $CorsAllowedOrigins
}

$fqdn = Deploy-AuraLinuxContainerWebApp `
    -ResourceGroup $ResourceGroupName `
    -Location $Location `
    -PlanName $SharedPlanName `
    -AppName $names.Bff `
    -ContainerImage $image `
    -GhcrUser $creds.User `
    -GhcrToken $creds.Token `
    -WebSitesPort 8080 `
    -ExtraAppSettings $extra `
    -Sku $AppServiceSku

$httpsRoot = "https://$fqdn"
Write-Host ""
Write-Host "=== aura-bff-api deployed ===" -ForegroundColor Green
Write-Host "  PackageVersion (image tag): $tag"
Write-Host "  Web App name : $($names.Bff)"
Write-Host "  Public URL   : $httpsRoot"
Write-Host "  AUTHZ_URL    : $authzUrl"
Write-Host "  SUPERVISOR_URL : $supUrl"
Write-Host ""
Write-Host "  Expo / SWA hints:"
Write-Host "    EXPO_PUBLIC_API_BASE_URL=$httpsRoot/v1"
Write-Host "    EXPO_PUBLIC_WS_BASE_URL=wss://$fqdn"
Write-Host ""

if (-not $SkipHealthWait) {
    Wait-AuraHealthEndpoint -HttpsRoot $httpsRoot -Label 'aura-bff-api'
}

if ($SummaryOutPath) {
    @{
        bff_https           = $httpsRoot
        bff_app             = $names.Bff
        authz_url_used      = $authzUrl
        supervisor_url_used = $supUrl
    } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
}
