# Aura VM Deployment Guide

Deploy all Aura services to an Azure Ubuntu VM using Docker Compose.

## Architecture

```
                    Port 80
Browser ──────────► nginx-proxy
                    ├── /v1/*    ──► aura-bff-api (8080)
                    ├── /ws      ──► aura-bff-api (WebSocket)
                    └── /*       ──► aura-frontend (80)

Internal (Docker network only):
  aura-bff-api ──► aura-authz (8081)
               ──► aura-supervisor (8082) ──► aura-worker (8083)
```

All services run as Docker containers behind an nginx reverse proxy on port 80.

## Prerequisites

- Azure VM (Ubuntu 24.04, e.g. Standard_B2ats_v2)
- SSH access to the VM (`.pem` key or `az ssh vm`)
- GHCR images pushed (via `terraform apply` in `aura-deployment/`)
- GitHub PAT with `read:packages` scope

## VM Details

| Property | Value |
|---|---|
| Resource Group | `auravm` |
| VM Name | `aura` |
| Location | `northcentralus` |
| Size | `Standard_B2ats_v2` |
| OS | Ubuntu 24.04 LTS |
| SSH User | `azureuser` |
| Public IP | `52.162.205.226` |

## Step 1: SSH into the VM

```powershell
# Using .pem key
ssh -i "C:\Users\dsush\.ssh\aura-ssh-key.pem" azureuser@52.162.205.226

# Or using Azure CLI (no local key needed)
az ssh vm --resource-group auravm --vm-name aura
```

## Step 2: Install Docker

```bash
sudo apt-get update -qq
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo tee /etc/apt/keyrings/docker.asc > /dev/null
sudo chmod a+r /etc/apt/keyrings/docker.asc

ARCH=$(dpkg --print-architecture)
CODENAME=$(. /etc/os-release; echo $VERSION_CODENAME)
echo "deb [arch=$ARCH signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $CODENAME stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update -qq
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker --version
docker compose version
```

## Step 3: Clone the repo

```bash
git clone https://github.com/sushanth262/AURA.git ~/Aura
cd ~/Aura
git checkout main
```

## Step 4: Generate docker-compose.yml from template

The template at `aura-deployment/docker-compose.vm.yml` uses placeholders. Replace them with your values:

```bash
TAG="local-ui-fix-8"
OWNER="sushanth262"
VM_IP=$(curl -s ifconfig.me)

sed -e "s|{{WORKER_IMAGE}}|ghcr.io/$OWNER/aura-worker:$TAG|g" \
    -e "s|{{SUPERVISOR_IMAGE}}|ghcr.io/$OWNER/aura-supervisor:$TAG|g" \
    -e "s|{{AUTHZ_IMAGE}}|ghcr.io/$OWNER/aura-authz:$TAG|g" \
    -e "s|{{BFF_IMAGE}}|ghcr.io/$OWNER/aura-bff-api:$TAG|g" \
    -e "s|{{FRONTEND_IMAGE}}|ghcr.io/$OWNER/aura-frontend:$TAG|g" \
    -e "s|{{AUTH_DEV_JWT_SECRET}}|aura-dev-secret-change-me-use-32chars-minimum!!|g" \
    -e "s|{{INTERNAL_SHARED_SECRET}}|aura-internal-demo|g" \
    -e "s|{{CORS_ALLOWED_ORIGINS}}|http://$VM_IP,http://localhost|g" \
    aura-deployment/docker-compose.vm.yml > ~/docker-compose.yml

# Copy nginx proxy config
cp aura-deployment/nginx-vm-proxy.conf ~/nginx-vm-proxy.conf
```

## Step 5: Login to GHCR

```bash
export GITHUB_TOKEN="<your PAT with read:packages>"
echo $GITHUB_TOKEN | docker login ghcr.io -u sushanth262 --password-stdin
```

## Step 6: Deploy

```bash
cd ~
docker compose pull
docker compose up -d
```

## Step 7: Verify

```bash
# Check all containers are running
docker compose ps

# Health checks
curl -s http://localhost/healthz    # BFF via proxy
curl -s http://localhost/health     # Frontend via proxy

# Check logs
docker compose logs -f aura-bff-api
```

## Endpoints

| Service | URL |
|---|---|
| Frontend | http://52.162.205.226 |
| BFF API | http://52.162.205.226/v1 |
| WebSocket | ws://52.162.205.226/ws |
| BFF Health | http://52.162.205.226/healthz |

## Updating

To deploy a new image tag:

```bash
cd ~/Aura
git pull

# Re-generate compose file with new tag
TAG="new-tag-here"
OWNER="sushanth262"
VM_IP=$(curl -s ifconfig.me)

sed -e "s|{{WORKER_IMAGE}}|ghcr.io/$OWNER/aura-worker:$TAG|g" \
    -e "s|{{SUPERVISOR_IMAGE}}|ghcr.io/$OWNER/aura-supervisor:$TAG|g" \
    -e "s|{{AUTHZ_IMAGE}}|ghcr.io/$OWNER/aura-authz:$TAG|g" \
    -e "s|{{BFF_IMAGE}}|ghcr.io/$OWNER/aura-bff-api:$TAG|g" \
    -e "s|{{FRONTEND_IMAGE}}|ghcr.io/$OWNER/aura-frontend:$TAG|g" \
    -e "s|{{AUTH_DEV_JWT_SECRET}}|aura-dev-secret-change-me-use-32chars-minimum!!|g" \
    -e "s|{{INTERNAL_SHARED_SECRET}}|aura-internal-demo|g" \
    -e "s|{{CORS_ALLOWED_ORIGINS}}|http://$VM_IP,http://localhost|g" \
    aura-deployment/docker-compose.vm.yml > ~/docker-compose.yml

cp aura-deployment/nginx-vm-proxy.conf ~/nginx-vm-proxy.conf

cd ~
docker compose pull
docker compose up -d --remove-orphans
```

## Rebuilding Frontend with VM IP

If the frontend shows network errors, rebuild the image with the VM IP baked in:

1. Update `aura-deployment/terraform.tfvars`:
   ```hcl
   frontend_expo_public_api_base_url = "http://52.162.205.226/v1"
   frontend_expo_public_ws_base_url  = "ws://52.162.205.226"
   frontend_rebuild_stamp = "6"
   ```

2. Rebuild and push:
   ```powershell
   cd aura-deployment
   terraform apply -target=module.frontend
   ```

3. Redeploy on VM:
   ```bash
   cd ~
   docker compose pull aura-frontend
   docker compose up -d
   ```

## Troubleshooting

```bash
# View logs for a specific service
docker compose logs -f aura-bff-api
docker compose logs -f aura-supervisor
docker compose logs -f aura-worker

# Restart a single service
docker compose restart aura-bff-api

# Full restart
docker compose down
docker compose up -d

# Check resource usage
docker stats --no-stream
```

## NSG Ports

Ensure these ports are open in the VM's Network Security Group:

| Port | Purpose |
|---|---|
| 22 | SSH |
| 80 | Frontend + API (nginx proxy) |

Port 8080 is no longer needed externally since the nginx proxy handles routing.

## Automated Deployment (PowerShell)

From your local machine (requires Azure CLI login):

```powershell
cd aura-deployment/scripts
.\deploy-azure-vm.ps1 -PackageVersion 'local-ui-fix-8' -UniqueSuffix 'jdoe01' -VmPublicIp '52.162.205.226'
```

Or via the full-release script:

```powershell
.\deploy-azure-full-release.ps1 -PackageVersion 'local-ui-fix-8' -UniqueSuffix 'jdoe01' -VM
```
