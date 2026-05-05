<#
.SYNOPSIS
  Pull the Aura frontend container image from GHCR, extract the static web bundle, deploy to Azure Static Web Apps.

.DESCRIPTION
  1. docker pull <image> (unless -SkipDockerPull)
  2. docker create + docker cp static files from the nginx stage (default path /usr/share/nginx/html)
  3. npx @azure/static-web-apps-cli deploy <staging>

  Matches aura-deployment/services/frontend/Dockerfile: assets live under /usr/share/nginx/html.

  Target Static Web App (defaults - change -StaticSiteResourceId if yours differs):
    /subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourcegroups/aura/providers/Microsoft.Web/staticSites/Aura

  Token sources (first wins):
    1. -UseAzureCliForToken -> az staticwebapp secrets list (needs az login + permission)
    2. -DeploymentToken (SecureString)
    3. AZURE_STATIC_WEB_APPS_API_TOKEN or SWA_CLI_DEPLOYMENT_TOKEN

  Prerequisites: Docker CLI, Node/npm (for npx swa), optional az for token fetch.
  Private GHCR: docker login ghcr.io before running.

.EXAMPLE
  cd <repo-root>
  docker login ghcr.io -u YOUR_USER
  .\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -UseAzureCliForToken

.EXAMPLE
  .\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -ContainerImage 'ghcr.io/sushanth262/aura-frontend:v1.2.3' -UseAzureCliForToken
#>

[CmdletBinding()]
param(
    [string] $ContainerImage = 'ghcr.io/sushanth262/aura-frontend:latest',

    # Path inside the image where nginx serves static files (Dockerfile COPY --from=builder /app/dist).
    [string] $StaticFilesPathInImage = '/usr/share/nginx/html',

    [switch] $SkipDockerPull,

    [ValidateSet('production', 'preview')]
    [string] $Environment = 'production',

    [SecureString] $DeploymentToken,

    [string] $StaticSiteResourceId = '/subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourcegroups/aura/providers/Microsoft.Web/staticSites/Aura',

    [switch] $UseAzureCliForToken,

    # Keep extracted files on disk for inspection (default: temp folder removed after deploy).
    [string] $StagingDirectory
)

$ErrorActionPreference = 'Stop'

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
        Write-Error 'Could not read deployment token from az staticwebapp secrets list. Copy the token from Portal -> Static Web App -> Overview -> Manage deployment token.'
    }

    return $apiKey.Trim()
}

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Error 'Docker CLI not found on PATH. Install Docker Desktop or Docker Engine.'
}

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

Or set AZURE_STATIC_WEB_APPS_API_TOKEN from:
  Portal -> Static Web App '$StaticSiteName' -> Overview -> Manage deployment token

Resource: /subscriptions/$SubscriptionId/resourceGroups/$ResourceGroupName/providers/Microsoft.Web/staticSites/$StaticSiteName
"@
}

$ownStaging = $false
if ([string]::IsNullOrWhiteSpace($StagingDirectory)) {
    $StagingDirectory = Join-Path $env:TEMP ('aura-swa-deploy-' + [guid]::NewGuid().ToString('N'))
    $ownStaging = $true
}

New-Item -ItemType Directory -Path $StagingDirectory -Force | Out-Null

$cid = $null
try {
    if (-not $SkipDockerPull) {
        Write-Host "Pulling $ContainerImage ..."
        docker pull $ContainerImage
        if ($LASTEXITCODE -ne 0) {
            Write-Error "docker pull failed. For private GHCR images run: docker login ghcr.io"
        }
    }

    Write-Host 'Creating ephemeral container to copy static files...'
    $cid = docker create $ContainerImage
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($cid)) {
        Write-Error "docker create failed for image $ContainerImage"
    }
    $cid = $cid.Trim()

    $src = "${cid}:${StaticFilesPathInImage}/."
    Write-Host "docker cp $src -> $StagingDirectory"
    docker cp $src $StagingDirectory
    if ($LASTEXITCODE -ne 0) {
        Write-Error "docker cp failed. Check -StaticFilesPathInImage (default matches Aura Dockerfile nginx html path)."
    }

    $indexHtml = Join-Path $StagingDirectory 'index.html'
    if (-not (Test-Path -LiteralPath $indexHtml)) {
        Write-Error "No index.html under $StagingDirectory. Image layout may differ from Aura nginx image."
    }

    Write-Host "Deploying to Azure Static Web App '$StaticSiteName' ($Environment) from $StagingDirectory ..."

    $env:SWA_CLI_DEPLOYMENT_TOKEN = $token
    try {
        npx --yes '@azure/static-web-apps-cli' deploy $StagingDirectory --env $Environment
    }
    finally {
        Remove-Item Env:SWA_CLI_DEPLOYMENT_TOKEN -ErrorAction SilentlyContinue
    }

    Write-Host "Deploy finished -> resource group '$ResourceGroupName', app '$StaticSiteName'."
}
finally {
    if ($cid) {
        docker rm $cid 2>&1 | Out-Null
    }
    if ($ownStaging) {
        Remove-Item -LiteralPath $StagingDirectory -Recurse -Force -ErrorAction SilentlyContinue
    }
}
