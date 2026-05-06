# AURA

<p align="center">
  <img src="aura-frontend/icons/Aura_logo_cropped.png" alt="AURA — AI-driven. Multi-agent. Grounded insights. Diagnose. Correlate. Resolve." width="480" />
</p>

**Agentic Understanding & Root-cause Analysis**

AURA is an AI-driven, multi-agent diagnostic platform that correlates telemetry, source code, and operational context to produce grounded root cause analysis and remediation guidance—built around a supervisor orchestrator, specialized agents, retrieval memory, and security-first data handling.

## Deploy

- **Backend APIs (four containers):** built via Terraform under **`aura-deployment/`** (`aura-authz`, `aura-bff-api`, `aura-supervisor`, `aura-worker`). Local full stack: **`docker compose -f aura-deployment/docker-compose.backend.yml up --build`** from repo root. Operators’ guide: **`aura-deployment/README.md`** and **`docs/BFF_AUTH_LOGIN.md`** (JWT, ports, env vars).
- **Web UI (Azure Static Web Apps):** the frontend image is **`aura-deployment/services/frontend`** on **GHCR**. Publish with **`aura-deployment/scripts/deploy-azure-full-release.ps1 -PublishStaticWebApp`** (or **`Publish-AuraAzureStaticWebApp`** in **`deploy-azure-backend-common.ps1`**). Details: **`aura-deployment/scripts/README.md`**.

## Documentation

| Document | Description |
|----------|-------------|
| [**Architecture (C4 + Mermaid)**](docs/ARCHITECTURE.md) | End-to-end architecture: system context, containers, components, user flows, sequences, and operational scenarios |
| [AI Diagnostic Agent (Word)](docs/AI%20Diagnostic%20Agent.docx) | Source narrative: components, hybrid bridge, security, deployment |
| [AI diagnostics with diagram details (PDF)](docs/AIdiagosticsWithDiagramdetails.pdf) | Consolidated report and diagram references |

Start with [**docs/ARCHITECTURE.md**](docs/ARCHITECTURE.md) for the canonical, diagram-heavy view aligned with the C4 model.
