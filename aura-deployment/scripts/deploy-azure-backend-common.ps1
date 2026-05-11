# Shared helpers for Azure Linux Web App (container), Docker Desktop Kubernetes (-Local),
# backend manifest export, and optional Static Web Apps publish — dot-source only.
# Used by deploy-azure-* service scripts and deploy-azure-full-release.ps1.

$script:AuraDeploymentScriptsDir = $PSScriptRoot

function Assert-AzCli {
    if (-not (Get-Command az -ErrorAction SilentlyContinue)) {
        throw 'Azure CLI (az) not found. Install: https://learn.microsoft.com/cli/azure/install-azure-cli-windows'
    }
}

function Assert-AzLoggedIn {
    Assert-AzCli
    az account show --only-show-errors -o none 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Not logged in. Run: az login"
    }
}

function Ensure-ResourceGroup {
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][string]$Location
    )
    $exists = az group exists --name $Name -o tsv
    if ($exists -eq 'false') {
        Write-Host "Creating resource group '$Name' ($Location)..."
        az group create --name $Name --location $Location --only-show-errors
        if ($LASTEXITCODE -ne 0) { throw "az group create failed" }
    }
}

function Ensure-LinuxAppServicePlan {
    param(
        [Parameter(Mandatory)][string]$ResourceGroup,
        [Parameter(Mandatory)][string]$PlanName,
        [Parameter(Mandatory)][string]$Location,
        [string]$Sku = 'FREE'
    )
    az appservice plan show --name $PlanName --resource-group $ResourceGroup -o none --only-show-errors 2>$null
    if ($LASTEXITCODE -eq 0) { return }

    Write-Host "Creating Linux App Service plan '$PlanName' (sku=$Sku)..."
    az appservice plan create `
        --resource-group $ResourceGroup `
        --name $PlanName `
        --location $Location `
        --sku $Sku `
        --is-linux `
        --only-show-errors
    if ($LASTEXITCODE -ne 0) {
        throw @"
App Service plan create failed. FREE SKU custom containers are not available in all subscriptions/regions.
Retry with -AppServiceSku B1 for a low-cost Linux plan that reliably supports containers:
  https://learn.microsoft.com/azure/app-service/quickstart-custom-container
"@
    }
}

function Get-WebAppDefaultHostname {
    param(
        [Parameter(Mandatory)][string]$ResourceGroup,
        [Parameter(Mandatory)][string]$AppName
    )
    $hostName = az webapp show --resource-group $ResourceGroup --name $AppName --query defaultHostName -o tsv --only-show-errors 2>$null
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($hostName)) {
        return $null
    }
    return $hostName.Trim()
}

function Resolve-WebAppHttpsRoot {
    param(
        [Parameter(Mandatory)][string]$ResourceGroup,
        [Parameter(Mandatory)][string]$AppName
    )
    $h = Get-WebAppDefaultHostname -ResourceGroup $ResourceGroup -AppName $AppName
    if (-not $h) { return $null }
    return "https://$h"
}

function Test-AuraHealthEndpoint {
    param(
        [Parameter(Mandatory)][string]$HttpsRoot,
        [int]$TimeoutSec = 15
    )
    # Accepts http:// or https:// (Kubernetes localhost NodePorts use http).
    $base = $HttpsRoot.TrimEnd('/')
    $uri = "$base/healthz"
    try {
        $r = Invoke-WebRequest -Uri $uri -UseBasicParsing -TimeoutSec $TimeoutSec -MaximumRedirection 5
        return ($r.StatusCode -eq 200)
    }
    catch {
        return $false
    }
}

function Wait-AuraHealthEndpoint {
    param(
        [Parameter(Mandatory)][string]$HttpsRoot,
        [string]$Label = 'service',
        [int]$TimeoutSec = 300,
        [int]$IntervalSec = 8
    )
    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    $base = $HttpsRoot.TrimEnd('/')
    $uri = "$base/healthz"
    # Avoid "(...)" inside double quotes — PowerShell treats "(Token)" like a command invocation (e.g. Get-Date).
    Write-Host ('Waiting for {0} health at {1} - timeout {2}s...' -f $Label, $uri, $TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        if (Test-AuraHealthEndpoint -HttpsRoot $HttpsRoot -TimeoutSec 20) {
            Write-Host "OK: $Label healthy."
            return
        }
        Start-Sleep -Seconds $IntervalSec
    }
    throw "Health check timed out for $Label : $uri"
}

function Assert-WebAppExistsOrThrow {
    param(
        [Parameter(Mandatory)][string]$ResourceGroup,
        [Parameter(Mandatory)][string]$AppName,
        [Parameter(Mandatory)][string]$DependencyLabel
    )
    $h = Get-WebAppDefaultHostname -ResourceGroup $ResourceGroup -AppName $AppName
    if (-not $h) {
        throw "Dependency '$DependencyLabel' Web App not found: '$AppName' in resource group '$ResourceGroup'. Deploy it first."
    }
}

function Assert-PeerHealthyOrThrow {
    param(
        [Parameter(Mandatory)][string]$ResourceGroup,
        [Parameter(Mandatory)][string]$AppName,
        [Parameter(Mandatory)][string]$DependencyLabel,
        [switch]$SkipHealthCheck
    )
    Assert-WebAppExistsOrThrow -ResourceGroup $ResourceGroup -AppName $AppName -DependencyLabel $DependencyLabel
    $root = Resolve-WebAppHttpsRoot -ResourceGroup $ResourceGroup -AppName $AppName
    if ($SkipHealthCheck) {
        Write-Host "[SkipHealthCheck] Assuming $DependencyLabel reachable at $root"
        return $root
    }
    if (-not (Test-AuraHealthEndpoint -HttpsRoot $root)) {
        throw (
            "$DependencyLabel Web App '$AppName' exists but /healthz did not return HTTP 200 yet.`n" +
            "Deploy order: worker -> supervisor -> authz -> BFF. Wait for cold start or run without -SkipDependencyChecks after worker is warm.`n" +
            "Endpoint tried: $root/healthz"
        )
    }
    Write-Host "Dependency OK: $DependencyLabel at $root"
    return $root
}

function Deploy-AuraLinuxContainerWebApp {
    param(
        [Parameter(Mandatory)][string]$ResourceGroup,
        [Parameter(Mandatory)][string]$Location,
        [Parameter(Mandatory)][string]$PlanName,
        [Parameter(Mandatory)][string]$AppName,
        [Parameter(Mandatory)][string]$ContainerImage,
        [Parameter(Mandatory)][string]$GhcrUser,
        [Parameter(Mandatory)][string]$GhcrToken,
        [Parameter(Mandatory)][int]$WebSitesPort,
        [hashtable]$ExtraAppSettings = @{},
        [string]$Sku = 'FREE'
    )

    Ensure-ResourceGroup -Name $ResourceGroup -Location $Location
    Ensure-LinuxAppServicePlan -ResourceGroup $ResourceGroup -PlanName $PlanName -Location $Location -Sku $Sku

    $exists = $false
    az webapp show --resource-group $ResourceGroup --name $AppName -o none --only-show-errors 2>$null
    if ($LASTEXITCODE -eq 0) { $exists = $true }

    if (-not $exists) {
        Write-Host "Creating Web App '$AppName' with image $ContainerImage ..."
        az webapp create `
            --resource-group $ResourceGroup `
            --plan $PlanName `
            --name $AppName `
            --deployment-container-image-name $ContainerImage `
            --only-show-errors
        if ($LASTEXITCODE -ne 0) {
            throw "az webapp create failed for '$AppName'"
        }
    }
    else {
        Write-Host "Updating container image for existing Web App '$AppName' ..."
    }

    az webapp config container set `
        --resource-group $ResourceGroup `
        --name $AppName `
        --docker-custom-image-name $ContainerImage `
        --docker-registry-server-url https://ghcr.io `
        --docker-registry-server-user $GhcrUser `
        --docker-registry-server-password $GhcrToken `
        --only-show-errors
    if ($LASTEXITCODE -ne 0) {
        throw "az webapp config container set failed for '$AppName'"
    }

    # Always-on not available on FREE; WEBSITES_PORT tells Azure which container port to route.
    $pairs = @(
        "WEBSITES_PORT=$WebSitesPort",
        "WEBSITES_ENABLE_APP_SERVICE_STORAGE=false"
    )
    foreach ($k in $ExtraAppSettings.Keys) {
        $pairs += "$($k)=$($ExtraAppSettings[$k])"
    }

    # Avoid `@pairs` splat ambiguity — az expects multiple KEY=VAL tokens after --settings.
    $azSettingsArgs = @(
        'webapp', 'config', 'appsettings', 'set',
        '--resource-group', $ResourceGroup,
        '--name', $AppName,
        '--only-show-errors',
        '--settings'
    ) + $pairs
    & az @azSettingsArgs
    if ($LASTEXITCODE -ne 0) {
        throw "az webapp config appsettings set failed for '$AppName'"
    }

    Write-Host "Restarting '$AppName'..."
    az webapp restart --resource-group $ResourceGroup --name $AppName --only-show-errors
    if ($LASTEXITCODE -ne 0) {
        throw "az webapp restart failed for '$AppName'"
    }

    $fqdn = Get-WebAppDefaultHostname -ResourceGroup $ResourceGroup -AppName $AppName
    return $fqdn
}

function Resolve-GhcrCredentials {
    param(
        [string]$GhcrUser,
        [string]$GhcrToken
    )
    $u = $GhcrUser
    if ([string]::IsNullOrWhiteSpace($u)) {
        $u = $env:GHCR_USERNAME
    }
    if ([string]::IsNullOrWhiteSpace($u)) {
        $u = $env:GITHUB_ACTOR
    }
    $t = $GhcrToken
    if ([string]::IsNullOrWhiteSpace($t)) {
        $t = $env:GITHUB_TOKEN
    }
    if ([string]::IsNullOrWhiteSpace($t)) {
        $t = $env:GHCR_TOKEN
    }
    if ([string]::IsNullOrWhiteSpace($u) -or [string]::IsNullOrWhiteSpace($t)) {
        throw @"
GHCR pull credentials missing. Pass -GhcrUser and -GhcrToken, or set environment variables:
  GHCR_USERNAME (or GITHUB_ACTOR) and GITHUB_TOKEN / GHCR_TOKEN (PAT with read:packages).
"@
    }
    return @{ User = $u.Trim(); Token = $t.Trim() }
}

function Resolve-AuraPackageImageTag {
    <#
      Canonical container tag for GHCR (Docker image tag). Prefer -PackageVersion in scripts;
      -ImageTag is accepted as an alias for backward compatibility.
    #>
    param(
        [string]$PackageVersion,
        [string]$ImageTag
    )
    $pv = if ($null -ne $PackageVersion) { $PackageVersion.Trim() } else { '' }
    $it = if ($null -ne $ImageTag) { $ImageTag.Trim() } else { '' }
    if ($pv -ne '' -and $it -ne '' -and $pv -ne $it) {
        throw 'Use only one of -PackageVersion or -ImageTag, or pass the same value for both.'
    }
    if ($pv -ne '') { return $pv }
    if ($it -ne '') { return $it }
    return 'latest'
}

function Get-AuraSafeReleaseFileStem {
    param([Parameter(Mandatory)][string]$PackageVersion)
    $s = $PackageVersion.Trim()
    if ([string]::IsNullOrWhiteSpace($s)) { return 'latest' }
    return ($s -replace '[\\/:*?"<>|\s]', '-').Trim('-')
}

function Get-AuraStandardAppNames {
    # Azure Web App names: alphanumeric + hyphens; must be globally unique on azurewebsites.net.
    param(
        [Parameter(Mandatory)][string]$UniqueSuffix
    )
    $s = ($UniqueSuffix.Trim().ToLower() -replace '[^a-z0-9\-]', '')
    if ([string]::IsNullOrWhiteSpace($s)) {
        throw '-UniqueSuffix must contain letters/digits (e.g. jdoe01).'
    }
    if ($s.Length -gt 20) {
        $s = $s.Substring(0, 20)
    }
    return @{
        Worker     = "aura-$s-worker"
        Supervisor = "aura-$s-supervisor"
        Authz      = "aura-$s-authz"
        Bff        = "aura-$s-bff"
    }
}

function Write-AuraStackSummaryFile {
    param(
        [Parameter(Mandatory)][string]$Path,
        [Parameter(Mandatory)][hashtable]$Endpoints,
        [string]$PackageVersion,
        [string]$GhcrOwner,
        [string]$UniqueSuffix
    )
    $bffRaw = $Endpoints['bff_https']
    if ([string]::IsNullOrWhiteSpace($bffRaw)) {
        throw 'Write-AuraStackSummaryFile: endpoints must include bff_https.'
    }
    $bff = ($bffRaw.Trim()).TrimEnd('/')
    $apiBase = ''
    $wsBase = ''
    if ($bff -match '^(?i)https://(.+)$') {
        $apiBase = "$bff/v1"
        $wsBase = ('wss://' + $Matches[1]).TrimEnd('/')
    }
    elseif ($bff -match '^(?i)http://(.+)$') {
        $apiBase = "$bff/v1"
        $wsBase = ('ws://' + $Matches[1]).TrimEnd('/')
    }
    else {
        throw "Write-AuraStackSummaryFile: bff_https must start with http:// or https:// (got: $bffRaw)."
    }
    $obj = [ordered]@{
        generated_at_utc = (Get-Date).ToUniversalTime().ToString('o')
        endpoints        = $Endpoints
        expo_hints       = @{
            EXPO_PUBLIC_API_BASE_URL = $apiBase
            EXPO_PUBLIC_WS_BASE_URL  = $wsBase
        }
    }
    $pv = if ($null -ne $PackageVersion) { $PackageVersion.Trim() } else { '' }
    $owner = if ($null -ne $GhcrOwner) { $GhcrOwner.Trim() } else { '' }
    if ($pv -ne '' -and $owner -ne '') {
        $obj.package_version = $pv
        $obj.ghcr_owner      = $owner
        if (-not [string]::IsNullOrWhiteSpace($UniqueSuffix)) {
            $obj.unique_suffix = $UniqueSuffix.Trim()
        }
        $obj.images = [ordered]@{
            'aura-worker'     = "ghcr.io/$owner/aura-worker:$pv"
            'aura-supervisor' = "ghcr.io/$owner/aura-supervisor:$pv"
            'aura-authz'      = "ghcr.io/$owner/aura-authz:$pv"
            'aura-bff-api'    = "ghcr.io/$owner/aura-bff-api:$pv"
            'aura-frontend'   = "ghcr.io/$owner/aura-frontend:$pv"
        }
        $obj.terraform_hints = [ordered]@{
            image_tag                         = $pv
            frontend_expo_public_api_base_url = $apiBase
            frontend_expo_public_ws_base_url  = $wsBase
        }
        $obj.notes = @(
            'Backend Web Apps use images listed under images.*',
            'Rebuild/push aura-frontend with Terraform using image_tag and frontend_expo_public_* matching terraform_hints before publishing Static Web Apps.'
        )
    }
    $json = $obj | ConvertTo-Json -Depth 8
    Set-Content -LiteralPath $Path -Value $json -Encoding UTF8
    Write-Host "Wrote stack summary / release manifest: $Path"
}

#region Kubernetes (Docker Desktop) — used when deploy-azure-* scripts pass -Local

function Escape-YamlSingleQuotedLiteral {
    param([Parameter(Mandatory)][string]$Value)
    return ($Value -replace "'", "''")
}

function Assert-KubectlCli {
    if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
        throw 'kubectl not found on PATH. Enable Kubernetes in Docker Desktop and ensure kubectl is available.'
    }
}

function Warn-KubectlDockerDesktopContext {
    $ctx = (& kubectl config current-context 2>$null | Out-String).Trim()
    if ([string]::IsNullOrWhiteSpace($ctx)) { return }
    if ($ctx -notmatch 'docker-desktop') {
        Write-Warning "kubectl context is '$ctx'. For Docker Desktop Kubernetes use a context named like 'docker-desktop'."
    }
}

function Get-AuraKubernetesNamespaceFromSuffix {
    param([Parameter(Mandatory)][string]$UniqueSuffix)
    $s = ($UniqueSuffix.Trim().ToLower() -replace '[^a-z0-9\-]', '')
    if ([string]::IsNullOrWhiteSpace($s)) {
        throw '-UniqueSuffix must contain letters/digits for Kubernetes namespace (e.g. jdoe01).'
    }
    if ($s.Length -gt 50) {
        $s = $s.Substring(0, 50).TrimEnd('-')
    }
    return "aura-$s"
}

function Get-AuraK8sDockerDesktopManifestDir {
    $p = Join-Path $script:AuraDeploymentScriptsDir '..'
    $p = Join-Path $p 'k8s'
    $p = Join-Path $p 'docker-desktop'
    return (Resolve-Path -LiteralPath $p -ErrorAction Stop).Path
}

function Expand-AuraK8sManifestPlaceholders {
    param(
        [Parameter(Mandatory)][string]$YamlContent,
        [Parameter(Mandatory)][hashtable]$Replacements
    )
    $out = $YamlContent
    foreach ($key in $Replacements.Keys) {
        $out = $out.Replace([string]$key, [string]$Replacements[$key])
    }
    return $out
}

function Invoke-AuraKubectlApplyYamlString {
    param([Parameter(Mandatory)][string]$YamlContent)
    $tmp = [System.IO.Path]::ChangeExtension([System.IO.Path]::GetTempFileName(), '.yaml')
    try {
        Set-Content -LiteralPath $tmp -Value $YamlContent -Encoding UTF8
        & kubectl apply -f $tmp
        if ($LASTEXITCODE -ne 0) { throw 'kubectl apply failed.' }
    }
    finally {
        Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
    }
}

function Ensure-AuraKubernetesNamespace {
    param([Parameter(Mandatory)][string]$Namespace)
    & kubectl create namespace $Namespace --dry-run=client -o yaml | kubectl apply -f -
    if ($LASTEXITCODE -ne 0) { throw "kubectl create/apply namespace failed: $Namespace" }
}

function Ensure-AuraK8sGhcrPullSecret {
    param(
        [Parameter(Mandatory)][string]$Namespace,
        [Parameter(Mandatory)][string]$DockerUsername,
        [Parameter(Mandatory)][string]$DockerPassword
    )
    & kubectl create secret docker-registry ghcr-pull `
        --docker-server='https://ghcr.io' `
        --docker-username=$DockerUsername `
        --docker-password=$DockerPassword `
        -n $Namespace `
        --dry-run=client -o yaml | kubectl apply -f -
    if ($LASTEXITCODE -ne 0) { throw 'kubectl apply ghcr-pull secret failed.' }
}

function Resolve-AuraK8sImagePullSecretsYamlBlock {
    param([switch]$SkipGhcrImagePullSecret)
    if ($SkipGhcrImagePullSecret) { return '' }
    return @'
      imagePullSecrets:
        - name: ghcr-pull

'@
}

function Wait-AuraKubernetesRollout {
    param(
        [Parameter(Mandatory)][string]$Namespace,
        [Parameter(Mandatory)][string]$DeploymentName,
        [int]$TimeoutSec = 240
    )
    & kubectl rollout status "deployment/$DeploymentName" -n $Namespace "--timeout=${TimeoutSec}s"
    if ($LASTEXITCODE -ne 0) {
        throw "kubectl rollout status failed for $DeploymentName in namespace $Namespace"
    }
}

function Assert-HttpDependencyHealthyOrThrow {
    param(
        [Parameter(Mandatory)][string]$RootUrl,
        [Parameter(Mandatory)][string]$DependencyLabel,
        [switch]$SkipHealthCheck
    )
    $root = $RootUrl.TrimEnd('/')
    if ($SkipHealthCheck) {
        Write-Host "[SkipDependencyChecks] Assuming $DependencyLabel reachable at $root"
        return $root
    }
    if (-not (Test-AuraHealthEndpoint -HttpsRoot $root)) {
        throw @"
$DependencyLabel did not return HTTP 200 at $root/healthz.
Deploy order on Kubernetes: worker -> supervisor -> authz -> BFF.
"@
    }
    Write-Host "Dependency OK: $DependencyLabel at $root"
    return $root
}

function Deploy-AuraK8sServiceFromDockerDesktopTemplate {
    param(
        [Parameter(Mandatory)][string]$TemplateFileName,
        [Parameter(Mandatory)][hashtable]$Replacements
    )
    $dir = Get-AuraK8sDockerDesktopManifestDir
    $path = Join-Path $dir $TemplateFileName
    if (-not (Test-Path -LiteralPath $path)) { throw "Kubernetes template missing: $path" }
    $raw = Get-Content -LiteralPath $path -Raw -Encoding UTF8
    $yaml = Expand-AuraK8sManifestPlaceholders -YamlContent $raw -Replacements $Replacements
    Invoke-AuraKubectlApplyYamlString -YamlContent $yaml
}

#endregion

#region Azure backend manifest (was export-azure-backend-endpoints.ps1)

function Export-AuraAzureBackendEndpointsManifest {
    param(
        [Parameter(Mandatory)][string]$UniqueSuffix,
        [Parameter(Mandatory)][string]$ResourceGroupName,
        [Parameter(Mandatory)][string]$OutputPath,
        [string]$SubscriptionId,
        [string]$PackageVersion,
        [string]$GhcrOwner = 'sushanth262'
    )
    Assert-AzLoggedIn
    if ($SubscriptionId) {
        az account set --subscription $SubscriptionId --only-show-errors | Out-Null
        if ($LASTEXITCODE -ne 0) { throw 'az account set failed' }
    }
    $names = Get-AuraStandardAppNames -UniqueSuffix $UniqueSuffix
    $endpoints = @{
        worker_https     = (Resolve-WebAppHttpsRoot -ResourceGroup $ResourceGroupName -AppName $names.Worker)
        supervisor_https = (Resolve-WebAppHttpsRoot -ResourceGroup $ResourceGroupName -AppName $names.Supervisor)
        authz_https      = (Resolve-WebAppHttpsRoot -ResourceGroup $ResourceGroupName -AppName $names.Authz)
        bff_https        = (Resolve-WebAppHttpsRoot -ResourceGroup $ResourceGroupName -AppName $names.Bff)
    }
    foreach ($p in @(
            @{ Key = 'worker_https'; App = $names.Worker }
            @{ Key = 'supervisor_https'; App = $names.Supervisor }
            @{ Key = 'authz_https'; App = $names.Authz }
            @{ Key = 'bff_https'; App = $names.Bff }
        )) {
        if ([string]::IsNullOrWhiteSpace($endpoints[$p.Key])) {
            throw "Could not resolve $($p.Key) for Web App '$($p.App)' in resource group '$ResourceGroupName'."
        }
    }
    Write-AuraStackSummaryFile `
        -Path $OutputPath `
        -Endpoints $endpoints `
        -PackageVersion $PackageVersion `
        -GhcrOwner $GhcrOwner `
        -UniqueSuffix $UniqueSuffix
}

#endregion

#region Local Kubernetes release manifest (localhost NodePorts; single stack per cluster)

function Export-AuraLocalKubernetesEndpointsManifest {
    param(
        [Parameter(Mandatory)][string]$KubernetesNamespace,
        [Parameter(Mandatory)][string]$UniqueSuffix,
        [Parameter(Mandatory)][string]$OutputPath,
        [string]$PackageVersion,
        [string]$GhcrOwner = 'sushanth262'
    )
    $endpoints = @{
        worker_https       = 'http://127.0.0.1:30883'
        supervisor_https   = 'http://127.0.0.1:30882'
        authz_https        = 'http://127.0.0.1:30881'
        bff_https          = 'http://127.0.0.1:30880'
        frontend_https     = 'http://127.0.0.1:30900'
    }
    Write-AuraStackSummaryFile `
        -Path $OutputPath `
        -Endpoints $endpoints `
        -PackageVersion $PackageVersion `
        -GhcrOwner $GhcrOwner `
        -UniqueSuffix $UniqueSuffix
    $extraNote = "Kubernetes namespace: $KubernetesNamespace - NodePorts 30880-30883 (APIs) and 30900 (frontend) are cluster-wide; run only one Aura stack per cluster."
    Write-Host $extraNote
}

#endregion

#region Azure Static Web Apps publish (was deploy-azure-static-web-app.ps1)

function Get-AuraPlainTokenFromSecureString {
    param([SecureString]$Secure)
    if (-not $Secure) { return $null }
    $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($Secure)
    try { [System.Runtime.InteropServices.Marshal]::PtrToStringUni($bstr) }
    finally { [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr) }
}

function Resolve-AuraStaticSiteFromResourceId {
    param([string]$ResourceId)
    $m = [regex]::Match(
        $ResourceId.Trim(),
        '(?i)^/subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft\.Web/staticSites/([^/]+)$'
    )
    if (-not $m.Success) {
        throw "StaticSiteResourceId must look like: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/staticSites/{name}"
    }
    @{
        SubscriptionId     = $m.Groups[1].Value
        ResourceGroupName = $m.Groups[2].Value
        StaticSiteName    = $m.Groups[3].Value
    }
}

function Read-AuraBackendStackSummaryJsonFile {
    param([Parameter(Mandatory)][string]$Path)
    if (-not (Test-Path -LiteralPath $Path)) {
        throw "Backend stack summary not found: $Path"
    }
    Get-Content -LiteralPath $Path -Raw -Encoding UTF8 | ConvertFrom-Json
}

function Resolve-AuraBffHintsForFrontendDeploy {
    param(
        [string]$BackendStackSummaryJson,
        [string]$BffHttpsRoot,
        [string]$ExpectedApiBaseUrl,
        [string]$ExpectedWsBaseUrl
    )
    $api = ''
    $ws = ''
    $root = ''

    if (-not [string]::IsNullOrWhiteSpace($BackendStackSummaryJson)) {
        $o = Read-AuraBackendStackSummaryJsonFile -Path $BackendStackSummaryJson
        if ($null -ne $o.endpoints -and $null -ne $o.endpoints.bff_https) {
            $root = ([string]$o.endpoints.bff_https).TrimEnd('/')
        }
        if ($null -ne $o.expo_hints) {
            $eh = $o.expo_hints
            if ($eh.EXPO_PUBLIC_API_BASE_URL) { $api = [string]$eh.EXPO_PUBLIC_API_BASE_URL }
            if ($eh.EXPO_PUBLIC_WS_BASE_URL) { $ws = [string]$eh.EXPO_PUBLIC_WS_BASE_URL }
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($BffHttpsRoot)) {
        $root = $BffHttpsRoot.TrimEnd('/')
    }
    if (-not [string]::IsNullOrWhiteSpace($ExpectedApiBaseUrl)) {
        $api = $ExpectedApiBaseUrl.Trim()
    }
    if (-not [string]::IsNullOrWhiteSpace($ExpectedWsBaseUrl)) {
        $ws = $ExpectedWsBaseUrl.Trim()
    }

    if (-not [string]::IsNullOrWhiteSpace($root)) {
        if ([string]::IsNullOrWhiteSpace($api)) { $api = "$root/v1" }
        if ([string]::IsNullOrWhiteSpace($ws)) {
            if ($root -match '(?i)^https://') {
                $ws = ($root -replace '(?i)^https://', 'wss://').TrimEnd('/')
            }
            elseif ($root -match '(?i)^http://') {
                $ws = ($root -replace '(?i)^http://', 'ws://').TrimEnd('/')
            }
        }
    }
    elseif (-not [string]::IsNullOrWhiteSpace($api)) {
        try {
            $uri = [Uri]$api
            $root = $uri.GetLeftPart([UriPartial]::Authority)
        }
        catch {
            $root = ''
        }
    }

    return @{ ApiBase = $api; WsBase = $ws; BffRoot = $root }
}

function Test-AuraBffRootHealthProbe {
    param([Parameter(Mandatory)][string]$BffRoot)
    $u = "$(($BffRoot).TrimEnd('/'))/healthz"
    try {
        $r = Invoke-WebRequest -Uri $u -UseBasicParsing -TimeoutSec 25 -MaximumRedirection 5
        return ($r.StatusCode -eq 200)
    }
    catch {
        return $false
    }
}

function Assert-AuraFrontendBundleTargetsBffHost {
    param(
        [Parameter(Mandatory)][string]$StagingDirectory,
        [string]$ExpectedHost,
        [switch]$RejectLocalhost8080
    )
    $js = @(Get-ChildItem -LiteralPath $StagingDirectory -Filter '*.js' -Recurse -File -ErrorAction SilentlyContinue)
    if ($js.Count -eq 0) { return }

    if ($RejectLocalhost8080) {
        foreach ($f in $js) {
            $c = Get-Content -LiteralPath $f.FullName -Raw -ErrorAction SilentlyContinue
            if ($null -ne $c -and $c.Contains('localhost:8080')) {
                throw "Bundled script '$($f.Name)' still references localhost:8080. Rebuild aura-frontend with Terraform frontend_expo_public_* matching your deployed BFF."
            }
        }
    }

    if ([string]::IsNullOrWhiteSpace($ExpectedHost)) { return }

    $found = $false
    foreach ($f in $js) {
        $c = Get-Content -LiteralPath $f.FullName -Raw -ErrorAction SilentlyContinue
        if ($null -ne $c -and $c.Contains($ExpectedHost)) {
            $found = $true
            break
        }
    }
    if (-not $found) {
        Write-Warning "No *.js under staging contained host '$ExpectedHost'. If the SPA fails in the browser, rebuild the frontend image with EXPO_PUBLIC_* aimed at this BFF."
    }
}

function Get-AuraStaticWebAppDeployTokenFromAzCli {
    param(
        [string]$SubscriptionId,
        [string]$ResourceGroupName,
        [string]$StaticSiteName
    )
    Assert-AzCli
    Write-Host "Using Azure CLI for Static Web Apps deployment token (subscription=$SubscriptionId, rg=$ResourceGroupName, site=$StaticSiteName)..."
    az account set --subscription $SubscriptionId --only-show-errors 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "az account set failed. Run 'az login' for subscription $SubscriptionId"
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
            throw "az staticwebapp secrets list failed: $jsonText"
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
        throw 'Could not read deployment token from az staticwebapp secrets list. Copy the token from Portal -> Static Web App -> Manage deployment token.'
    }

    return $apiKey.Trim()
}

function Publish-AuraAzureStaticWebApp {
    [CmdletBinding()]
    param(
        [string]$ContainerImage = 'ghcr.io/sushanth262/aura-frontend:latest',
        [string]$StaticFilesPathInImage = '/usr/share/nginx/html',
        [switch]$SkipDockerPull,
        [ValidateSet('production', 'preview')]
        [string]$Environment = 'production',
        [SecureString]$DeploymentToken,
        [string]$StaticSiteResourceId = '/subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourcegroups/aura/providers/Microsoft.Web/staticSites/Aura',
        [switch]$UseAzureCliForToken,
        [string]$StagingDirectory,
        [string]$BackendStackSummaryJson,
        [string]$BffHttpsRoot,
        [string]$ExpectedApiBaseUrl,
        [string]$ExpectedWsBaseUrl,
        [switch]$SkipBffConnectivityCheck,
        [switch]$StrictBffBundleMatch
    )

    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw 'Docker CLI not found on PATH.'
    }

    $ctx = Resolve-AuraStaticSiteFromResourceId -ResourceId $StaticSiteResourceId
    $SubscriptionId = $ctx.SubscriptionId
    $RgName = $ctx.ResourceGroupName
    $StaticSiteName = $ctx.StaticSiteName

    $token = $null
    if ($UseAzureCliForToken) {
        $token = Get-AuraStaticWebAppDeployTokenFromAzCli `
            -SubscriptionId $SubscriptionId `
            -ResourceGroupName $RgName `
            -StaticSiteName $StaticSiteName
    }

    if (-not $token) {
        $token = Get-AuraPlainTokenFromSecureString -Secure $DeploymentToken
    }
    if (-not $token) {
        $token = $env:AZURE_STATIC_WEB_APPS_API_TOKEN
    }
    if (-not $token) {
        $token = $env:SWA_CLI_DEPLOYMENT_TOKEN
    }

    if (-not $token) {
        throw @"
No Static Web Apps deployment token. Use -UseAzureCliForToken after az login, or set AZURE_STATIC_WEB_APPS_API_TOKEN from Portal -> Manage deployment token.
Resource: /subscriptions/$SubscriptionId/resourceGroups/$RgName/providers/Microsoft.Web/staticSites/$StaticSiteName
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
                throw "docker pull failed. For private GHCR run: docker login ghcr.io"
            }
        }

        Write-Host 'Creating ephemeral container to copy static files...'
        $cid = docker create $ContainerImage
        if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($cid)) {
            throw "docker create failed for image $ContainerImage"
        }
        $cid = $cid.Trim()

        $src = "${cid}:${StaticFilesPathInImage}/."
        Write-Host "docker cp $src -> $StagingDirectory"
        docker cp $src $StagingDirectory
        if ($LASTEXITCODE -ne 0) {
            throw "docker cp failed. Check -StaticFilesPathInImage (Aura Dockerfile uses /usr/share/nginx/html)."
        }

        $indexHtml = Join-Path $StagingDirectory 'index.html'
        if (-not (Test-Path -LiteralPath $indexHtml)) {
            throw "No index.html under $StagingDirectory."
        }

        $bffHints = Resolve-AuraBffHintsForFrontendDeploy `
            -BackendStackSummaryJson $BackendStackSummaryJson `
            -BffHttpsRoot $BffHttpsRoot `
            -ExpectedApiBaseUrl $ExpectedApiBaseUrl `
            -ExpectedWsBaseUrl $ExpectedWsBaseUrl

        if (-not [string]::IsNullOrWhiteSpace($bffHints.BffRoot)) {
            Write-Host ''
            Write-Host 'Frontend talks to aura-bff-api — expected bundle targets:' -ForegroundColor Cyan
            Write-Host "  BFF origin    : $($bffHints.BffRoot)"
            Write-Host "  API base /v1  : $($bffHints.ApiBase)"
            Write-Host "  WebSocket     : $($bffHints.WsBase)"
            if (-not $SkipBffConnectivityCheck) {
                if (-not (Test-AuraBffRootHealthProbe -BffRoot $bffHints.BffRoot)) {
                    throw "BFF not reachable at $($bffHints.BffRoot.TrimEnd('/'))/healthz. Deploy backends first or pass -SkipBffConnectivityCheck."
                }
                Write-Host 'BFF /healthz reachable.' -ForegroundColor Green
            }
        }
        elseif (-not $SkipBffConnectivityCheck) {
            Write-Warning "No BFF URL supplied. Confirm $ContainerImage was built with production EXPO_PUBLIC_* or the SPA may keep localhost:8080."
        }

        if ($StrictBffBundleMatch) {
            $hostPart = ''
            if (-not [string]::IsNullOrWhiteSpace($bffHints.ApiBase)) {
                try { $hostPart = ([Uri]$bffHints.ApiBase).Host } catch { }
            }
            $rejectLocal = ([string]::IsNullOrWhiteSpace($hostPart)) -or ($hostPart -ne 'localhost')
            Assert-AuraFrontendBundleTargetsBffHost `
                -StagingDirectory $StagingDirectory `
                -ExpectedHost $hostPart `
                -RejectLocalhost8080:$rejectLocal
        }

        Write-Host "Deploying to Azure Static Web App '$StaticSiteName' ($Environment) from $StagingDirectory ..."

        $env:SWA_CLI_DEPLOYMENT_TOKEN = $token
        try {
            npx --yes '@azure/static-web-apps-cli' deploy $StagingDirectory --env $Environment
        }
        finally {
            Remove-Item Env:SWA_CLI_DEPLOYMENT_TOKEN -ErrorAction SilentlyContinue
        }

        Write-Host "Deploy finished -> resource group '$RgName', app '$StaticSiteName'."
    }
    finally {
        if ($cid) {
            docker rm $cid 2>&1 | Out-Null
        }
        if ($ownStaging) {
            Remove-Item -LiteralPath $StagingDirectory -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

#endregion
