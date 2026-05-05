# Scripts

## Deploy frontend to Azure Static Web Apps

`deploy-azure-static-web-app.ps1` pulls the published frontend image from **GHCR**, copies the nginx static root (`/usr/share/nginx/html`, same path as in [`services/frontend/Dockerfile`](../services/frontend/Dockerfile)), and deploys that folder with the [Azure Static Web Apps CLI](https://learn.microsoft.com/azure/static-web-apps/).

Run from anywhere (repo root is convenient):

```powershell
# Private registry: log in first
docker login ghcr.io -u YOUR_USER

az login
.\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -UseAzureCliForToken

# Or set token from Portal -> Static Web App -> Manage deployment token
$env:AZURE_STATIC_WEB_APPS_API_TOKEN = '<your-deployment-token>'
.\aura-deployment\scripts\deploy-azure-static-web-app.ps1
```

From `aura-deployment`: `.\scripts\deploy-azure-static-web-app.ps1`

### Prerequisites

- **Docker** CLI (pull + create + cp)
- **Node.js / npm** (for `npx` SWA CLI only; no local Expo build)
- **`az`** only if you use `-UseAzureCliForToken`

### Parameters (high level)

| Parameter | Purpose |
|-----------|---------|
| `-ContainerImage` | Image ref (default: `ghcr.io/sushanth262/aura-frontend:latest`) |
| `-SkipDockerPull` | Use already-pulled local image |
| `-StaticFilesPathInImage` | Path inside image (default matches Aura Dockerfile) |
| `-StagingDirectory` | Extract here and keep files; default is temp dir removed after deploy |
| `-Environment` | `production` or `preview` |
| `-StaticSiteResourceId` | ARM id for token lookup via `-UseAzureCliForToken` |

See the script header for full help.

### Logo or UI looks outdated after deploy

The script deploys **whatever static files are inside the container image**, not your working tree. After changing assets under `aura-frontend/icons/` or branding code, **rebuild and push the frontend image to GHCR** with Terraform (`terraform apply` from `aura-deployment`), then run this script again so `docker pull` gets the new digest.

If GHCR digest stays the same after bumping **`frontend_rebuild_stamp`**, Docker BuildKit was reusing the **`expo export`** layer. The Dockerfile now passes the stamp as **`CACHEBUST`** so that layer reruns; bump the stamp again after pulling latest Terraform/Dockerfile. (Older setups without `CACHEBUST` needed `terraform apply -replace=...` or a no-cache build.)

Alternative without editing vars:  
`terraform apply -replace=module.frontend.module.docker_image.docker_image.this`

If the UI still looks old after a new digest, hard-refresh the browser; Static Web Apps can cache aggressively.

## Local Docker provider

See `setup-local-docker-provider.ps1` in this folder for Docker-related setup used by deployment tooling.
