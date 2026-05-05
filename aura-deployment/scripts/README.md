# Scripts

## Deploy frontend to Azure Static Web Apps

`deploy-azure-static-web-app.ps1` builds the Expo web app in **`aura-frontend`** (sibling folder at the monorepo root) and deploys with the [Azure Static Web Apps CLI](https://learn.microsoft.com/azure/static-web-apps/).

Run from the **repository root** (`Aura`, containing both `aura-frontend` and `aura-deployment`):

```powershell
cd <path-to-Aura-repo>

# Option A — deployment token via Azure CLI
az login
.\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -UseAzureCliForToken

# Option B — token from the portal (Static Web App → Overview → Manage deployment token)
$env:AZURE_STATIC_WEB_APPS_API_TOKEN = '<your-deployment-token>'
.\aura-deployment\scripts\deploy-azure-static-web-app.ps1

# Option B — prompt for token
.\aura-deployment\scripts\deploy-azure-static-web-app.ps1 -DeploymentToken (Read-Host -AsSecureString)
```

You can also `cd aura-deployment` and run `.\scripts\deploy-azure-static-web-app.ps1` (paths resolve the same way).

### Prerequisites

- Node.js and npm (for the build inside `aura-frontend`)
- For **`-UseAzureCliForToken`**: [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) and permission to read secrets on the target Static Web App

### Target app and overrides

Default Static Web App (override with `-StaticSiteResourceId`):

`/subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourcegroups/aura/providers/Microsoft.Web/staticSites/Aura`

| Parameter | Purpose |
|-----------|---------|
| `-AppRoot` | Path to the frontend app (default: `<repo-root>/aura-frontend`) |
| `-SkipNpmInstall` | Skip `npm ci` / `npm install` |
| `-SkipBuild` | Skip `expo export` |
| `-Environment` | `production` (default) or `preview` |
| `-OutputDir` | Build output folder relative to `AppRoot` (default: `dist`) |
| `-StaticSiteResourceId` | ARM id used with `-UseAzureCliForToken` |

Full parameter documentation is in the comment block at the top of `deploy-azure-static-web-app.ps1`.

## Local Docker provider

See `setup-local-docker-provider.ps1` in this folder for Docker-related setup used by deployment tooling.
