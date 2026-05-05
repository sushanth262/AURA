# Scripts

## Deploy to Azure Static Web Apps

`deploy-azure-static-web-app.ps1` builds the Expo web export and deploys it with the [Azure Static Web Apps CLI](https://learn.microsoft.com/azure/static-web-apps/).

Run from the **`aura-frontend`** directory (the script defaults `AppRoot` to that folder):

```powershell
cd aura-frontend

# Option A — fetch the deployment token with Azure CLI (recommended if you have az + permissions)
az login
.\scripts\deploy-azure-static-web-app.ps1 -UseAzureCliForToken

# Option B — token from the portal (Static Web App → Overview → Manage deployment token)
$env:AZURE_STATIC_WEB_APPS_API_TOKEN = '<your-deployment-token>'
.\scripts\deploy-azure-static-web-app.ps1

# Option B — prompt for token at runtime
.\scripts\deploy-azure-static-web-app.ps1 -DeploymentToken (Read-Host -AsSecureString)
```

### Prerequisites

- Node.js and npm (same as the rest of `aura-frontend`)
- For **`-UseAzureCliForToken`**: [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) installed, signed in with access to list secrets on the target Static Web App

### Target app

By default the script targets this Static Web App resource (override with `-StaticSiteResourceId` if yours differs):

`/subscriptions/b0111f22-31ef-406d-88af-95034f5c7c1d/resourcegroups/aura/providers/Microsoft.Web/staticSites/Aura`

The deployment token is always bound to that app; the script only changes *how* the token is supplied, not which app you use, unless you change the resource id.

### Useful parameters

| Parameter | Purpose |
|-----------|---------|
| `-SkipNpmInstall` | Skip `npm ci` / `npm install` |
| `-SkipBuild` | Skip `expo export` (use if `dist` is already built) |
| `-Environment` | `production` (default) or `preview` |
| `-OutputDir` | Build output folder relative to `AppRoot` (default: `dist`) |
| `-StaticSiteResourceId` | ARM id used with `-UseAzureCliForToken` to pick subscription, resource group, and app name |

For the full list, see the comment block at the top of `deploy-azure-static-web-app.ps1`.
