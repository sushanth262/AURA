# AURA — Web UI Wireframes

Covers the five primary screens of the AURA Web UI (§3 of `PRODUCTION_SPECIFICATIONS.md`).

---

## Screen 1 — Incident Intake

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AURA                                    [Incident History]  [● dsushanth ▾]│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   New Investigation                                                          │
│   ─────────────────────────────────────────────────────────────────────     │
│                                                                              │
│   Title *                                                                    │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  e.g. "API gateway 5xx spike — us-east-1"                           │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   Severity *                    Scope (service / cluster / region) *        │
│   ┌───────────────────────┐     ┌───────────────────────────────────────┐   │
│   │  P2 — Major        ▾  │     │  payments-api / prod-us-east-1        │   │
│   └───────────────────────┘     └───────────────────────────────────────┘   │
│                                                                              │
│   Time Window                                                                │
│   ┌───────────────────────────┐   to   ┌───────────────────────────────┐    │
│   │  2026-05-03  14:30        │        │  2026-05-03  15:00            │    │
│   └───────────────────────────┘        │  (leave blank if ongoing)     │    │
│                                        └───────────────────────────────┘    │
│                                                                              │
│   Symptoms *                                              0 / 2000           │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Describe what you observed. Include error messages, affected       │   │
│   │  user segments, and any recent changes you are aware of.            │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   Artifacts  (optional)                            [+ Add artifact]          │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Type             Source              Content preview               │   │
│   │  ─────────────────────────────────────────────────────────────     │   │
│   │  STACK_TRACE      Datadog alert       java.lang.NullPointer…  [✕]  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   ─────────────────────────────────────────────────────────────────────     │
│                               [Cancel]    [Start Investigation  →]          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Notes**
- Severity drives agent parallelism priority; P1 jumps the queue.
- Time window `end = null` signals "ongoing" to the Supervisor.
- Artifacts expand in a drawer; each limited to 10 KB.
- "Start Investigation" POSTs to `/api/incidents` and navigates immediately to Screen 2 with the returned `task_id`.

---

## Screen 2 — Live Investigation Progress

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AURA                                    [Incident History]  [● dsushanth ▾]│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ← Back    API gateway 5xx spike — us-east-1           P2  │ INVESTIGATING  │
│            INC-2847  ·  Started 14:32  ·  us-east-1                         │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  ┌─────────────────────────────────────────────────┐  ┌───────────────────┐ │
│  │  Investigation Timeline                   LIVE ● │  │  Quick Summary    │ │
│  │  ─────────────────────────────────────────────  │  │  ─────────────── │ │
│  │                                                  │  │                   │ │
│  │  ✅ 14:32:01  Task claimed by supervisor         │  │  Agents running   │ │
│  │  ✅ 14:32:02  Graph planned  (3 parallel agents) │  │  ┌─────────────┐  │ │
│  │                                                  │  │  │ Telemetry ● │  │ │
│  │  ⠿ 14:32:03  Telemetry agent  —  querying…      │  │  │ Code      ● │  │ │
│  │  ⠿ 14:32:03  Code agent       —  querying…      │  │  │ Context   ● │  │ │
│  │  ⠿ 14:32:03  Context agent    —  querying…      │  │  └─────────────┘  │ │
│  │                                                  │  │                   │ │
│  │                                                  │  │  Elapsed          │ │
│  │                                                  │  │  00:01:12         │ │
│  │                                                  │  │                   │ │
│  │                                                  │  │  Est. remaining   │ │
│  │                                                  │  │  ~01:48           │ │
│  └─────────────────────────────────────────────────┘  └───────────────────┘ │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  Agent Activity                                                        │  │
│  │  ─────────────────────────────────────────────────────────────────   │  │
│  │                                                                        │  │
│  │  📡 Telemetry   ██████████░░░░░░░░░░  52%   Querying Prometheus…     │  │
│  │  💻 Code        ████░░░░░░░░░░░░░░░░  21%   Fetching recent commits… │  │
│  │  📄 Context     ██████████████░░░░░░  68%   Searching Jira tickets…  │  │
│  │                                                                        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Notes**
- Page subscribes to `WS /ws/investigations/{task_id}` on load.
- Timeline items are appended on each `TaskProgressEvent`.
- Spinner `⠿` replaces `✅` until the agent emits `AGENT_COMPLETE`.
- Progress bars are derived from the elapsed/deadline ratio in `AgentTask.deadline`.
- On `SYNTHESIS_COMPLETE` event, the page transitions automatically to Screen 3.

---

## Screen 3 — Evidence Review

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AURA                                    [Incident History]  [● dsushanth ▾]│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ← Back    API gateway 5xx spike — us-east-1           P2  │ AWAITING REVIEW│
│            INC-2847  ·  Completed 14:34:51  ·  us-east-1                    │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  Confidence                                                            │  │
│  │  ────────────────────────────────────────────────────────────────    │  │
│  │                                                                        │  │
│  │   ████████████████████░░░░  0.81 / 1.0   HIGH                         │  │
│  │                                                                        │  │
│  │   Citation strength   0.85  ·  Agent agreement  0.90  ·               │  │
│  │   Memory match boost  +0.08  ·  Rejection penalty  0.00               │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  Diagnostic Narrative                                                  │  │
│  │  ─────────────────────────────────────────────────────────────────   │  │
│  │                                                                        │  │
│  │  A 4× increase in HTTP 5xx errors began at 14:31 UTC on the           │  │
│  │  payments-api service in us-east-1. Telemetry confirms the onset      │  │
│  │  correlates with deploy payments-api@v2.14.1 at 14:29 UTC [¹].       │  │
│  │  Commit a3f9c2d introduced a nil-pointer dereference in the           │  │
│  │  PaymentProcessor.charge() path [²]. A similar regression occurred   │  │
│  │  in INC-2291 (2026-03-11) with an identical stack trace pattern [³].  │  │
│  │                                                                        │  │
│  │  [¹] Prometheus · [²] GitHub commit · [³] Prior incident INC-2291    │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────┐  │
│  │  📡 Telemetry    │  │  💻 Code         │  │  📄 Context              │  │
│  │  ─────────────  │  │  ─────────────   │  │  ────────────────────    │  │
│  │  METRIC_ANOMALY  │  │  COMMIT_         │  │  REGRESSION_TICKET       │  │
│  │  p95 latency     │  │  REGRESSION      │  │  PAY-4821 "5xx on        │  │
│  │  780ms → 4.2s    │  │  a3f9c2d         │  │  charge endpoint"        │  │
│  │  onset 14:31     │  │  author: k.lee   │  │  opened 14:33            │  │
│  │                  │  │  conf: 0.88      │  │                          │  │
│  │  ERROR_BURST     │  │                  │  │  RUNBOOK_MATCH           │  │
│  │  NullPointer     │  │  DEPLOY_         │  │  "Payment 5xx            │  │
│  │  442/min         │  │  CORRELATION     │  │  Runbook" §3.2  0.91    │  │
│  │                  │  │  v2.14.1 @14:29  │  │                          │  │
│  └──────────────────┘  └──────────────────┘  └──────────────────────────┘  │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Root Cause Candidates                                                       │
│  ─────────────────────────────────────────────────────────────────────────  │
│  1. ● Nil-pointer in PaymentProcessor.charge() introduced in a3f9c2d  0.88 │
│  2. ○ Connection pool exhaustion under increased load                   0.32 │
│                                                                              │
│  Recommended Actions                                                         │
│  ─────────────────────────────────────────────────────────────────────────  │
│  • Rollback payments-api to v2.13.9                                         │
│  • Apply hotfix per runbook PAY-5xx §3.2                                    │
│  • Open regression ticket against commit a3f9c2d                            │
│                                                                              │
│  Prior Incident Match   INC-2291 (2026-03-11)  similarity 0.91   [View →]   │
│                                                                              │
│                                       [Reject ↩]   [Approve & Remediate →] │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Notes**
- Confidence bar is color-coded: green ≥ 0.75, amber 0.5–0.74, red < 0.5.
- Footnote references `[¹]` `[²]` `[³]` in the narrative are clickable deep-links to the originating evidence record.
- Agent panels are tabs on mobile viewports.
- Root cause candidates are ranked by descending confidence; non-primary candidates are collapsed by default.
- "Reject ↩" opens Screen 4a; "Approve & Remediate →" opens Screen 4b.

---

## Screen 4a — HITL Rejection

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AURA                                    [Incident History]  [● dsushanth ▾]│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ← Back to Evidence                                                          │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                                                                       │   │
│  │   Reject Investigation                                                │   │
│  │   ─────────────────────────────────────────────────────────────     │   │
│  │                                                                       │   │
│  │   INC-2847  ·  API gateway 5xx spike — us-east-1                     │   │
│  │                                                                       │   │
│  │   Rejection reason *                         0 / 500                 │   │
│  │   ┌─────────────────────────────────────────────────────────────┐   │   │
│  │   │  Describe why the diagnosis is incorrect or incomplete.      │   │
│  │   │  The supervisor will use this to replan the investigation.   │   │
│  │   │                                                              │   │
│  │   └─────────────────────────────────────────────────────────────┘   │   │
│  │                                                                       │   │
│  │   Common reasons                                                      │   │
│  │   [ ] Wrong service identified                                        │   │
│  │   [ ] Incorrect time window                                           │   │
│  │   [ ] Missing data source — specify below                             │   │
│  │   [ ] Root cause already known                                        │   │
│  │                                                                       │   │
│  │                    [Cancel]    [Send back for replanning  ↩]         │   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Notes**
- POSTs `{ decision: "REJECTED", reason: "..." }` to `/api/investigations/{task_id}/hitl`.
- On submit, supervisor state transitions `HITL_PENDING → REPLANNING`.
- The page navigates back to Screen 2 (live progress) to show the replan.

---

## Screen 4b — HITL Approval & Remediation Trigger

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AURA                                    [Incident History]  [● dsushanth ▾]│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ← Back to Evidence                                                          │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                                                                       │   │
│  │   Approve & Remediate                                                 │   │
│  │   ─────────────────────────────────────────────────────────────     │   │
│  │                                                                       │   │
│  │   INC-2847  ·  API gateway 5xx spike — us-east-1  ·  Confidence 0.81 │   │
│  │                                                                       │   │
│  │   Select actions to execute *                                         │   │
│  │                                                                       │   │
│  │   ┌─────────────────────────────────────────────────────────────┐   │   │
│  │   │  ☑  Rollback payments-api to v2.13.9                         │   │
│  │   │     Automated · Est. 3 min · Reversible                      │   │
│  │   │                                                              │   │
│  │   │  ☑  Open regression ticket for commit a3f9c2d               │   │
│  │   │     Automated · Creates Jira ticket · Irreversible           │   │
│  │   │                                                              │   │
│  │   │  ☐  Apply hotfix per runbook PAY-5xx §3.2                   │   │
│  │   │     Manual · Requires on-call confirmation                   │   │
│  │   └─────────────────────────────────────────────────────────────┘   │   │
│  │                                                                       │   │
│  │   ⚠  You are about to execute automated changes in production.       │   │
│  │      This action is logged and attributed to dsushanth@…             │   │
│  │                                                                       │   │
│  │   Confirm your identity to proceed                                    │   │
│  │   ┌─────────────────────────────────────┐                            │   │
│  │   │  Enter your password or use SSO  🔑 │                            │   │
│  │   └─────────────────────────────────────┘                            │   │
│  │                                                                       │   │
│  │   [Cancel]                      [Execute selected actions  ✓]        │   │
│  │                                                                       │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Notes**
- Re-authentication step satisfies the §3.5 HITL elevated-scope requirement.
- Each action shows reversibility and automation level so operators can make informed selections.
- POSTs `{ decision: "APPROVED" }` then `{ approved_action_ids: [...] }` to the remediation endpoint.
- On execution, navigates to Screen 5.

---

## Screen 5 — Remediation in Progress & Completion

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AURA                                    [Incident History]  [● dsushanth ▾]│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  API gateway 5xx spike — us-east-1           P2  │ ✅ RESOLVED              │
│  INC-2847  ·  Resolved 14:41:03  ·  us-east-1                               │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  Remediation Log                                                       │  │
│  │  ─────────────────────────────────────────────────────────────────   │  │
│  │                                                                        │  │
│  │  ✅ 14:39:11  Rollback initiated — payments-api → v2.13.9             │  │
│  │  ✅ 14:40:44  Rollback complete  — all pods healthy                   │  │
│  │  ✅ 14:41:01  Jira ticket created — PAY-4824 (a3f9c2d regression)     │  │
│  │  ✅ 14:41:03  Incident memory updated in Vector DB                    │  │
│  │                                                                        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  Investigation Summary                                                 │  │
│  │  ─────────────────────────────────────────────────────────────────   │  │
│  │                                                                        │  │
│  │  Root cause    Nil-pointer in PaymentProcessor.charge() (a3f9c2d)     │  │
│  │  Confidence    0.81                                                    │  │
│  │  Time to HITL  2m 51s                                                  │  │
│  │  Reviewer      dsushanth                                               │  │
│  │  Decision      APPROVED                                                │  │
│  │  Resolved at   14:41:03 UTC                                            │  │
│  │                                                                        │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ─────────────────────────────────────────────────────────────────────────  │
│                              [View full evidence]   [New investigation  +]  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen 6 — Incident History

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  AURA                                    [+ New Investigation]  [● dsushanth▾│
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Incident History                                                            │
│  ─────────────────────────────────────────────────────────────────────────  │
│                                                                              │
│  Search  ┌───────────────────────────────────────┐  Severity [All ▾]        │
│          │  🔍  Filter by title or service…       │  Status   [All ▾]        │
│          └───────────────────────────────────────┘  Date     [Last 30d ▾]   │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  ID        Title                              Sev  Status   Conf  Age│   │
│  │  ─────────────────────────────────────────────────────────────────  │   │
│  │  INC-2847  API gateway 5xx — us-east-1         P2  Resolved  0.81  2h│   │
│  │  INC-2831  Auth service latency spike          P2  Resolved  0.74  1d│   │
│  │  INC-2819  Checkout timeout — EU region        P1  Resolved  0.93  3d│   │
│  │  INC-2801  Worker pool queue depth alert       P3  Resolved  0.68  5d│   │
│  │  INC-2790  DB connection pool exhaustion       P1  Resolved  0.89  7d│   │
│  │  INC-2776  Cache invalidation storm            P2  Rejected  0.41  9d│   │
│  │  INC-2762  Memory leak — recommendation-svc    P3  Resolved  0.72 12d│   │
│  │  ...                                                                  │   │
│  │                                                                       │   │
│  │  ← Previous   Page 1 of 8   Next →                                   │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Notes**
- Rows are clickable and navigate to the Evidence Review (Screen 3) for that incident.
- "Rejected" status shown with amber styling; HITL-rejected incidents can be re-opened.
- Confidence column is absent for in-progress investigations.

---

## Navigation Flow

```
Screen 1 (Intake)
    │  POST /api/incidents → task_id
    ▼
Screen 2 (Live Progress)
    │  WebSocket SYNTHESIS_COMPLETE event
    ▼
Screen 3 (Evidence Review)
    ├──[Reject ↩]──────────────────────────────────────────┐
    │                                                       ▼
    │                                            Screen 4a (Rejection)
    │                                                       │  replan
    │                                                       └──► Screen 2
    │
    └──[Approve →]──► Screen 4b (Approval + Remediation)
                                │  execute
                                ▼
                      Screen 5 (Resolved)
                                │
                                └──► Screen 6 (History)  ◄── Nav link (any screen)
```
