# Scripts

There are no frontend-local deploy scripts here. The web app image is built from **`aura-deployment/services/frontend/Dockerfile`** and pushed to **GHCR**.

To deploy the published image to **Azure Static Web Apps** (pull container → extract static files → SWA CLI), see **[`aura-deployment/scripts/README.md`](../../aura-deployment/scripts/README.md)** and **`deploy-azure-static-web-app.ps1`** in that folder.
