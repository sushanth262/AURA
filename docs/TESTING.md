# Aura — Test commands

Central list of test and smoke commands. **Add new sections here** as you implement phases (or link to phase-specific files).

**Repo root:** `c:\Users\dsush\source\repos\Aura` (adjust paths if yours differ).

---

## Run everything (quick)

### Backend (Go)

```powershell
cd aura-backend
go test ./... -count=1
```

With race detector (matches CI):

```powershell
cd aura-backend
go test ./... -count=1 -race
```

Verbose:

```powershell
cd aura-backend
go test ./... -count=1 -v
```

### Frontend (Jest)

```powershell
cd aura-frontend
npm install
npm test
```

### All automated (PowerShell one-liner from repo root)

```powershell
Push-Location aura-backend; go test ./... -count=1; if (-not $?) { Pop-Location; exit 1 }
Pop-Location
Push-Location aura-frontend; npm test; if (-not $?) { Pop-Location; exit 1 }
Pop-Location
Write-Host "All automated tests passed." -ForegroundColor Green
```

### Bash (Linux / macOS / GitHub Actions)

```bash
cd aura-backend && go test ./... -count=1 -race
cd ../aura-frontend && npm ci && npm test
```

---

## CI

| Workflow | Trigger | Command |
|----------|---------|---------|
| [`.github/workflows/backend-test.yml`](../.github/workflows/backend-test.yml) | Push/PR touching `aura-backend/**` | `go test ./... -count=1 -race` |

---

## Backend — by package

Run from `aura-backend/`.

| Package | Phase | Command |
|---------|-------|---------|
| All | — | `go test ./... -count=1` |
| Config | 0 | `go test ./internal/config/... -count=1 -v` |
| Orchestration (policies) | 0 | `go test ./internal/orchestration -count=1 -v` |
| Registry | 0 | `go test ./internal/orchestration/registry/... -count=1 -v` |
| Contract (JSON shapes) | 0 | `go test ./internal/orchestration/contract/... -count=1 -v` |
| Graph (planner, runner, checkpoint) | 1 | `go test ./internal/orchestration/graph/... -count=1 -v` |
| Worker pipeline + execute API | 3 | `go test ./internal/workersvc/... -count=1 -v` |
| Connector runtime + circuit breaker | 5 | `go test ./internal/connectors/... -count=1 -v` |
| Supervisor agent worker client | 3 | `go test ./internal/supervisor/... -count=1 -v -run TestHTTPAgentWorkerClient` |

**Phase 0 + 1 together:**

```powershell
go test ./internal/config/... ./internal/orchestration/... -count=1 -v
```

**Single test:**

```powershell
go test ./internal/orchestration/graph/... -count=1 -v -run TestRunner_GoldenEventSequence
```

**Coverage (optional):**

```powershell
go test ./internal/orchestration/... -count=1 -coverprofile=coverage.out
go tool cover -func=coverage.out
```

---

## Backend — by phase (supervisor / agent pool)

See [SUPERVISOR_AGENT_POOL_PLAN.md](./SUPERVISOR_AGENT_POOL_PLAN.md) for implementation details.

| Phase | Status | Automated | Manual smoke |
|-------|--------|-----------|--------------|
| **0** | Done | `go test ./internal/config/... ./internal/orchestration/contract/... ./internal/orchestration/registry/... ./internal/orchestration -count=1` | — |
| **1** | Done | `go test ./internal/orchestration/graph/... -count=1 -v` | Graph engine + create incident (below) |
| **2** | Done | `go test ./internal/orchestration/graph/... ./internal/orchestration/registry/... -count=1 -v` | Enable comms: `ENABLED_AGENTS=telemetry,code,context,communications`; 5 swimlanes after GRAPH_PLANNED |
| **3** | Done | `go test ./internal/workersvc/... ./internal/supervisor/... -count=1 -v` | Worker execute + supervisor `AGENT_EXECUTION_MODE=worker` (below) |
| **4** | Done | `go test ./internal/orchestration/graph/... ./internal/workersvc/pipeline/... -count=1 -v -run 'FuseSynthesis|Communications|FourParallel'` | Comms enabled: 5 lanes, 3 comms finding types, evidence includes Communications tab |
| **5** | Done | `go test ./internal/connectors/... -count=1 -v` | Connector runtime + circuit breaker; optional Grafana live probe |
| **6** | Planned | _(add commands here)_ | Redis checkpoint, resume |

### Phase 1 — manual smoke (graph engine)

**Prerequisites:** BFF `:8080`, authz `:8081`, supervisor `:8082`, worker `:8083` (Compose or four `go run` processes).

```powershell
# Supervisor: graph engine (default)
$env:GRAPH_ENGINE_MODE = "engine"
$env:ENABLED_AGENTS = "telemetry,code,context"
# go run ./cmd/aura-supervisor   # in aura-backend
```

**Health:**

```powershell
Invoke-RestMethod http://localhost:8082/healthz
Invoke-RestMethod http://localhost:8080/healthz
```

**Create investigation + get task id:**

```powershell
$resp = Invoke-RestMethod -Method Post -Uri "http://localhost:8080/v1/auth/dev-token" `
  -ContentType "application/json" `
  -Body '{"sub":"demo-operator","roles":["operator"],"tenant_id":"demo"}'
$token = $resp.access_token

$inc = Invoke-RestMethod -Method Post -Uri "http://localhost:8080/v1/api/incidents" `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body '{"scenario_key":"inc2847_api_gateway","severity":"P2","title":"Phase 1 smoke"}'
$taskId = $inc.task_id
Write-Host "Task ID: $taskId"
Write-Host "WS: ws://localhost:8080/ws/investigations/$taskId?token=$token"
```

**UI:** `cd aura-frontend; npm run web` → open Expo URL → login → progress for `$taskId`.

**Legacy parity:**

```powershell
$env:GRAPH_ENGINE_MODE = "legacy"
# restart supervisor; repeat create incident — timeline should match engine mode
```

**Subset of agents:**

```powershell
$env:ENABLED_AGENTS = "telemetry"
$env:GRAPH_ENGINE_MODE = "engine"
# restart supervisor; only telemetry retrieve lane should run
```

### Phase 3 — manual smoke (worker execute)

**Worker direct:**

```powershell
cd aura-backend
go run ./cmd/aura-worker   # :8083

$body = @{
  incident_id = "inc-2847"
  task_id     = "task-smoke"
  domain      = "telemetry"
  fixture_key = "inc2847_api_gateway"
  connectors  = @("grafana")
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri "http://localhost:8083/v1/agents/telemetry/execute" `
  -ContentType "application/json" -Body $body
```

**Supervisor delegates to worker:**

```powershell
$env:WORKER_URL = "http://localhost:8083"
$env:AGENT_EXECUTION_MODE = "worker"
$env:GRAPH_ENGINE_MODE = "engine"
# go run ./cmd/aura-supervisor
# create incident via BFF; AGENT_COMPLETE should include findings from worker pipeline
```

Default `AGENT_EXECUTION_MODE=inline` keeps the snapshot fetcher path (`GET /v1/sources/{source}`).

### Phase 4 — manual smoke (communications connectors)

Requires communications agent and comms connectors on worker:

```powershell
# Worker
$env:WORKER_ENABLED_SOURCES = "grafana,github,jira,slack,teams,email"
go run ./cmd/aura-worker

# Supervisor
$env:ENABLED_AGENTS = "telemetry,code,context,communications"
$env:WORKER_SOURCES = "grafana,github,jira,slack,teams,email"
$env:WORKER_URL = "http://127.0.0.1:8083"
$env:AGENT_EXECUTION_MODE = "worker"
go run ./cmd/aura-supervisor
```

**Verify:** findings timeline includes `CHANNEL_ALERT_MENTION`, `ONCALL_PING`, `EMAIL_THREAD`; evidence narrative has **Communications Agent**; `confidence_breakdown.timeline_overlap_boost` is `0.05` when comms timestamps fall inside telemetry spike window.

### Phase 5 — manual smoke (connector runtime)

Worker now routes MCP through `internal/connectors` with per-connector circuit breakers:

```powershell
# Default: fixture mode (unchanged behavior)
go run ./cmd/aura-worker

Invoke-RestMethod "http://127.0.0.1:8083/v1/sources/grafana?scenario_key=inc2847_api_gateway"

# Optional Grafana live probe (falls back to fixture on probe failure)
$env:CONNECTOR_GRAFANA_MODE = "live"
$env:GRAFANA_URL = "http://localhost:3000"
go run ./cmd/aura-worker
```

Circuit open returns HTTP **503** `CIRCUIT_OPEN` on `/v1/sources/{source}` after repeated failures.

---

## Local stack for manual tests

**Docker (all four Go services):**

```powershell
cd c:\Users\dsush\source\repos\Aura
docker compose -f aura-deployment/docker-compose.backend.yml up --build
```

**Four terminals** — `go.mod` is under `aura-backend/` (not repo root):

```powershell
cd c:\Users\dsush\source\repos\Aura\aura-backend   # required for every go run / go test
```

Terminal 1:

```powershell
go run ./cmd/aura-worker
```

Terminal 2:

```powershell
go run ./cmd/aura-authz
```

Terminal 3:

```powershell
$env:WORKER_URL = "http://127.0.0.1:8083"
$env:GRAPH_ENGINE_MODE = "engine"
go run ./cmd/aura-supervisor
```

Terminal 4:

```powershell
go run ./cmd/aura-bff-api
```

**Frontend:**

```powershell
cd aura-frontend
npm install
npm run web
```

| Service | URL |
|---------|-----|
| UI (Expo web) | URL printed by `npm run web` (often `http://localhost:8081` or `19006`) |
| BFF API | http://localhost:8080 |
| Supervisor | http://localhost:8082/healthz (not `/`) |
| Authz | http://localhost:8081/healthz |
| Worker | http://localhost:8083/healthz |

Env reference: [BFF_AUTH_LOGIN.md](./BFF_AUTH_LOGIN.md).

---

## Frontend tests

```powershell
cd aura-frontend
npm test
npm run test:coverage
npm run type-check
npm run lint
```

**Watch mode:**

```powershell
npm run test:watch
```

---

## Adding new tests (checklist)

When you add a phase or feature:

1. **Unit/integration tests** in the appropriate package under `aura-backend/` or `aura-frontend/`.
2. **Register the command** in the table [Backend — by package](#backend--by-package) or [Backend — by phase](#backend--by-phase-supervisor--agent-pool).
3. **CI:** extend [`.github/workflows/backend-test.yml`](../.github/workflows/backend-test.yml) only if tests live outside `go test ./...` (e.g. separate e2e job).
4. **Manual smoke:** add copy-paste steps under the phase row if operators need to verify behavior.
5. **Optional:** add a `-run TestYourNewTest` example under [Single test](#backend--by-package).

**Template for a new phase row:**

```markdown
| **N** | Done / Planned | `go test ./internal/your/package/... -count=1 -v` | Short manual steps |
```

---

## Troubleshooting

| Issue | Check |
|-------|--------|
| UI can’t reach API | `EXPO_PUBLIC_API_BASE_URL` → `http://localhost:8080/v1`; BFF running |
| Supervisor base URL “doesn’t work” | Use `/healthz`; UI is Expo, not `:8082` |
| WS unauthorized | Dev token from `/v1/auth/dev-token`; pass `?token=` on WS URL |
| Port 8081 in use | Authz uses 8081; run Expo on `19006` (`$env:PORT = "19006"; npm run web`) |
| No connector snapshots | Set supervisor `WORKER_URL=http://127.0.0.1:8083` and run worker |
| `cannot find main module` | Run `go` commands from **`aura-backend`**, not repo root (`Aura/`) |
| `connection was closed` on `:8080` | **BFF is not running** or **8080 is taken** (Docker/Expo). See below |

### Port 8080 already in use

`dev-token` goes to **aura-bff-api**, not supervisor. If nothing is running—or Docker/Expo owns 8080—you get *connection closed* or wrong responses.

**Check what owns 8080 (PowerShell):**

```powershell
netstat -ano | findstr ":8080"
Get-Process -Id <PID_from_last_column> | Select-Object ProcessName, Path
```

**Fix A — start BFF** (from `aura-backend`):

```powershell
go run ./cmd/aura-bff-api
```

You should see: `aura-bff-api listening on :8080`. Then:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/healthz
```

**Fix B — 8080 busy:** stop the other app, or run BFF on another port:

```powershell
$env:HTTP_ADDR = ":8090"
go run ./cmd/aura-bff-api
# Then use http://localhost:8090/v1/auth/dev-token and set frontend:
# EXPO_PUBLIC_API_BASE_URL=http://localhost:8090/v1
# EXPO_PUBLIC_WS_BASE_URL=ws://localhost:8090
```

**Fix C — full stack via Compose** (BFF published on 8080):

```powershell
cd c:\Users\dsush\source\repos\Aura
docker compose -f aura-deployment/docker-compose.backend.yml up --build
```

Wait until `aura-bff-api` is healthy, then call `http://localhost:8080/healthz` → `aura-bff-api`.

### Port 8081 already in use (authz)

Expo web often binds **8081**, same as **aura-authz** → `bind: Only one usage of each socket address`.

**Check:**

```powershell
netstat -ano | findstr ":8081"
Get-Process -Id <PID> | Select-Object ProcessName
```

**Fix A — run authz on another port** (then point BFF at it):

```powershell
# Terminal: authz
$env:HTTP_ADDR = ":8085"
go run ./cmd/aura-authz

# Terminal: BFF (must match)
$env:AUTHZ_URL = "http://127.0.0.1:8085"
go run ./cmd/aura-bff-api
```

**Fix B — run Expo on another port** (keep authz on 8081):

```powershell
cd aura-frontend
$env:PORT = "19006"
npm run web
```

**Suggested local port map** (avoids clashes with Docker / Expo):

| Service | Default | Use when busy |
|---------|---------|----------------|
| BFF | 8080 | **8090** (Docker often owns 8080) |
| Authz | 8081 | **8085** (Expo often owns 8081) |
| Supervisor | 8082 | 8082 |
| Worker | 8083 | 8083 |
| Expo web | 8081 | **19006** |

**All four Go services on alternate ports** (copy-paste, four terminals, `aura-backend`):

```powershell
# 1 — worker
go run ./cmd/aura-worker

# 2 — authz
$env:HTTP_ADDR = ":8085"
go run ./cmd/aura-authz

# 3 — supervisor
$env:WORKER_URL = "http://127.0.0.1:8083"
$env:GRAPH_ENGINE_MODE = "engine"
go run ./cmd/aura-supervisor

# 4 — BFF
$env:HTTP_ADDR = ":8090"
$env:AUTHZ_URL = "http://127.0.0.1:8085"
$env:SUPERVISOR_URL = "http://127.0.0.1:8082"
$env:CORS_ALLOWED_ORIGINS = "http://localhost:19006,http://localhost:8090"
go run ./cmd/aura-bff-api
```

**Frontend** (`aura-frontend/.env` or shell before `npm run web`):

```powershell
$env:EXPO_PUBLIC_API_BASE_URL = "http://localhost:8090/v1"
$env:EXPO_PUBLIC_WS_BASE_URL = "ws://localhost:8090"
$env:PORT = "19006"
npm run web
```

**Smoke:**

```powershell
Invoke-RestMethod http://127.0.0.1:8090/healthz
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8090/v1/auth/dev-token" -ContentType "application/json" -Body '{"sub":"demo","roles":["operator"]}'
```

**Or free 8080:** `docker compose -f aura-deployment/docker-compose.backend.yml down` then run BFF on `:8080` again.

---

## Document map

| Doc | Purpose |
|-----|---------|
| **This file** | Runnable test commands (living doc) |
| [SUPERVISOR_AGENT_POOL_PLAN.md](./SUPERVISOR_AGENT_POOL_PLAN.md) | Phased implementation plan |
| [BFF_AUTH_LOGIN.md](./BFF_AUTH_LOGIN.md) | Local ports, JWT, API examples |
