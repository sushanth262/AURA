# Aura deployment (Terraform + Docker)

This folder builds and pushes **container images** to **GHCR**: the Expo web bundle (**`aura-frontend`**) and four Go services (**`aura-authz`**, **`aura-bff-api`**, **`aura-supervisor`**, **`aura-worker`**).

## Terraform

From this directory (after `docker login ghcr.io` and `TF_VAR_github_token` set):

1. `terraform init` (use `-plugin-dir` if you mirror the Docker provider — see [`scripts/setup-local-docker-provider.ps1`](scripts/setup-local-docker-provider.ps1))
2. Copy `terraform.tfvars.example` → `terraform.tfvars` and adjust `ghcr_owner`, stamps, etc.
3. `terraform apply`

**Outputs:** `frontend_image`, `aura_authz_image`, `aura_bff_api_image`, `aura_supervisor_image`, `aura_worker_image`, plus matching `*_digest` values.

Force rebuilds without source changes: bump **`frontend_rebuild_stamp`** or **`backend_rebuild_stamp`** in `terraform.tfvars`.

## Local backend (all four APIs)

From **repository root**:

```bash
docker compose -f aura-deployment/docker-compose.backend.yml up --build
```

The UI / Expo app should target the **BFF** only: `http://localhost:8080/v1` (REST) and `ws://localhost:8080` (WebSocket). Environment matrix and curl examples: **[`docs/BFF_AUTH_LOGIN.md`](../docs/BFF_AUTH_LOGIN.md)**.

## Deploy backend APIs to Azure App Service (FREE SKU default)

PowerShell scripts under **`scripts/`** deploy **four backends** — **Azure App Service** by default, or **Docker Desktop Kubernetes** with **`-Local`** (manifests in **`k8s/docker-desktop/`**, APIs **30880–30883**, UI **`deploy-azure-frontend.ps1`** on **30900**). **One script per backend** plus **`deploy-azure-frontend.ps1`** for local Kubernetes UI; pass **`-PackageVersion`** (GHCR tag, aligned with Terraform **`image_tag`**) — **`latest`** when omitted (**`-ImageTag`** alias).

| Order | Script | Peer checks |
|------:|--------|-------------|
| 1 | **`deploy-azure-worker.ps1`** | — |
| 2 | **`deploy-azure-supervisor.ps1`** | Worker exists + `/healthz` (unless `-SkipDependencyChecks`) |
| 3 | **`deploy-azure-authz.ps1`** | — |
| 4 | **`deploy-azure-bff-api.ps1`** | AuthZ + supervisor `/healthz` |
| 5 | **`deploy-azure-frontend.ps1`** (`-Local` only) | BFF up; CORS includes UI origin (**`-SkipLocalFrontend`** on full-release omits this step) |

**`deploy-azure-full-release.ps1`** runs backends (**`-Local`:** then **`deploy-azure-frontend.ps1`** unless **`-SkipLocalFrontend`**) and writes **`aura-release-{suffix}-{tag}.json`**. **`-PublishStaticWebApp`** uploads the frontend on Azure (**not** combined with **`-Local`**).

Shared naming: **`-UniqueSuffix jdoe01`** → `aura-jdoe01-worker`, `aura-jdoe01-supervisor`, `aura-jdoe01-authz`, `aura-jdoe01-bff`. Services auto-link via HTTPS URLs discovered with **`az webapp show`** (`WORKER_URL`, `AUTHZ_URL`, `SUPERVISOR_URL`, matching JWT + internal secret).

Requires **`az login`**, GHCR pull PAT (**`GITHUB_TOKEN`** / **`GHCR_TOKEN`** + **`GHCR_USERNAME`**), and images already in GHCR from Terraform. Details: **[`scripts/README.md`](scripts/README.md)**.

## Deploy web UI to Azure Static Web Apps

Does **not** deploy the Go APIs — it publishes the **static frontend** from **`aura-frontend`** via **`Publish-AuraAzureStaticWebApp`** (invoked by **`deploy-azure-full-release.ps1 -PublishStaticWebApp`**). See **[`scripts/README.md`](scripts/README.md)**.

Set **`frontend_expo_public_*`** in Terraform before building the frontend image so the bundle targets your **BFF**. **`Publish-AuraAzureStaticWebApp`** accepts **`-BackendStackSummaryJson`** (release manifest) to probe **`/healthz`** and optional **`-StrictBffBundleMatch`**.

After `terraform apply`, pull the image tag your pipeline produced (or `:latest`), then run the script so SWA receives the matching static bundle.
