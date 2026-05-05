<#
.SYNOPSIS
  Build the Expo web bundle (static) and deploy to Azure Static Web Apps (Free tier friendly).

.DESCRIPTION
  1. Runs npm ci / npm install in aura-frontend (sibling folder under the repo root)
  2. Runs npx expo export --platform web → output folder (default: dist)
  3. Deploys with Azure Static Web Apps CLI (swa deploy)

  Script lives under aura-deployment/scripts; default AppRoot is <repo>/aura-frontend.

  Target Static Web App (defaults — change params if yours differs):
    /subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourcegroups/aura/providers/Microsoft.Web/staticSites/Aura

  Deployment always goes to whichever Static Web App issued the deployment token (token is bound to that app).

  Token sources (first wins):
    1. -UseAzureCliForToken  →  az staticwebapp secrets list (needs az login + permission)
    2. -DeploymentToken (SecureString)
    3. Environment: AZURE_STATIC_WEB_APPS_API_TOKEN or SWA_CLI_DEPLOYMENT_TOKEN
    4. Portal → Static Web App → Overview → Manage deployment token

.EXAMPLE
  cd <repo-root>
  .\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -UseAzureCliForToken

.EXAMPLE
  $env:AZURE_STATIC_WEB_APPS_API_TOKEN = '<token-from-portal>'
  .\aura-deployment\scripts\deploy-azure-static-web-app.ps1

.EXAMPLE
  .\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -DeploymentToken (Read-Host -AsSecureString) -SkipBuild
#>

[CmdletBinding()]
param(
    [string] $AppRoot,
    [string] $OutputDir = 'dist',
    [ValidateSet('production', 'preview')]
    [string] $Environment = 'production',
    [switch] $SkipNpmInstall,
    [switch] $SkipBuild,
    [SecureString] $DeploymentToken,

    # ARM id for the target Static Web App (deployment token is scoped to this app).
    [string] $StaticSiteResourceId = '/subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourcegroups/aura/providers/Microsoft.Web/staticSites/Aura',

    [switch] $UseAzureCliForToken
)

$ErrorActionPreference = 'Stop'

if (-not $PSBoundParameters.ContainsKey('AppRoot') -or [string]::IsNullOrWhiteSpace($AppRoot)) {
    $deploymentRoot = Split-Path -LiteralPath $PSScriptRoot -Parent
    $repoRoot = Split-Path -LiteralPath $deploymentRoot -Parent
    $AppRoot = Join-Path $repoRoot 'aura-frontend'
}

function Get-PlainToken {
    param([SecureString] $Secure)
    if (-not $Secure) { return $null }
    $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($Secure)
    try { [System.Runtime.InteropServices.Marshal]::PtrToStringUni($bstr) }
    finally { [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr) }
}

function Resolve-StaticSiteFromResourceId {
    param([string] $ResourceId)
    $m = [regex]::Match(
        $ResourceId.Trim(),
        '(?i)^/subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft\.Web/staticSites/([^/]+)$'
    )
    if (-not $m.Success) {
        Write-Error "StaticSiteResourceId must look like: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/staticSites/{name}"
    }
    @{
        SubscriptionId    = $m.Groups[1].Value
        ResourceGroupName = $m.Groups[2].Value
        StaticSiteName    = $m.Groups[3].Value
    }
}

function Get-DeploymentTokenFromAzCli {
    param(
        [string] $SubscriptionId,
        [string] $ResourceGroupName,
        [string] $StaticSiteName
    )
    $az = Get-Command az -ErrorAction SilentlyContinue
    if (-not $az) {
        Write-Error 'Azure CLI (az) not found on PATH. Install: https://learn.microsoft.com/cli/azure/install-azure-cli-windows'
    }

    Write-Host "Using Azure CLI for deployment token (subscription=$SubscriptionId, rg=$ResourceGroupName, site=$StaticSiteName)..."
    az account set --subscription $SubscriptionId --only-show-errors 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Error "az account set failed. Run 'az login' and ensure you have access to subscription $SubscriptionId"
    }

    $apiKey = az staticwebapp secrets list `
        --name $StaticSiteName `
        --resource-group $ResourceGroupName `
        --subscription $SubscriptionId `
        --query 'properties.apiKey' `
        -o tsv `
        --only-show-errors

    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($apiKey)) {
        $jsonText = az staticwebapp secrets list `
            --name $StaticSiteName `
            --resource-group $ResourceGroupName `
            --subscription $SubscriptionId `
            -o json `
            --only-show-errors
        if ($LASTEXITCODE -ne 0) {
            Write-Error "az staticwebapp secrets list failed: $jsonText"
        }
        $obj = $jsonText | ConvertFrom-Json
        if ($null -ne $obj.properties -and $null -ne $obj.properties.apiKey) {
            $apiKey = [string]$obj.properties.apiKey
        }
        elseif ($obj -is [object[]] -and $obj.Count -gt 0 -and $obj[0].properties.apiKey) {
            $apiKey = [string]$obj[0].properties.apiKey
        }
    }

    if ([string]::IsNullOrWhiteSpace($apiKey)) {
        Write-Error 'Could not read deployment token from az staticwebapp secrets list. Copy the token from Portal → Static Web App → Overview → Manage deployment token.'
    }

    return $apiKey.Trim()
}

# Resolve subscription / RG / name from resource id when provided
$ctx = Resolve-StaticSiteFromResourceId -ResourceId $StaticSiteResourceId
$SubscriptionId = $ctx.SubscriptionId
$ResourceGroupName = $ctx.ResourceGroupName
$StaticSiteName = $ctx.StaticSiteName

$token = $null
if ($UseAzureCliForToken) {
    $token = Get-DeploymentTokenFromAzCli `
        -SubscriptionId $SubscriptionId `
        -ResourceGroupName $ResourceGroupName `
        -StaticSiteName $StaticSiteName
}

if (-not $token) {
    $token = Get-PlainToken -Secure $DeploymentToken
}
if (-not $token) {
    $token = $env:AZURE_STATIC_WEB_APPS_API_TOKEN
}
if (-not $token) {
    $token = $env:SWA_CLI_DEPLOYMENT_TOKEN
}

if (-not $token) {
    Write-Error @"
No deployment token found.

Options:
  .\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -UseAzureCliForToken
    (after az login; reads apiKey for Static Web App '$StaticSiteName' in rg '$ResourceGroupName')

Or set AZURE_STATIC_WEB_APPS_API_TOKEN from:
  Portal → Static Web App '$StaticSiteName' → Overview → Manage deployment token

Resource: /subscriptions/$SubscriptionId/resourceGroups/$ResourceGroupName/providers/Microsoft.Web/staticSites/$StaticSiteName
"@
}

if (-not (Test-Path -LiteralPath $AppRoot)) {
    Write-Error "AppRoot not found: $AppRoot (expected sibling aura-frontend next to aura-deployment, or pass -AppRoot)."
}

$outPath = Join-Path $AppRoot $OutputDir
Push-Location $AppRoot
try {
    if (-not $SkipNpmInstall) {
        if (Test-Path 'package-lock.json') {
            npm ci
        }
        else {
            npm install
        }
    }

    if (-not $SkipBuild) {
        Write-Host 'Exporting static web (Expo)...'
        npx expo export --platform web
    }

    if (-not (Test-Path -LiteralPath $outPath)) {
        Write-Error "Output folder missing after export: $outPath`nCheck Expo web output location in app.json / app.config."
    }

    Write-Host "Deploying to Azure Static Web App '$StaticSiteName' ($Environment) from $outPath ..."

    $env:SWA_CLI_DEPLOYMENT_TOKEN = $token
    try {
        npx --yes @azure/static-web-apps-cli deploy $outPath --env $Environment
    }
    finally {
        Remove-Item Env:SWA_CLI_DEPLOYMENT_TOKEN -ErrorAction SilentlyContinue
    }

    Write-Host "Deploy finished → resource group '$ResourceGroupName', app '$StaticSiteName'."
}
finally {
    Pop-Location
}
