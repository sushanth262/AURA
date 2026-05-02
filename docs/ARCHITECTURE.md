# AURA — Architecture

**AURA** (Agentic Understanding & Root-cause Analysis) is an AI-driven, multi-agent diagnostic and observability system. It correlates telemetry, source code, and operational context under a **security-first**, modular design to produce high-fidelity root cause analysis (RCA) and remediation guidance.

This document follows the [C4 model](https://c4model.com/) (Context → Containers → Components). Implementation details at **Code** level are intentionally abstract until concrete services and repositories are pinned.

---

## Table of contents

1. [Goals and design principles](#1-goals-and-design-principles)
2. [C4 Level 1 — System context](#2-c4-level-1--system-context)
3. [C4 Level 2 — Containers](#3-c4-level-2--containers)
4. [C4 Level 3 — Components](#4-c4-level-3--components)
5. [End-to-end user and system flows](#5-end-to-end-user-and-system-flows)
6. [Operational scenarios](#6-operational-scenarios)
7. [Security, privacy, and deployment](#7-security-privacy-and-deployment)
8. [Production readiness](#8-production-readiness)
9. [Document map](#9-document-map)

---

## 1. Goals and design principles

| Principle | Implication |
|-----------|-------------|
| **Multi-agent orchestration** | A central supervisor delegates to domain agents; execution is a dynamic graph, not a single linear prompt. |
| **Grounding over guessing** | RAG over code, docs, and **incident memory** reduces hallucinations and ties conclusions to evidence. |
| **Connector abstraction** | External systems (metrics, logs, Git, tickets) sit behind interchangeable connectors (including **MCP**-style tool protocols). |
| **Security by default** | PII and secret redaction, least-privilege credentials, and boundary-aware deployment before LLM processing. |
| **Resilient async work** | Long investigations are queued and checkpointed; the UI stays responsive with streaming progress. |

---

## 2. C4 Level 1 — System context

**Primary actors:** operators and on-call engineers, automated alerting, and (optionally) service owners approving remediation.

**External systems:** observability backends, source control, work tracking, knowledge bases, and customer-private environments reached via an outbound-safe hybrid bridge.

```mermaid
flowchart LR
  subgraph Actors
    OP[Operator / SRE]
    OC[On-call engineer]
  end

  subgraph AURA["AURA platform"]
    CORE[Orchestration & RCA core]
  end

  subgraph External["External systems"]
    OBS[Observability]
    VCS[Source control]
    ITSM[Tickets & docs]
    VDB[(Vector DB)]
    REDIS[(Queue / state)]
  end

  OP -->|Incidents, HITL| CORE
  OC -->|Read reports| CORE
  CORE --> OBS
  CORE --> VCS
  CORE --> ITSM
  CORE --> VDB
  CORE --> REDIS
```

---

## 3. C4 Level 2 — Containers

Logical deployable units and data stores. Exact boundaries may collapse into fewer physical services in early implementations; the **responsibilities** remain stable.

```mermaid
flowchart TB
  subgraph Users
    UI[Web UI]
  end

  subgraph AURA_containers["AURA — containers"]
    API[API / BFF]
    ORCH[Supervisor orchestrator service]
    SEC[Security & redaction service]
    AGPOOL[Agent worker pool]
    BRIDGE[Hybrid data bridge client / gateway]
  end

  REDIS[(Redis — tasks & checkpoints)]
  VDB[(Milvus / Chroma — embeddings)]
  LLM[LLM inference endpoint]
  EXT[External data sources via MCP / connectors]

  UI <-->|HTTPS + WebSockets| API
  API --> ORCH
  ORCH --> REDIS
  ORCH --> AGPOOL
  AGPOOL --> SEC
  SEC --> LLM
  AGPOOL --> VDB
  AGPOOL --> BRIDGE
  BRIDGE --> EXT
  ORCH -->|progress events| REDIS
  API -->|subscribe / push| UI
```

| Container | Responsibility |
|-----------|------------------|
| **Web UI** | Incident intake, live progress, evidence review, HITL gates, remediation triggers. |
| **API / BFF** | AuthN/Z, session, aggregation of reads/writes, WebSocket fan-out. |
| **Supervisor orchestrator** | Investigation graph (e.g. LangGraph-style), sequencing vs parallelism, synthesis and confidence scoring. |
| **Security & redaction** | PII masking, token scrubbing, policy-driven redaction profiles per source. |
| **Agent worker pool** | Runs telemetry, code/RAG, and context/doc agents; applies masking before model calls. |
| **Hybrid data bridge** | Outbound-initiated secure channel (e.g. multiplexed gRPC) for on-prem / private data without inbound firewall holes. |
| **Redis** | Task queue, async decoupling, orchestration checkpoints, fault tolerance. |
| **Vector DB** | Embeddings for code, docs, playbooks, **incident memory** for recurring-pattern detection. |

---

## 4. C4 Level 3 — Components

Components live mainly inside the **Supervisor orchestrator** and **Agent worker pool**, with shared use of **MCP** (or equivalent) for tool execution.

```mermaid
flowchart TB
  subgraph Supervisor["Supervisor (orchestrator)"]
    SG[State graph engine]
    DEC[Decomposition & planning]
    SYN[Evidence synthesis & confidence]
    CP[Checkpoint / resume]
  end

  subgraph Workers["Specialized worker agents"]
    TEL[Telemetry / RCA agent]
    CODE[Code / fix agent]
    CTX[Context / doc agent]
  end

  subgraph Integration["Integration layer"]
    MCP[MCP / connector runtime]
    RAG[RAG & retrieval orchestration]
    MASK[Security layer hooks]
  end

  VDB[(Vector DB)]
  Q[(Redis queue & state)]

  DEC --> SG
  SG --> TEL & CODE & CTX
  TEL & CODE & CTX --> MCP
  TEL & CODE & CTX --> RAG
  RAG --> VDB
  TEL & CODE & CTX --> MASK
  TEL & CODE & CTX --> SYN
  SYN --> CP
  CP --> Q
  SG --> Q
```

### 4.1 Component responsibilities

| Component | Role |
|-----------|------|
| **State graph engine** | Non-linear investigation: branches, joins, replanning based on intermediate agent outputs. |
| **Decomposition & planning** | Turns intake (alert, symptom, stack trace) into a dynamic investigation graph. |
| **Telemetry / RCA agent** | Metrics/logs queries (PromQL, KQL, vendor APIs); spikes, saturation, regional anomalies, error bursts. |
| **Code / fix agent** | RAG over code; correlates regressions with commits/blame; proposes candidate fixes. |
| **Context / doc agent** | Tickets, runbooks, RFCs; distinguishes intentional change vs accidental regression. |
| **MCP / connector runtime** | Standardized execution to observability, Git providers, ITSM, etc., swappable without core logic changes. |
| **Evidence synthesis** | Fuses multi-agent outputs, assigns confidence, prepares narrative for UI and HITL. |
| **Checkpoint / resume** | Persists graph position and partial evidence for retries and partial failures. |

### 4.2 Structural relationships (compact)

```mermaid
classDiagram
  class Supervisor {
    +StateGraph state
    +decompose(incident)
    +synthesize(evidence)
    +checkpoint(store)
  }
  class WorkerAgent {
    +domain
    +querySource()
    +applyRAG()
  }
  class MCPConnector {
    +toolType
    +execute()
  }
  class SecurityLayer {
    +maskPII(data)
    +applyPolicy(source)
  }
  class VectorDB {
    +storeEmbedding(context)
    +search(query)
  }
  class TaskStore {
    +enqueue(task)
    +saveCheckpoint(node)
  }

  Supervisor "1" *-- "many" WorkerAgent : manages
  WorkerAgent ..> SecurityLayer : data passes through
  WorkerAgent --> MCPConnector : uses
  WorkerAgent --> VectorDB : retrieves
  Supervisor --> TaskStore : checkpoints
  MCPConnector --> ExternalSource : connects
  class ExternalSource {
    <<external>>
  }
```

---

## 5. End-to-end user and system flows

### 5.1 Investigation lifecycle (logical flow)

From intake through human gate to memory update — aligned with the reference lifecycle in project materials.

```mermaid
flowchart TD
  A[Issue intake: user / alert] --> B[Decomposition: supervisor graph generation]
  B --> C{Parallel retrieval}
  C --> D[Telemetry agent: metrics / logs]
  C --> E[Code agent: Git + RAG]
  C --> F[Context agent: Jira / docs]
  D & E & F --> G[Security layer: PII masking / policy redaction]
  G --> H[Correlation: synthesis & confidence scoring]
  H --> I[Output: UI via WebSockets / streaming]
  I --> J{HITL gate: human review}
  J -->|Rejected| K[Refine logic / adjust memory]
  K --> B
  J -->|Approved| L[Remediation execution]
  L --> M[Update incident memory in vector DB]
```

### 5.2 Sequence — asynchronous orchestration

```mermaid
sequenceDiagram
  autonumber
  participant User
  participant UI
  participant Queue as Redis / task store
  participant Supervisor
  participant Agents as Worker agents
  participant MCP as MCP / connectors

  User->>UI: Submit incident
  UI->>Queue: Enqueue investigation task
  UI-->>User: Task ID (immediate response)

  Supervisor->>Queue: Fetch / claim task
  Note over Supervisor: State machine / graph start
  Supervisor->>Queue: Checkpoint initial state
  Supervisor->>Agents: Delegate domain tasks

  par Parallel retrieval
    Agents->>MCP: Query data sources
    MCP-->>Agents: Raw data
    Agents->>Agents: Apply PII masking
    Agents-->>Supervisor: Grounded insights
  end

  Supervisor->>Queue: Update status / checkpoint
  Queue-->>UI: Progress events (e.g. WebSockets)
  UI-->>User: Live progress

  Supervisor->>Supervisor: Evidence synthesis & scoring
  Supervisor->>UI: Diagnostic narrative + HITL request
  UI-->>User: Review & approve / reject
```

### 5.3 Supervisor state machine (conceptual)

```mermaid
stateDiagram-v2
  [*] --> Intake
  Intake --> Planning: incident normalized
  Planning --> Retrieving: graph edges ready
  Retrieving --> Telemetry
  Retrieving --> Code
  Retrieving --> Context
  Telemetry --> Synthesis
  Code --> Synthesis
  Context --> Synthesis
  Synthesis --> HITL: confidence & report ready
  HITL --> Remediation: approved
  HITL --> Planning: rejected / refine
  Remediation --> MemoryUpdate: actions executed or logged
  MemoryUpdate --> [*]
```

### 5.4 Component interaction — single “vertical slice”

```mermaid
flowchart LR
  subgraph Input
    INC[Incident + optional artifacts]
  end

  INC --> SUP[Supervisor]
  SUP --> A1[Telemetry]
  SUP --> A2[Code+RAG]
  SUP --> A3[Context]
  A1 & A2 & A3 --> SEC[Redaction]
  SEC --> SUP
  SUP --> OUT[RCA + confidence + citations]
  OUT --> MEM[(Incident memory writeback)]
```

---

## 6. Operational scenarios

### 6.1 Recurring performance regression

```mermaid
sequenceDiagram
  participant Alert
  participant Supervisor
  participant Ctx as Context agent
  participant Tel as Telemetry agent
  participant VDB as Vector DB

  Alert->>Supervisor: New incident
  Supervisor->>VDB: Semantic search — similar past incidents
  VDB-->>Ctx: Prior incident narrative + resolution
  Ctx-->>Supervisor: Historical pattern candidate
  Supervisor->>Tel: Validate metrics vs past signature
  Tel-->>Supervisor: Matching p95 / error pattern
  Supervisor->>Supervisor: Raise confidence (memory match)
  Supervisor->>Supervisor: HITL sees prior context, faster approval
```

### 6.2 Hybrid cloud / on-prem ingestion

```mermaid
flowchart TB
  subgraph Customer["Customer environment"]
    ONPREM[Apps & private Git / telemetry]
    AGENT["Bridge agent — outbound only (no inbound ports at customer edge)"]
  end

  subgraph Cloud["AURA cloud"]
    GW[Bridge gateway]
    ORCH[Orchestrator & agents]
    SEC[Redaction]
  end

  ONPREM --> AGENT
  AGENT -->|Outbound gRPC / multiplexed stream| GW
  GW --> ORCH
  ORCH --> SEC
```

---

## 7. Security, privacy, and deployment

| Topic | Approach |
|-------|----------|
| **PII / secrets** | Automated scrubbing (emails, IPs, tokens); policy profiles per connector and tenant. |
| **Access** | Least privilege, scoped credentials, auditable tool calls. |
| **Residency** | Optional full processing inside customer network; hybrid bridge for controlled egress. |
| **Deployment models** | **SaaS**; **customer VPC**; **on-prem** (including local LLM options). |

---

## 8. Production readiness

- **Horizontal scaling:** Independent pools for telemetry-heavy vs code-heavy workloads.
- **Fault tolerance:** Retries on transient connector failures; **checkpointing** avoids full investigation restarts.
- **Real-time UX:** WebSockets / server push for progress and HITL prompts.
- **Continuous improvement:** Human accept/reject on remediation feeds back into **confidence scoring** and retrieval ranking.

---

## 9. Document map

| Artifact | Purpose |
|----------|---------|
| [README.md](../README.md) | Project entry, link to this architecture |
| [AI Diagnostic Agent.docx](./AI%20Diagnostic%20Agent.docx) | Narrative architecture and workflow (source) |
| [AIdiagosticsWithDiagramdetails.pdf](./AIdiagosticsWithDiagramdetails.pdf) | Consolidated report with diagram snippets (source) |
| **This file** | C4-aligned, Mermaid-first architecture for AURA |

---

*This architecture is derived from the materials under `docs/` and is the living target for AURA implementations. Update this document when container boundaries, stores, or connector standards change.*
