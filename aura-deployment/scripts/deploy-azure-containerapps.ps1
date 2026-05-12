<#
.SYNOPSIS
  Deploy all Aura services to Azure Container Apps (free tier, Consumption plan).

.DESCRIPTION
  1. Registers Microsoft.App and Microsoft.OperationalInsights providers if needed.
  2. Creates (or reuses) a Container Apps Environment (Consumption plan - free tier).
  3. Deploys all five services as individual Container Apps:
       - worker, supervisor, authz: internal ingress (cluster-only)
       - bff-api: external ingress (public HTTPS + WebSocket)
       - frontend: external ingress (public HTTPS)
  4. Writes a release manifest JSON with public endpoints.

  Free tier includes:
    - 180,000 vCPU-seconds/month
    - 360,000 GiB-seconds/month
    - 2 million requests/month

  Prerequisites:
    - az login
    - GHCR images pushed at PackageVersion
    - $env:GITHUB_TOKEN set (or pass -GhcrToken)

.EXAMPLE
  az login
  $env:GITHUB_TOKEN = '<pat read:packages>'
  .\deploy-azure-containerapps.ps1 -PackageVersion 'local-ui-fix-8' -UniqueSuffix 'jdoe01'

.EXAMPLE
  .\deploy-azure-containerapps.ps1 -PackageVersion 'latest' -UniqueSuffix 'jdoe01' -Location 'eastus'
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$PackageVersion,
    [Parameter(Mandatory)][string]$UniqueSuffix,

    [string]$ResourceGroupName = 'aura',
    [string]$Location = 'westus',
    [string]$EnvironmentName = 'aura-cae',

    [string]$GhcrOwner = 'sushanth262',
    [string]$GhcrUser,
    [string]$GhcrToken,
    [string]$SubscriptionId,

    [string]$AuthDevJwtSecret = 'aura-dev-secret-change-me-use-32chars-minimum!!',
    [string]$InternalSharedSecret = 'aura-internal-demo',
    [string]$CorsAllowedOrigins,

    [switch]$SkipGhcrImagePullSecret,
    [switch]$SkipFrontend,
    [switch]$SkipEnvironmentCreate,

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

if ([string]::IsNullOrWhiteSpace($ReleaseManifestPath)) {
    $ReleaseManifestPath = Join-Path $PWD ('aura-release-{0}-{1}-containerapps.json' -f $suffix, $stem)
}

$images = @{
    Worker     = "ghcr.io/$GhcrOwner/aura-worker`:$tag"
    Supervisor = "ghcr.io/$GhcrOwner/aura-supervisor`:$tag"
    Authz      = "ghcr.io/$GhcrOwner/aura-authz`:$tag"
    Bff        = "ghcr.io/$GhcrOwner/aura-bff-api`:$tag"
    Frontend   = "ghcr.io/$GhcrOwner/aura-frontend`:$tag"
}

$appNames = @{
    Worker     = "aura-$suffix-worker"
    Supervisor = "aura-$suffix-supervisor"
    Authz      = "aura-$suffix-authz"
    Bff        = "aura-$suffix-bff"
    Frontend   = "aura-$suffix-frontend"
}

Write-Host ''
Write-Host '=== Aura Container Apps deployment ===' -ForegroundColor Cyan
Write-Host "    PackageVersion (GHCR tag): $tag"
Write-Host "    Location: $Location"
Write-Host "    Environment: $EnvironmentName"
Write-Host "    Apps: $($appNames.Values -join ', ')"
Write-Host ''

# ---- Resolve GHCR credentials ----
$registryArgs = @()
if (-not $SkipGhcrImagePullSecret) {
    $creds = Resolve-GhcrCredentials -GhcrUser $GhcrUser -GhcrToken $GhcrToken
    $registryArgs = @(
        '--registry-server', 'ghcr.io',
        '--registry-username', $creds.User,
        '--registry-password', $creds.Token
    )
}

# ---- Step 1: Register providers ----
Write-Host '--- 1/5 registering providers ---' -ForegroundColor Cyan
foreach ($ns in @('Microsoft.App', 'Microsoft.OperationalInsights')) {
    $oldEA = $ErrorActionPreference
    $ErrorActionPreference = 'SilentlyContinue'
    $regState = az provider show --namespace $ns --query "registrationState" -o tsv --only-show-errors 2>&1 | Where-Object { $_ -notmatch 'ERROR' }
    $ErrorActionPreference = $oldEA

    if ($regState -ne 'Registered') {
        Write-Host "  Registering $ns..."
        $ErrorActionPreference = 'SilentlyContinue'
        az provider register --namespace $ns --only-show-errors 2>&1 | Out-Null
        $ErrorActionPreference = $oldEA

        $waited = 0
        while ($waited -lt 300) {
            Start-Sleep -Seconds 15
            $waited += 15
            $ErrorActionPreference = 'SilentlyContinue'
            $regState = az provider show --namespace $ns --query "registrationState" -o tsv --only-show-errors 2>&1 | Where-Object { $_ -notmatch 'ERROR' }
            $ErrorActionPreference = $oldEA
            if ($regState -eq 'Registered') { break }
            Write-Host "    Waiting for $ns... ($regState, ${waited}s)" -ForegroundColor Yellow
        }
        if ($regState -ne 'Registered') { throw "$ns provider failed to register within 5 minutes." }
    }
    Write-Host "  $ns - Registered" -ForegroundColor Green
}

# ---- Step 2: Resource group + environment ----
Write-Host '--- 2/5 resource group + environment ---' -ForegroundColor Cyan
Ensure-ResourceGroup -Name $ResourceGroupName -Location $Location

if (-not $SkipEnvironmentCreate) {
    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = 'SilentlyContinue'
    az containerapp env show --name $EnvironmentName --resource-group $ResourceGroupName -o none --only-show-errors 2>&1 | Out-Null
    $envExists = ($LASTEXITCODE -eq 0)
    $ErrorActionPreference = $oldPref

    if ($envExists) {
        Write-Host "Container Apps Environment '$EnvironmentName' already exists." -ForegroundColor Green
    }
    else {
        Write-Host "Creating Container Apps Environment '$EnvironmentName' (this may take 1-2 minutes)..."
        Invoke-AzWithRetry -Label "containerapp env create ($EnvironmentName)" -ScriptBlock {
            az containerapp env create `
                --name $EnvironmentName `
                --resource-group $ResourceGroupName `
                --location $Location `
                --only-show-errors `
                -o none
        }
        Write-Host "Environment '$EnvironmentName' created." -ForegroundColor Green
    }
}

# ---- Helper: deploy or update a container app ----
function Deploy-ContainerApp {
    param(
        [Parameter(Mandatory)][string]$AppName,
        [Parameter(Mandatory)][string]$Image,
        [int]$TargetPort = 8080,
        [string]$IngressType = 'internal',
        [hashtable]$EnvVars = @{},
        [double]$Cpu = 0.25,
        [double]$Memory = 0.5
    )

    $envArgs = @()
    foreach ($key in $EnvVars.Keys) {
        $envArgs += "$key=$($EnvVars[$key])"
    }

    $oldPref = $ErrorActionPreference
    $ErrorActionPreference = 'SilentlyContinue'
    az containerapp show --name $AppName --resource-group $ResourceGroupName -o none --only-show-errors 2>&1 | Out-Null
    $appExists = ($LASTEXITCODE -eq 0)
    $ErrorActionPreference = $oldPref

    if ($appExists) {
        Write-Host "    Updating $AppName..."
        $updateCmd = @(
            'containerapp', 'update',
            '--name', $AppName,
            '--resource-group', $ResourceGroupName,
            '--image', $Image,
            '--only-show-errors',
            '-o', 'none'
        )
        if ($envArgs.Count -gt 0) {
            $updateCmd += '--set-env-vars'
            $updateCmd += $envArgs
        }
        Invoke-AzWithRetry -Label "containerapp update ($AppName)" -ScriptBlock {
            & az @updateCmd
        }
    }
    else {
        Write-Host "    Creating $AppName..."
        $createCmd = @(
            'containerapp', 'create',
            '--name', $AppName,
            '--resource-group', $ResourceGroupName,
            '--environment', $EnvironmentName,
            '--image', $Image,
            '--target-port', $TargetPort,
            '--ingress', $IngressType,
            '--cpu', $Cpu,
            '--memory', "${Memory}Gi",
            '--min-replicas', '0',
            '--max-replicas', '1',
            '--only-show-errors',
            '-o', 'none'
        )
        $createCmd += $registryArgs
        if ($envArgs.Count -gt 0) {
            $createCmd += '--env-vars'
            $createCmd += $envArgs
        }
        Invoke-AzWithRetry -Label "containerapp create ($AppName)" -ScriptBlock {
            & az @createCmd
        }
    }
}

function Get-ContainerAppFqdn {
    param([Parameter(Mandatory)][string]$AppName)
    $fqdn = Invoke-AzWithRetry -Label "get fqdn ($AppName)" -ScriptBlock {
        az containerapp show --name $AppName --resource-group $ResourceGroupName --query "properties.configuration.ingress.fqdn" -o tsv --only-show-errors
    }
    return ($fqdn | Where-Object { $_ -match '\.' } | Select-Object -First 1)
}

# ---- Step 3: Deploy internal services ----
Write-Host '--- 3/5 deploying internal services ---' -ForegroundColor Cyan

Write-Host '  worker...'
Deploy-ContainerApp -AppName $appNames.Worker -Image $images.Worker -TargetPort 8083 -IngressType 'internal' -EnvVars @{
    HTTP_ADDR              = ':8083'
    WORKER_ENABLED_SOURCES = 'grafana,github,jira'
}
Write-Host '  worker deployed.' -ForegroundColor Green

$workerFqdn = Get-ContainerAppFqdn -AppName $appNames.Worker
$workerUrl = "https://$workerFqdn"

Write-Host '  supervisor...'
Deploy-ContainerApp -AppName $appNames.Supervisor -Image $images.Supervisor -TargetPort 8082 -IngressType 'internal' -EnvVars @{
    HTTP_ADDR              = ':8082'
    AUTH_DEV_MOCK          = 'true'
    AUTH_DEV_JWT_SECRET    = $AuthDevJwtSecret
    POLICY_VERSION         = 'stub-v1'
    WORKER_URL             = $workerUrl
    WORKER_SOURCES         = 'grafana,github,jira'
    INTERNAL_SHARED_SECRET = $InternalSharedSecret
}
Write-Host '  supervisor deployed.' -ForegroundColor Green

$supervisorFqdn = Get-ContainerAppFqdn -AppName $appNames.Supervisor
$supervisorUrl = "https://$supervisorFqdn"

Write-Host '  authz...'
Deploy-ContainerApp -AppName $appNames.Authz -Image $images.Authz -TargetPort 8081 -IngressType 'internal' -EnvVars @{
    HTTP_ADDR      = ':8081'
    POLICY_VERSION = 'stub-v1'
}
Write-Host '  authz deployed.' -ForegroundColor Green

$authzFqdn = Get-ContainerAppFqdn -AppName $appNames.Authz
$authzUrl = "https://$authzFqdn"

# ---- Step 4: Deploy external services ----
Write-Host '--- 4/5 deploying external services ---' -ForegroundColor Cyan

$corsOrigins = $CorsAllowedOrigins
if ([string]::IsNullOrWhiteSpace($corsOrigins)) {
    $corsOrigins = 'https://calm-stone-06e76f010.7.azurestaticapps.net,http://localhost:8081,http://localhost:19006'
}

Write-Host '  bff-api...'
Deploy-ContainerApp -AppName $appNames.Bff -Image $images.Bff -TargetPort 8080 -IngressType 'external' -Cpu 0.5 -Memory 1.0 -EnvVars @{
    HTTP_ADDR              = ':8080'
    AUTH_DEV_MOCK          = 'true'
    AUTH_DEV_JWT_SECRET    = $AuthDevJwtSecret
    AUTHZ_URL              = $authzUrl
    SUPERVISOR_URL         = $supervisorUrl
    INTERNAL_SHARED_SECRET = $InternalSharedSecret
    CORS_ALLOWED_ORIGINS   = $corsOrigins
}
Write-Host '  bff-api deployed.' -ForegroundColor Green

$bffFqdn = Get-ContainerAppFqdn -AppName $appNames.Bff

if (-not $SkipFrontend) {
    Write-Host '  frontend...'
    Deploy-ContainerApp -AppName $appNames.Frontend -Image $images.Frontend -TargetPort 80 -IngressType 'external' -Cpu 0.25 -Memory 0.5
    Write-Host '  frontend deployed.' -ForegroundColor Green
}

$frontendFqdn = $null
if (-not $SkipFrontend) {
    $frontendFqdn = Get-ContainerAppFqdn -AppName $appNames.Frontend

    # Update CORS to include the frontend URL
    $feOrigin = "https://$frontendFqdn"
    if ($corsOrigins -notmatch [regex]::Escape($frontendFqdn)) {
        $updatedCors = $corsOrigins + ',' + $feOrigin
        Write-Host '  Updating BFF CORS to include frontend URL...'
        Invoke-AzWithRetry -Label 'update BFF CORS' -ScriptBlock {
            az containerapp update `
                --name $appNames.Bff `
                --resource-group $ResourceGroupName `
                --set-env-vars "CORS_ALLOWED_ORIGINS=$updatedCors" `
                --only-show-errors -o none
        }
    }
}

# ---- Step 5: Release manifest ----
Write-Host '--- 5/5 release manifest ---' -ForegroundColor Cyan

$manifest = @{
    deployment_target = 'container-apps'
    environment       = $EnvironmentName
    location          = $Location
    image_tag         = $tag
    ghcr_owner        = $GhcrOwner
    endpoints         = @{
        bff_public = "https://$bffFqdn"
        bff_api    = "https://$bffFqdn/v1"
        bff_ws     = "wss://$bffFqdn"
    }
    services = @{
        worker     = @{ app = $appNames.Worker;     image = $images.Worker;     ingress = 'internal'; fqdn = $workerFqdn }
        supervisor = @{ app = $appNames.Supervisor;  image = $images.Supervisor; ingress = 'internal'; fqdn = $supervisorFqdn }
        authz      = @{ app = $appNames.Authz;       image = $images.Authz;      ingress = 'internal'; fqdn = $authzFqdn }
        bff_api    = @{ app = $appNames.Bff;          image = $images.Bff;        ingress = 'external'; fqdn = $bffFqdn }
    }
    timestamp = (Get-Date -Format 'o')
}
if ($frontendFqdn) {
    $manifest.endpoints.frontend = "https://$frontendFqdn"
    $manifest.services.frontend = @{ app = $appNames.Frontend; image = $images.Frontend; ingress = 'external'; fqdn = $frontendFqdn }
}

$manifest | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $ReleaseManifestPath -Encoding UTF8

Write-Host ''
Write-Host '============================================' -ForegroundColor Green
Write-Host ' Aura Container Apps deployment complete!' -ForegroundColor Green
Write-Host '============================================' -ForegroundColor Green
Write-Host ''
Write-Host "  Environment : $EnvironmentName ($Location)" -ForegroundColor White
Write-Host "  BFF API     : https://$bffFqdn/v1" -ForegroundColor Cyan
Write-Host "  BFF WS      : wss://$bffFqdn" -ForegroundColor Cyan
if ($frontendFqdn) {
    Write-Host "  Frontend    : https://$frontendFqdn" -ForegroundColor Cyan
}
Write-Host ''
Write-Host "  Release manifest: $ReleaseManifestPath" -ForegroundColor White
Write-Host ''
Write-Host '  az commands:' -ForegroundColor DarkGray
Write-Host "    az containerapp list -g $ResourceGroupName -o table" -ForegroundColor DarkGray
Write-Host "    az containerapp logs show -n $($appNames.Bff) -g $ResourceGroupName --follow" -ForegroundColor DarkGray
Write-Host ''
