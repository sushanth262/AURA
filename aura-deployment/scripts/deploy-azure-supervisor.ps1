<#
.SYNOPSIS
  Deploy aura-supervisor — expects aura-worker already deployed (same -UniqueSuffix).

.DESCRIPTION
  Resolves WORKER_URL from the sibling worker Web App (Azure CLI), verifies /healthz unless -SkipDependencyChecks.
  Configures WORKER_URL, WORKER_SOURCES, JWT secret (must match BFF), INTERNAL_SHARED_SECRET, POLICY_VERSION.

    Use -PackageVersion for the GHCR tag pushed by Terraform; -ImageTag is an alias.

  -Local deploys to Docker Desktop Kubernetes (see worker.yaml sibling manifests).

.EXAMPLE
  .\deploy-azure-supervisor.ps1 -UniqueSuffix 'jdoe01' -ResourceGroupName 'aura-rg' -PackageVersion 'abc1234'

.EXAMPLE
  .\deploy-azure-supervisor.ps1 -UniqueSuffix 'jdoe01' -PackageVersion 'latest' -Local `
    -AuthDevJwtSecret 'aura-dev-secret-change-me-use-32chars-minimum!!' `
    -InternalSharedSecret 'aura-internal-demo'
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$UniqueSuffix,
    [string]$ResourceGroupName = 'aura',
    [string]$Location = 'centralus',
    [string]$SharedPlanName = 'aura-free-plan',
    [string]$AppServiceSku = 'B1',
    [string]$GhcrOwner = 'sushanth262',
    [string]$PackageVersion,
    [string]$ImageTag,
    [string]$GhcrUser,
    [string]$GhcrToken,
    [string]$SubscriptionId,
    [string]$WorkerWebAppName,
    [string]$WorkerHttpsRoot,
    [string]$AuthDevJwtSecret = 'aura-dev-secret-change-me-use-32chars-minimum!!',
    [string]$InternalSharedSecret = 'aura-internal-demo',
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
$supervisorImage = "ghcr.io/$GhcrOwner/aura-supervisor`:$tag"

if ($Local) {
    Assert-KubectlCli
    Warn-KubectlDockerDesktopContext
    $ns = Get-AuraKubernetesNamespaceFromSuffix -UniqueSuffix $UniqueSuffix
    Ensure-AuraKubernetesNamespace -Namespace $ns
    $workerHostRoot = 'http://127.0.0.1:30883'
    $null = Assert-HttpDependencyHealthyOrThrow -RootUrl $workerHostRoot -DependencyLabel 'aura-worker' -SkipHealthCheck:$SkipDependencyChecks
    if (-not $SkipGhcrImagePullSecret) {
        $credsL = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken
        Ensure-AuraK8sGhcrPullSecret -Namespace $ns -DockerUsername $credsL.User -DockerPassword $credsL.Token
    }
    $pullBlock = Resolve-AuraK8sImagePullSecretsYamlBlock -SkipGhcrImagePullSecret:$SkipGhcrImagePullSecret
    $workerClusterUrl = 'http://aura-worker:8083'
    Deploy-AuraK8sServiceFromDockerDesktopTemplate -TemplateFileName 'supervisor.yaml' -Replacements @{
        '{{NAMESPACE}}'               = $ns
        '{{SUPERVISOR_IMAGE}}'        = $supervisorImage
        '{{IMAGE_PULL_SECRETS}}'      = $pullBlock
        '{{AUTH_DEV_JWT_SECRET}}'     = (Escape-YamlSingleQuotedLiteral -Value $AuthDevJwtSecret)
        '{{INTERNAL_SHARED_SECRET}}'  = (Escape-YamlSingleQuotedLiteral -Value $InternalSharedSecret)
        '{{WORKER_URL}}'              = $workerClusterUrl
    }
    Wait-AuraKubernetesRollout -Namespace $ns -DeploymentName 'aura-supervisor'
    $httpsRoot = 'http://127.0.0.1:30882'
    Write-Host ''
    Write-Host '=== aura-supervisor deployed (Kubernetes) ===' -ForegroundColor Green
    Write-Host "  Namespace    : $ns"
    Write-Host "  WORKER_URL   : $workerClusterUrl"
    Write-Host "  Host URL     : $httpsRoot"
    Write-Host ''
    if (-not $SkipHealthWait) {
        Wait-AuraHealthEndpoint -HttpsRoot $httpsRoot -Label 'aura-supervisor'
    }
    if ($SummaryOutPath) {
        @{
            supervisor_https  = $httpsRoot
            kubernetes_namespace = $ns
            worker_url_used   = $workerClusterUrl
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
$wApp = if ($WorkerWebAppName) { $WorkerWebAppName } else { $names.Worker }

if ([string]::IsNullOrWhiteSpace($WorkerHttpsRoot)) {
    $workerRoot = Assert-PeerHealthyOrThrow `
        -ResourceGroup $ResourceGroupName `
        -AppName $wApp `
        -DependencyLabel 'aura-worker' `
        -SkipHealthCheck:$SkipDependencyChecks
}
else {
    $workerRoot = $WorkerHttpsRoot.TrimEnd('/')
    if (-not $SkipDependencyChecks) {
        if (-not (Test-AuraHealthEndpoint -HttpsRoot $workerRoot)) {
            throw "Worker not healthy at $workerRoot/healthz - deploy worker first or fix WorkerHttpsRoot."
        }
    }
}

$extra = @{
    AUTH_DEV_MOCK              = 'true'
    AUTH_DEV_JWT_SECRET        = $AuthDevJwtSecret
    POLICY_VERSION             = 'stub-v1'
    WORKER_URL                 = $workerRoot
    WORKER_SOURCES             = 'grafana,github,jira'
    INTERNAL_SHARED_SECRET     = $InternalSharedSecret
}

$fqdn = Deploy-AuraLinuxContainerWebApp `
    -ResourceGroup $ResourceGroupName `
    -Location $Location `
    -PlanName $SharedPlanName `
    -AppName $names.Supervisor `
    -ContainerImage $supervisorImage `
    -GhcrUser $creds.User `
    -GhcrToken $creds.Token `
    -WebSitesPort 8082 `
    -ExtraAppSettings $extra `
    -Sku $AppServiceSku

$httpsRoot = "https://$fqdn"
Write-Host ""
Write-Host "=== aura-supervisor deployed ===" -ForegroundColor Green
Write-Host "  PackageVersion (image tag): $tag"
Write-Host "  Web App name : $($names.Supervisor)"
Write-Host "  Public URL   : $httpsRoot"
Write-Host "  WORKER_URL   : $workerRoot"
Write-Host ""

if (-not $SkipHealthWait) {
    Wait-AuraHealthEndpoint -HttpsRoot $httpsRoot -Label 'aura-supervisor'
}

if ($SummaryOutPath) {
    @{
        supervisor_https = $httpsRoot
        supervisor_app   = $names.Supervisor
        worker_url_used  = $workerRoot
    } | ConvertTo-Json | Set-Content -LiteralPath $SummaryOutPath -Encoding UTF8
}
