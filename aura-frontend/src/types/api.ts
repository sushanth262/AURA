// TypeScript types derived from aura-frontend/specs/openapi.yaml
// Keep in sync with the spec — do not edit manually.

// ── Enums ────────────────────────────────────────────────────────────────────

export type Severity = 'P1' | 'P2' | 'P3' | 'P4';

export type InvestigationStatus =
  | 'QUEUED'
  | 'INTAKE'
  | 'PLANNING'
  | 'RETRIEVING'
  | 'SYNTHESIS'
  | 'HITL_PENDING'
  | 'REPLANNING'
  | 'REMEDIATION'
  | 'MEMORY_WRITEBACK'
  | 'COMPLETE'
  | 'PARTIAL_EVIDENCE'
  | 'FAILED';

export type AgentDomain = 'telemetry' | 'code' | 'context' | 'supervisor';

export type ArtifactType =
  | 'STACK_TRACE'
  | 'LOG_EXCERPT'
  | 'ALERT_PAYLOAD'
  | 'METRIC_SNAPSHOT'
  | 'OTHER';

export type FindingType =
  | 'METRIC_ANOMALY'
  | 'LOG_ERROR_BURST'
  | 'SATURATION_EVENT'
  | 'REGIONAL_ANOMALY'
  | 'COMMIT_REGRESSION_CANDIDATE'
  | 'DEPLOY_CORRELATION'
  | 'CODE_PATTERN_MATCH'
  | 'FIX_PROPOSAL'
  | 'INTENTIONAL_CHANGE'
  | 'REGRESSION_TICKET'
  | 'RUNBOOK_MATCH'
  | 'RFC_CONTEXT';

export type HITLDecisionEnum = 'APPROVED' | 'REJECTED';

export type RejectionCategory =
  | 'WRONG_SERVICE'
  | 'INCORRECT_TIME_WINDOW'
  | 'MISSING_DATA_SOURCE'
  | 'ROOT_CAUSE_KNOWN'
  | 'OTHER';

export type ActionRisk = 'Low' | 'Med' | 'High';
export type ActionAutomation = 'Automated' | 'Manual';
export type AgentStatus = 'SUCCESS' | 'PARTIAL' | 'FAILED' | 'SKIPPED';

export type WSEventType =
  | 'TASK_CLAIMED'
  | 'AGENT_STARTED'
  | 'AGENT_COMPLETE'
  | 'SYNTHESIS_STARTED'
  | 'SYNTHESIS_COMPLETE'
  | 'HITL_REQUESTED'
  | 'HITL_RESOLVED'
  | 'REMEDIATION_STARTED'
  | 'REMEDIATION_COMPLETE'
  | 'TASK_FAILED';

// ── Common Objects ────────────────────────────────────────────────────────────

export interface TimeWindow {
  start: string;       // ISO-8601
  end:   string | null; // null = ongoing
}

export interface Scope {
  service:  string;
  cluster?: string | null;
  region?:  string | null;
}

export interface Artifact {
  artifact_type: ArtifactType;
  content:       string;       // max 10 240 bytes
  source?:       string | null;
}

export interface EvidenceRef {
  ref_id:        string;
  source_type:   string;
  source_id:     string;
  display_label: string;
  url?:          string | null;
  metadata?:     Record<string, unknown>;
}

export interface Finding {
  finding_id:          string;
  domain:              AgentDomain;
  type:                FindingType;
  description:         string;
  confidence:          number;
  supporting_evidence: EvidenceRef[];
  timeline_ts?:        string | null;
}

export interface ConfidenceBreakdown {
  citation_strength:    number;
  agent_agreement:      number;
  memory_match_boost:   number;
  rejection_penalty:    number;
}

export interface RootCauseCandidate {
  candidate_id: string;
  description:  string;
  confidence:   number;
  is_primary:   boolean;
  citations:    EvidenceRef[];
}

export interface RecommendedAction {
  action_id:                  string;
  description:                string;
  automation:                 ActionAutomation;
  reversible:                 boolean;
  risk:                       ActionRisk;
  estimated_duration_seconds?: number | null;
  runbook_ref?:               string | null;
}

export interface AgentSummary {
  domain:                 AgentDomain;
  summary:                string;
  finding_count:          number;
  status:                 AgentStatus;
  execution_duration_ms?: number | null;
}

export interface PriorIncidentMatch {
  incident_id:      string;
  title?:           string;
  resolved_at?:     string | null;
  similarity_score: number;
}

export interface RemediationLogEntry {
  timestamp:   string;
  description: string;
  status:      'complete' | 'failed';
}

export interface HistoryStats {
  total_count:                    number;
  avg_time_to_diagnose_seconds?:  number | null;
  avg_confidence_score?:          number | null;
  resolved_pct?:                  number | null;
}

// ── Request Bodies ────────────────────────────────────────────────────────────

export interface IncidentSubmission {
  title:       string;
  severity:    Severity;
  scope:       Scope;
  time_window: TimeWindow;
  symptoms:    string;
  artifacts?:  Artifact[];
}

export interface HITLDecision {
  decision:     HITLDecisionEnum;
  reason?:      string | null;
  categories?:  RejectionCategory[];
  reviewer_id:  string;
}

export interface RemediationTrigger {
  approved_action_ids: string[];
}

// ── Response Schemas ──────────────────────────────────────────────────────────

export interface IncidentQueuedResponse {
  task_id:     string;
  incident_id: string;
  status:      'QUEUED';
}

export interface IncidentStateResponse {
  incident_id: string;
  task_id:     string;
  status:      InvestigationStatus;
  severity:    Severity;
  title:       string;
  scope?:      Scope;
  created_at:  string;
  updated_at:  string;
}

export interface IncidentSummary {
  incident_id:       string;
  title:             string;
  severity:          Severity;
  status:            InvestigationStatus;
  confidence_score?: number | null;
  scope?:            Scope;
  created_at:        string;
  resolved_at?:      string | null;
}

export interface IncidentHistoryPage {
  items:    IncidentSummary[];
  page:     number;
  per_page: number;
  total:    number;
  stats:    HistoryStats;
}

export interface EvidenceBundle {
  incident_id:            string;
  task_id:                string;
  narrative:              string;
  confidence_score:       number | null;
  confidence_breakdown:   ConfidenceBreakdown;
  per_agent_summaries:    AgentSummary[];
  agent_findings:         Finding[];
  evidence_refs:          EvidenceRef[];
  root_cause_candidates:  RootCauseCandidate[];
  recommended_actions:    RecommendedAction[];
  prior_incident_matches: PriorIncidentMatch[];
  synthesized_at:         string;
  iteration:              number;
}

export interface HITLResponse {
  status:     'HITL_RESOLVED';
  next_state: 'REMEDIATION' | 'PLANNING';
  task_id:    string;
}

export interface RemediationResponse {
  remediation_task_id: string;
  status:              'QUEUED';
}

export interface InvestigationInProgressResponse {
  status:   InvestigationStatus;
  task_id:  string;
  message?: string;
}

// ── WebSocket Messages ────────────────────────────────────────────────────────

export interface TaskProgressEvent {
  task_id:      string;
  incident_id:  string;
  event_type:   WSEventType;
  agent_domain: AgentDomain | null;
  payload:      Record<string, unknown>;
  timestamp:    string;
  sequence_num: number;
}

export interface WSReplayRequest {
  type:         'REPLAY_FROM';
  sequence_num: number;
}

export interface WSPong {
  type: 'PONG';
}

// ── API error envelope ────────────────────────────────────────────────────────

export interface ApiError {
  error_code: string;
  message:    string;
  details?:   Record<string, string> | null;
}

// ── History filter params ─────────────────────────────────────────────────────

export interface HistoryFilters {
  page?:     number;
  per_page?: number;
  severity?: Severity;
  status?:   InvestigationStatus;
  since?:    string;
  q?:        string;
}
