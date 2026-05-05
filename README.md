# AURA

<p align="center">
  <img src="aura-frontend/icons/Aura_logo_cropped.png" alt="AURA — AI-driven. Multi-agent. Grounded insights. Diagnose. Correlate. Resolve." width="480" />
</p>

**Agentic Understanding & Root-cause Analysis**

AURA is an AI-driven, multi-agent diagnostic platform that correlates telemetry, source code, and operational context to produce grounded root cause analysis and remediation guidance—built around a supervisor orchestrator, specialized agents, retrieval memory, and security-first data handling.

## Deploy web UI (Azure Static Web Apps)

The web frontend is built into the container defined under **`aura-deployment/services/frontend`** (published to **GHCR**). To deploy that bundle to Azure Static Web Apps without a local Expo build, use **`aura-deployment/scripts/deploy-azure-static-web-app.ps1`**: it pulls the image, copies the nginx static root from the container, and runs the SWA CLI. Prerequisites, parameters, and copy-paste examples are in **[`aura-deployment/scripts/README.md`](aura-deployment/scripts/README.md)**.

## Documentation

| Document | Description |
|----------|-------------|
| [**Architecture (C4 + Mermaid)**](docs/ARCHITECTURE.md) | End-to-end architecture: system context, containers, components, user flows, sequences, and operational scenarios |
| [AI Diagnostic Agent (Word)](docs/AI%20Diagnostic%20Agent.docx) | Source narrative: components, hybrid bridge, security, deployment |
| [AI diagnostics with diagram details (PDF)](docs/AIdiagosticsWithDiagramdetails.pdf) | Consolidated report and diagram references |

Start with [**docs/ARCHITECTURE.md**](docs/ARCHITECTURE.md) for the canonical, diagram-heavy view aligned with the C4 model.
