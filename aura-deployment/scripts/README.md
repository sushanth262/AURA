# Scripts

PowerShell helpers for **Aura**: GHCR images (via Terraform), **Azure App Service** backends, optional **Docker Desktop Kubernetes** (`-Local`), and **Azure Static Web Apps** (via **`deploy-azure-full-release.ps1 -PublishStaticWebApp`** or **`Publish-AuraAzureStaticWebApp`** in **`deploy-azure-backend-common.ps1`**).

Broader context (Terraform variables, outputs): **[`../README.md`](../README.md)**. API/env details: **[`../../docs/BFF_AUTH_LOGIN.md`](../../docs/BFF_AUTH_LOGIN.md)**.

---

## Credentials (don’t confuse these)

| Credential | How you set it | Used for |
|------------|----------------|----------|
| **`docker login ghcr.io -u GITHUB_USER`** (password = GitHub PAT) | Interactive; Docker stores it for **Docker CLI** only | **`docker pull` / `docker push`** on your machine (e.g. SWA script pulling `aura-frontend`). Does **not** set `$env:GITHUB_TOKEN` or Terraform. |
| **`TF_VAR_github_token`** or `github_token` in `terraform.tfvars` | Environment variable or tfvars file | **Terraform** building/pushing images to GHCR (**needs `write:packages`**). |
| **`$env:GITHUB_TOKEN`** (and optionally **`GHCR_USERNAME`**) | PowerShell before each backend deploy script | **Azure:** **`az webapp config container set`** pulls private GHCR (**`read:packages`**). **Kubernetes (-Local):** creates **`ghcr-pull`** pull secret (unless **`-SkipGhcrImagePullSecret`**). |
| **`az login`** | Azure CLI | Azure subscription access (Web Apps; **`deploy-azure-full-release -PublishStaticWebApp -UseAzureCliForToken`** token fetch). |
| **kubectl** | Docker Desktop → Enable Kubernetes | **`-Local`** applies manifests under **`aura-deployment/k8s/docker-desktop/`**. |
| **SWA deployment token** | **`-UseAzureCliForToken`** on **`deploy-azure-full-release.ps1`**, or **`AZURE_STATIC_WEB_APPS_API_TOKEN`** | Upload step inside **`Publish-AuraAzureStaticWebApp`** — not GHCR. |

You can use **one PAT** with **`read:packages`** + **`write:packages`** for everything, but you must **set each mechanism separately** (Docker login does not populate `GITHUB_TOKEN`).

---

## Deployment order (production)

Do these **in order**. Paths assume **repository root**; from `aura-deployment` use `.\scripts\...` instead of `.\aura-deployment\scripts\...`.

### 1 — Push images to GHCR (Terraform)

From **`aura-deployment/`**:

1. Copy **`terraform.tfvars.example`** → **`terraform.tfvars`**, set **`ghcr_owner`**, **`github_token`** (or **`TF_VAR_github_token`**).
2. Optional first pass: you may leave **`frontend_expo_public_*`** empty until you know the BFF URL (step 2).
3. Run:

```powershell
cd aura-deployment
terraform init
terraform apply
```

Outputs include **`aura_*_image`** and **`frontend_image`**.

### 2 — Deploy backends (one script per service, dependency order)

**Azure (default):** **`az login`**, **`$env:GITHUB_TOKEN`** (`read:packages`), images from step 1.

**Docker Desktop Kubernetes:** pass **`-Local`** on each script (or use **`deploy-azure-full-release.ps1 -Local`**, which also deploys **`deploy-azure-frontend.ps1`** unless **`-SkipLocalFrontend`**). Expects **`kubectl`** pointed at Docker Desktop’s cluster (context name usually **`docker-desktop`**). Namespace: **`aura-{UniqueSuffix}`**. Fixed NodePorts: **30880** BFF, **30881–30883** APIs, **30900** static UI (**`http://127.0.0.1:30900`**) — **one Aura namespace per cluster**.

Shared helpers (**dot-source only**, not run directly): **`deploy-azure-backend-common.ps1`**.

#### Dependency order

Runtime edges: **supervisor → worker** (`WORKER_URL`); **BFF → authz** (`AUTHZ_URL`) and **BFF → supervisor** (`SUPERVISOR_URL`). **Authz** does not call other Aura apps.

| # | Script | Depends on (must exist / healthy¹) | Why |
|---|--------|-------------------------------------|-----|
| **1** | **`deploy-azure-worker.ps1`** | Terraform-built **`aura-worker`** image on GHCR | Mock connectors only; no other Aura Web App required. |
| **2** | **`deploy-azure-supervisor.ps1`** | **(1)** Worker Web App — script resolves HTTPS URL and sets **`WORKER_URL`** | Orchestration + WS + internal REST; calls worker over the network. |
| **3** | **`deploy-azure-authz.ps1`** | Terraform **`aura-authz`** image only | Policy API only; **no** dependency on worker or supervisor. |
| **4** | **`deploy-azure-bff-api.ps1`** | **(2)** Supervisor + **(3)** Authz — script resolves both URLs; optional **`/healthz`** checks | Public API + JWT + WS proxy to supervisor + AuthZ evaluate calls. |
| **5** | **`deploy-azure-frontend.ps1`** (Kubernetes `-Local` only) | **(4)** BFF healthy — include **`http://127.0.0.1:30900`** in **`CORS_ALLOWED_ORIGINS`** when deploying the BFF³ | Serves **`ghcr.io/…/aura-frontend`** via nginx on NodePort **30900**. |

¹ **Azure:** **`az webapp show`** + **`GET …/healthz`**. **`-Local`:** **`GET http://127.0.0.1:{NodePort}/healthz`** (frontend nginx serves **`/healthz`** after you rebuild/push the image from current Dockerfile). Omit checks with **`-SkipDependencyChecks`** / **`-SkipHealthWait`**.

² After **(1)**, **supervisor** and **authz** are **independent** — either order is valid. This guide uses **worker → supervisor → authz → BFF → frontend**.

³ SPA bundle must call the **BFF** (**`EXPO_PUBLIC_*`** baked at image build): local K8s use **`http://127.0.0.1:30880/v1`** and **`ws://127.0.0.1:30880`**. Set **`frontend_expo_public_*`** in **`terraform.tfvars`** and **`terraform apply`** so **`aura-frontend`** on GHCR matches this BFF before **`deploy-azure-frontend.ps1`**. **`deploy-azure-full-release.ps1 -Local`** appends **`http://127.0.0.1:30900`** to BFF CORS automatically; manual step-by-step must pass that origin on **`deploy-azure-bff-api.ps1`**.

**Package version (GHCR tag):** each service script accepts **`-PackageVersion`** — the Docker tag Terraform pushed for **all** Aura images in this release (same value as Terraform **`image_tag`** when you align versions). **`-ImageTag`** is a legacy alias; omit **both** to deploy **`latest`**.

Example (**same splat** on each call — adjust **`UniqueSuffix`** / **`ResourceGroupName`** / **`PackageVersion`**):

```powershell
$env:GITHUB_TOKEN = '<pat read:packages>'
$s = @{ UniqueSuffix='jdoe01'; ResourceGroupName='aura-rg'; Location='eastus'; PackageVersion='latest' }

.\aura-deployment\scripts\deploy-azure-worker.ps1 @s

.\aura-deployment\scripts\deploy-azure-supervisor.ps1 @s `
  -AuthDevJwtSecret 'aura-dev-secret-change-me-use-32chars-minimum!!' `
  -InternalSharedSecret 'aura-internal-demo'

.\aura-deployment\scripts\deploy-azure-authz.ps1 @s

.\aura-deployment\scripts\deploy-azure-bff-api.ps1 @s `
  -AuthDevJwtSecret 'aura-dev-secret-change-me-use-32chars-minimum!!' `
  -InternalSharedSecret 'aura-internal-demo' `
  -CorsAllowedOrigins 'https://YOUR_STATIC_APP.azurestaticapps.net,http://localhost:19006'
```

**Docker Desktop Kubernetes** (same order; add **`-Local`** to each line — **`$env:GITHUB_TOKEN`** still needed unless **`SkipGhcrImagePullSecret`** for public GHCR):

```powershell
$env:GITHUB_TOKEN = '<pat read:packages>'
$k = @{ UniqueSuffix='jdoe01'; PackageVersion='latest'; Local=$true }

.\aura-deployment\scripts\deploy-azure-worker.ps1 @k
.\aura-deployment\scripts\deploy-azure-supervisor.ps1 @k `
  -AuthDevJwtSecret 'aura-dev-secret-change-me-use-32chars-minimum!!' `
  -InternalSharedSecret 'aura-internal-demo'
.\aura-deployment\scripts\deploy-azure-authz.ps1 @k
.\aura-deployment\scripts\deploy-azure-bff-api.ps1 @k `
  -AuthDevJwtSecret 'aura-dev-secret-change-me-use-32chars-minimum!!' `
  -InternalSharedSecret 'aura-internal-demo' `
  -CorsAllowedOrigins 'http://127.0.0.1:30880,http://127.0.0.1:30900,http://localhost:19006'

.\aura-deployment\scripts\deploy-azure-frontend.ps1 @k -Local
```

#### Combined release: backends + manifest (+ optional SWA)

**`deploy-azure-full-release.ps1`** runs **1→4** (backends), then **`-Local`:** **`deploy-azure-frontend.ps1`** (unless **`-SkipLocalFrontend`**) **5**, then writes **`aura-release-{suffix}-{sanitized-tag}.json`** (**Azure:** **`az webapp show`**; **`-Local`:** localhost URLs including **`frontend_https`** on **30900**). With **`-PackageVersion`** + **`-GhcrOwner`**, the JSON includes **`package_version`**, **`images.*`**, **`terraform_hints`**.

**`-PublishStaticWebApp`** runs **`Publish-AuraAzureStaticWebApp`** (pull **`aura-frontend`** → SWA CLI). **Not supported with `-Local`** — use Azure only for SWA.

```powershell
$env:GITHUB_TOKEN = '<pat read:packages>'
.\aura-deployment\scripts\deploy-azure-full-release.ps1 `
  -PackageVersion $(git rev-parse --short HEAD) `
  -UniqueSuffix 'jdoe01' `
  -ResourceGroupName 'aura-rg' `
  -CorsAllowedOrigins 'https://YOUR_STATIC_APP.azurestaticapps.net,http://localhost:19006' `
  -PublishStaticWebApp -UseAzureCliForToken -StrictBffBundleMatch
```

**Local combined example:**

```powershell
$env:GITHUB_TOKEN = '<pat read:packages>'
.\aura-deployment\scripts\deploy-azure-full-release.ps1 `
  -PackageVersion 'latest' `
  -UniqueSuffix 'jdoe01' `
  -Local `
  -CorsAllowedOrigins 'http://127.0.0.1:30880,http://localhost:19006'
```

**`-PackageVersion`** is mandatory so images and manifest agree on one release tag.

| Backend parameter | Typical use |
|-------------------|-------------|
| `-UniqueSuffix` | **Azure:** **`aura-{suffix}-worker`**, … **Local:** Kubernetes namespace **`aura-{suffix}`**. |
| `-SharedPlanName` | Azure Linux plan (default **`aura-free-plan`**). Ignored when **`-Local`**. |
| `-AppServiceSku` | Azure SKU (**`FREE`** default). Ignored when **`-Local`**. |
| `-PackageVersion` | GHCR tag (match Terraform **`image_tag`**). |
| `-ImageTag` | Legacy alias for **`-PackageVersion`**. |
| `-Local` | Deploy to **Docker Desktop Kubernetes** instead of Azure Web Apps. |
| `-SkipLocalFrontend` | **`-Local`:** skip **`deploy-azure-frontend.ps1`** (APIs only). |
| `-SkipGhcrImagePullSecret` | **`-Local`:** omit **`imagePullSecrets`** / **`ghcr-pull`** (public images only). |
| *(default tag)* | Omit **`-PackageVersion`** and **`-ImageTag`** → **`latest`**. |
| `-SkipDependencyChecks` / `-SkipHealthWait` | Debugging only; prefer correct order + warm **`/healthz`**. |

### 3 — Point the frontend **build** at the live BFF

The SPA calls **aura-bff-api only** (`EXPO_PUBLIC_*`). Values are **baked in at `expo export`** inside the **`aura-frontend`** Docker build.

From **`aura-release-*.json`** (**`endpoints.bff_https`** / **`terraform_hints`**) or the Azure Portal, read the **BFF** origin (e.g. `https://aura-jdoe01-bff.azurewebsites.net`; locally **`http://127.0.0.1:30880`**). In **`aura-deployment/terraform.tfvars`**:

```hcl
frontend_expo_public_api_base_url = "https://aura-jdoe01-bff.azurewebsites.net/v1"
frontend_expo_public_ws_base_url  = "wss://aura-jdoe01-bff.azurewebsites.net"
```

Then from **`aura-deployment/`**:

```powershell
terraform apply
```

Bump **`frontend_rebuild_stamp`** if needed so the frontend image rebuilds.

### 4 — Publish the web UI to Azure Static Web Apps

1. **`docker login ghcr.io`** if **`aura-frontend`** on GHCR is private.
2. **`az login`** if you use **`-UseAzureCliForToken`**, **or** set **`AZURE_STATIC_WEB_APPS_API_TOKEN`**.

Use **`deploy-azure-full-release.ps1`** with **`-PublishStaticWebApp`** so the release manifest and SWA publish stay aligned (see § combined release). Alternatively dot-source **`deploy-azure-backend-common.ps1`** and call **`Publish-AuraAzureStaticWebApp`** with **`-BackendStackSummaryJson`**, **`-UseAzureCliForToken`**, **`-StrictBffBundleMatch`**, etc.

By default **`Publish-AuraAzureStaticWebApp`** probes **`GET {BFF}/healthz`** before upload when a BFF URL is known (manifest or **`-BffHttpsRoot`**).

### 5 — Smoke test

Open the Static Web App URL; confirm browser traffic goes to **`https://…-bff…/v1`** and WebSockets **`wss://…-bff…`**.

---

## `Publish-AuraAzureStaticWebApp` (Static Web Apps — in common module)

Defined in **`deploy-azure-backend-common.ps1`**. Pulls **`aura-frontend`** from GHCR, extracts **`/usr/share/nginx/html`** (see [`services/frontend/Dockerfile`](../services/frontend/Dockerfile)), runs **`npx @azure/static-web-apps-cli deploy`**.

Prerequisites: **Docker**, **Node/npm**, **`az`** when using **`-UseAzureCliForToken`**.

| Parameter | Purpose |
|-----------|---------|
| `-ContainerImage` | Frontend image (default `ghcr.io/sushanth262/aura-frontend:latest`) |
| `-BackendStackSummaryJson` | Release manifest (`bff_https`, `expo_hints`) |
| `-BffHttpsRoot` | BFF origin if no JSON (`http://` or `https://` supported for probes) |
| `-UseAzureCliForToken` / `-StaticSiteResourceId` | Resolve SWA deployment token via **`az staticwebapp secrets list`** |
| `-StrictBffBundleMatch` | Validate extracted **`*.js`** against expected BFF host |

Advanced: dot-source **`deploy-azure-backend-common.ps1`**, then **`Publish-AuraAzureStaticWebApp @params`**.

---


## Backend images only (Terraform) & local dev

Terraform builds **`aura-authz`**, **`aura-bff-api`**, **`aura-supervisor`**, **`aura-worker`**, and **`aura-frontend`** — see **`../README.md`**.

**Local stack (no Azure):** from repo root:

```bash
docker compose -f aura-deployment/docker-compose.backend.yml up --build
```

---

## Local Docker provider (Terraform on Windows)

See **`setup-local-docker-provider.ps1`** — staging **`terraform-provider-docker`** for **`terraform init -plugin-dir`**.
