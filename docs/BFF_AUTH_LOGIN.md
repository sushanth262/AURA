# AURA API — authentication, authorization, and mock login

The **`aura-backend`** Go codebase ships **four processes** (separate containers in deployment):

| Service | Role | Default listen |
|---------|------|----------------|
| **`aura-bff-api`** | Public **BFF**: `/v1/auth/dev-token`, `/v1/api/…`, WebSocket **proxy** to supervisor | `:8080` |
| **`aura-authz`** | **Authorization** service; `POST /v1/evaluate`; emits structured audit JSON to stdout | `:8081` |
| **`aura-supervisor`** | Investigations store, YAML fixtures, mock timeline, outbound calls to **aura-worker**; native WS `/ws/investigations/{taskId}` | `:8082` |
| **`aura-worker`** | **Connector mocks**: Grafana-style metrics, GitHub commits, Jira issues — data sliced from wireframe YAML (`source_mocks`) | `:8083` |

Terraform builds **four** images (`services/authz`, `bff`, `supervisor`, `worker`). Outputs: `aura_authz_image`, `aura_bff_api_image`, `aura_supervisor_image`, `aura_worker_image` (each with matching `_digest`).

---

## Quick start (Docker Compose)

From repo root:

```bash
docker compose -f aura-deployment/docker-compose.backend.yml up --build
```

- Mint JWT and call REST against **`http://localhost:8080`** (BFF).
- Supervisor listens on **8082** (optional debugging).
- Worker has **no published port** in the default compose file (reachable from supervisor on the Compose network).

---

## Environment variables (by service)

### aura-bff-api

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8080` | Listen address |
| `AUTH_DEV_MOCK` | `true` | HS256 dev tokens (`/v1/auth/dev-token`). |
| `AUTH_DEV_JWT_SECRET` | long dev string | Shared secret **must match supervisor** for WS upgrade relay + JWT verification. Min 16 chars. |
| `AUTH_ISSUER`, `AUTH_AUDIENCE`, `AUTH_JWKS_URL` | see code | Same semantics as before (`jwtauth`). |
| `AUTHZ_URL` | `http://127.0.0.1:8081` | **aura-authz** base URL |
| `SUPERVISOR_URL` | `http://127.0.0.1:8082` | **aura-supervisor** HTTP/WS base (scheme flipped for WS dial). |
| `INTERNAL_SHARED_SECRET` | _(empty)_ | If set, sent as `X-Internal-Secret` on BFF→supervisor calls; supervisor **must** match. |
| `CORS_ALLOWED_ORIGINS` | localhost Expo ports | Browser origins |

### aura-authz

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8081` if unset in shell per binary default | Listen |
| `POLICY_VERSION` | `stub-v1` | Logged on audit lines |

### aura-supervisor

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8082` | Listen |
| `AUTH_DEV_MOCK`, `AUTH_DEV_JWT_SECRET` | _(same as BFF)_ | Validates bearer on investigation WebSocket. |
| `POLICY_VERSION` | `stub-v1` | Echoed in synthesis payloads |
| `WORKER_URL` | _(empty)_ | **aura-worker** base URL; if empty, timeline skips connector snapshots. |
| `WORKER_SOURCES` | `grafana,github,jira` | Which connector IDs supervisor may call. Add `slack,teams,email` when `communications` is enabled. |
| `GRAPH_ENGINE_MODE` | `engine` | `engine` runs the orchestration graph; `legacy` uses the original `RunMockScenario` timeline. |
| `ENABLED_AGENTS` | `telemetry,code,context` | Agent domains in the graph; add `communications` for a fifth swimlane (Phase 2). |
| `MIN_AGENTS_FOR_SYNTHESIS` | `1` | Minimum successful agent nodes before synthesis. |
| `SYNTHESIS_JOIN` | `any_success` | `any_success` or `all_required` join policy into synthesis. |
| `AGENT_EXECUTION_MODE` | `inline` | `inline` uses snapshot fetcher; `worker` POSTs `AgentTask` to aura-worker execute API. |
| `INTERNAL_SHARED_SECRET` | _(empty)_ | Optional gate on `/internal/v1/…` |

### aura-worker

| Variable | Default | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8083` | Listen |
| `WORKER_ENABLED_SOURCES` | `grafana,github,jira` | Routes exposed by **this** process (subset or reorder later per replica). Add `slack,teams,email` for communications agent. |

**Execute API (Phase 3):** `POST /v1/agents/{domain}/execute` with `AgentTask` body → `AgentResult`. Legacy `GET /v1/sources/{source}?scenario_key=…` remains for inline mode.

| Variable | Default | Purpose |
|----------|---------|---------|
| `CONNECTOR_GRAFANA_MODE` | `fixture` | `fixture` uses YAML mocks; `live` probes `GRAFANA_URL/api/health` and merges fixture payload. |
| `GRAFANA_URL` | _(empty)_ | Required when `CONNECTOR_GRAFANA_MODE=live`. |

AuthZ audit JSON lines still go to **aura-authz** stdout (`decision_id`, `policy_version`, …).

---

## Mock login (development)

### 1. Single-process builds (optional)

From repo root, pick one Dockerfile under `aura-deployment/services/<service>/Dockerfile` or use Compose above.

### 2. Mint a demo JWT (via BFF)

```bash
curl -sS -X POST http://localhost:8080/v1/auth/dev-token \
  -H "Content-Type: application/json" \
  -d '{"sub":"demo-operator","roles":["operator"],"tenant_id":"demo"}'
```

### 3. Call the API (via BFF)

```bash
TOKEN="<paste access_token>"

curl -sS http://localhost:8080/v1/api/incidents/history \
  -H "Authorization: Bearer $TOKEN"
```

Create investigation (fixture defaults to `inc2847_api_gateway`):

```bash
curl -sS -X POST http://localhost:8080/v1/api/incidents \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"API gateway 5xx spike","severity":"P2","scope":{"service":"payments-api","region":"us-east-1"},"time_window":{"start":"2026-05-03T14:30:00Z","end":null},"symptoms":"Elevated 5xx"}'
```

### 4. WebSocket progress

Still via **BFF** (upgrade proxied to supervisor):

`ws://localhost:8080/ws/investigations/{task_id}?token=<access_token>`

---

## Stub AuthZ (`stub-v1`)

Same matrix as before: `operator`/`admin` create; `viewer`/`operator`/`admin` read + history. BFF calls **aura-authz** `POST /v1/evaluate` before hitting supervisor.

---

## Wireframe fixtures & connector mocks

YAML bundles live in `aura-backend/internal/fixturesdata/`.

- **Scenario shell**: timeline labels + metadata (`inc2847_api_gateway.yaml`).
- **`source_mocks`**: `grafana`, `github`, `jira` trees served by **aura-worker** at  
  `GET /v1/sources/{source}?scenario_key=<fixture_base_name>`.

Supervisor attaches responses under WebSocket **`connector_snapshot`** on **`AGENT_COMPLETE`** when `WORKER_URL` is set.

---

## Frontend configuration

- **`EXPO_PUBLIC_API_BASE_URL`**: `http://localhost:8080/v1` (BFF).
- **`EXPO_PUBLIC_WS_BASE_URL`**: `ws://localhost:8080` (same host; path `/ws/...` unchanged).

---

## Switching to real OIDC

1. Set **`AUTH_DEV_MOCK=false`** on **BFF and supervisor** (JWT verification paths).
2. Extend **`internal/jwtauth`** with JWKS, or validate at API Gateway and forward identity hints (still call **aura-authz** for policy).

---

## Free-tier deployment notes

Static UI + four containers exceeds **single** F1 slot unless you collapse Compose onto one VM/host or consolidate tier billing. Typical split: **SWA** for UI + **one small container host** (or Azure Container Apps consumption) running these four images behind private networking.

### Azure App Service (per-service scripts)

PowerShell deploy scripts under **`aura-deployment/scripts/`** provision backends in order (**worker → supervisor → authz → BFF**) — Azure Web Apps by default, or **Docker Desktop Kubernetes** with **`-Local`**. **`deploy-azure-full-release.ps1`** writes the combined **`aura-release-*.json`** manifest and can run **`-PublishStaticWebApp`**. See **`aura-deployment/scripts/README.md`** and **`aura-deployment/README.md`**.
