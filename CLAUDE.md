# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

Aura is currently in the **architectural design phase**. The repository contains specification documents only — no source code, build scripts, or tests have been scaffolded yet. The primary reference is [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## What Aura Is

An AI-driven, multi-agent root cause analysis (RCA) platform for incidents. When an alert fires, Aura runs a supervised investigation graph that pulls telemetry, correlates code changes, retrieves runbooks/tickets, synthesizes evidence, and proposes remediations — with a mandatory human approval gate before any action is taken.

## Architecture

### Containers (deployable units)

| Container | Responsibility |
|---|---|
| Web UI | Incident intake, live progress via WebSocket, HITL approval gate |
| API / BFF | AuthN/Z, session management, WebSocket fan-out |
| Supervisor Orchestrator | State-graph engine; plans and coordinates agent workers |
| Security & Redaction Service | PII masking and secrets scrubbing before any data reaches the LLM |
| Agent Worker Pool | Three domain agents (see below) |
| Hybrid Data Bridge | Outbound-only gRPC multiplexed channel for on-prem/private-cloud data |
| Redis | Task queue, orchestration checkpoints, async decoupling |
| Vector DB (Milvus/Chroma) | Embeddings for code, docs, playbooks, and incident memory |

### The three agent workers

- **Telemetry/RCA Agent** — queries metrics (PromQL) and logs (KQL/Azure Monitor); detects spikes and error bursts
- **Code/Fix Agent** — RAG over source code; correlates recent commits/deploys with the incident timeline; proposes fixes
- **Context/Doc Agent** — retrieves Jira/ServiceNow tickets, runbooks, and RFCs; distinguishes intentional change from regression

### Investigation lifecycle

```
1. Normalize intake   → alert/symptoms → structured incident object
2. Plan graph         → Supervisor generates parallel + sequential retrieval edges
3. Retrieve & mask    → all three agents query concurrently; Security service scrubs output
4. Synthesize         → correlate evidence, assign confidence scores, build narrative
5. HITL gate          → operator reviews; rejection loops back to planning
6. Remediate          → approved actions executed (runbook automation, ticket creation)
7. Memory writeback   → enriches Vector DB embeddings for future pattern detection
```

### Key patterns

- **Non-linear state graph** — investigation branches and rejoins based on evidence; replanning is first-class
- **MCP / Connector Runtime** — all external systems (Prometheus, Azure Monitor, Git, ITSM) are behind swappable MCP-style connectors
- **RAG-grounded synthesis** — every conclusion is cited from retrieved evidence to reduce hallucination
- **Checkpoint/resume** — graph position is persisted in Redis for fault tolerance and retries
- **Security by default** — PII/secrets scrubbed *before* LLM processing; least-privilege credentials per connector

### Deployment models

- SaaS (cloud-hosted)
- Customer VPC
- On-premises with optional local LLM (Hybrid Data Bridge handles outbound-only connectivity)

## Planned Tech Stack

- **LLM:** Claude (Anthropic) as the primary reasoning engine
- **Orchestration:** LangGraph-style state graph
- **Transport:** WebSockets (streaming progress) + HTTPS
- **Bridge:** gRPC (on-prem / hybrid)
- **Queue/state:** Redis
- **Vector store:** Milvus or Chroma
