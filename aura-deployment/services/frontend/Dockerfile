# Build context: repo root (Aura/)
# Run as: docker build -f aura-deployment/services/frontend/Dockerfile .
#
# ─────────────────────────────────────────────────────────────────────────────
# Stage 1: Build — Expo web export
# ─────────────────────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

# Target runtime env for the static bundle (re-applied before `expo export`).
ARG NODE_ENV=production

ENV CI=true
# Expo + RN trees often have peer ranges npm treats as conflicts; CI=true makes that fatal without this.
ENV NPM_CONFIG_LEGACY_PEER_DEPS=true

WORKDIR /app

# npm omits devDependencies when NODE_ENV=production; Metro/Babel need them (e.g. babel-plugin-module-resolver).
ENV NODE_ENV=development

# Install dependencies (no package-lock.json in repo yet — npm ci would fail; npm install works)
COPY aura-frontend/package.json ./
RUN npm install --ignore-scripts

# Full source (overwrites package.json with same file; refreshes deps if package.json changed)
COPY aura-frontend/ ./
RUN npm install --ignore-scripts

# Export static web bundle → /app/dist
# CACHEBUST is passed from Terraform (frontend_rebuild_stamp) so bumping the stamp invalidates
# this layer even when COPY layers would otherwise hit BuildKit cache unchanged.
#
# Baked into the static bundle — MUST match the deployed aura-bff-api origin for Azure/production.
# Override via terraform.tfvars or: docker build --build-arg EXPO_PUBLIC_API_BASE_URL=https://YOUR-BFF.azurewebsites.net/v1 ...
ARG EXPO_PUBLIC_API_BASE_URL=http://localhost:8080/v1
ARG EXPO_PUBLIC_WS_BASE_URL=ws://localhost:8080
ENV EXPO_PUBLIC_API_BASE_URL=$EXPO_PUBLIC_API_BASE_URL
ENV EXPO_PUBLIC_WS_BASE_URL=$EXPO_PUBLIC_WS_BASE_URL

ARG NODE_ENV
ARG CACHEBUST=
ENV NODE_ENV=$NODE_ENV
RUN printf '%s\n' "${CACHEBUST:-}" >/tmp/.aura-cachebust && npx expo export --platform web

# ─────────────────────────────────────────────────────────────────────────────
# Stage 2: Serve — nginx static file server
# ─────────────────────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine AS runner

LABEL org.opencontainers.image.title="AURA Frontend" \
      org.opencontainers.image.description="Aura AI Root Cause Analysis — Web UI" \
      org.opencontainers.image.url="https://github.com/sushanth262/Aura"

# Replace default nginx site with Aura SPA config
RUN rm /etc/nginx/conf.d/default.conf
COPY aura-deployment/services/frontend/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built web assets from Stage 1
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
