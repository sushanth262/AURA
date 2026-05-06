# Scripts

There are no frontend-local deploy scripts here. The web app image is built from **`aura-deployment/services/frontend/Dockerfile`** and pushed to **GHCR**.

To publish the built image to **Azure Static Web Apps**, see **[`aura-deployment/scripts/README.md`](../../aura-deployment/scripts/README.md)** (**`deploy-azure-full-release.ps1 -PublishStaticWebApp`** or **`Publish-AuraAzureStaticWebApp`** in **`deploy-azure-backend-common.ps1`**).
