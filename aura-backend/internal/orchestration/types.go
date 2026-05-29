package orchestration

import "time"

// AgentDomain identifies a worker lane in the investigation graph.
type AgentDomain string

const (
	DomainSupervisor AgentDomain = "supervisor"
	DomainTelemetry  AgentDomain = "telemetry"
	DomainCode       AgentDomain = "code"
	DomainContext    AgentDomain = "context"
)

// InvestigationStatus matches aura-frontend InvestigationStatus / supervisor store values.
type InvestigationStatus string

const (
	StatusQueued          InvestigationStatus = "QUEUED"
	StatusIntake          InvestigationStatus = "INTAKE"
	StatusPlanning        InvestigationStatus = "PLANNING"
	StatusRetrieving      InvestigationStatus = "RETRIEVING"
	StatusSynthesis       InvestigationStatus = "SYNTHESIS"
	StatusHITLPending     InvestigationStatus = "HITL_PENDING"
	StatusReplanning      InvestigationStatus = "REPLANNING"
	StatusPartialEvidence InvestigationStatus = "PARTIAL_EVIDENCE"
	StatusFailed          InvestigationStatus = "FAILED"
)

// AgentResultStatus is the per-agent execution outcome.
type AgentResultStatus string

const (
	AgentSuccess AgentResultStatus = "SUCCESS"
	AgentPartial AgentResultStatus = "PARTIAL"
	AgentFailed  AgentResultStatus = "FAILED"
	AgentSkipped AgentResultStatus = "SKIPPED"
)

// AgentTask is dispatched from the supervisor graph to a worker agent (Phase 3+).
type AgentTask struct {
	AgentTaskID string         `json:"agent_task_id"`
	IncidentID  string         `json:"incident_id"`
	TaskID      string         `json:"task_id"`
	Domain      AgentDomain    `json:"domain"`
	Instructions string        `json:"instructions,omitempty"`
	TimeWindow  map[string]any `json:"time_window,omitempty"`
	Connectors  []string       `json:"connectors,omitempty"`
	FixtureKey  string         `json:"fixture_key,omitempty"`
}

// Finding is a single grounded insight from an agent.
type Finding struct {
	FindingID          string         `json:"finding_id"`
	Domain             AgentDomain    `json:"domain"`
	Type               string         `json:"type"`
	Description        string         `json:"description"`
	Confidence         float64        `json:"confidence"`
	SupportingEvidence []any          `json:"supporting_evidence,omitempty"`
	TimelineTS         string         `json:"timeline_ts,omitempty"`
}

// AgentResult is returned after an agent completes (stub or live).
type AgentResult struct {
	Domain               AgentDomain       `json:"domain"`
	AgentID              string            `json:"agent_id,omitempty"`
	TaskID               string            `json:"task_id"`
	Status               AgentResultStatus `json:"status"`
	Findings             []Finding         `json:"findings,omitempty"`
	ExecutionDurationMS  int64             `json:"execution_duration_ms,omitempty"`
	CompletedAt          time.Time         `json:"completed_at"`
	ConnectorSnapshot    map[string]any    `json:"connector_snapshot,omitempty"`
	FindingCount         int               `json:"finding_count,omitempty"`
}

// ProgressEvent is emitted on the investigation WebSocket stream.
type ProgressEvent struct {
	TaskID      string         `json:"task_id"`
	IncidentID  string         `json:"incident_id"`
	EventType   string         `json:"event_type"`
	AgentDomain AgentDomain    `json:"agent_domain,omitempty"`
	Payload     map[string]any `json:"payload"`
	Timestamp   string         `json:"timestamp"`
	SequenceNum int            `json:"sequence_num"`
}

// GraphCheckpoint persists orchestration progress (in-memory or Redis later).
type GraphCheckpoint struct {
	TaskID          string              `json:"task_id"`
	IncidentID      string              `json:"incident_id"`
	CurrentStatus   InvestigationStatus `json:"current_status"`
	CompletedNodes  []string            `json:"completed_nodes"`
	FailedNodes     []string            `json:"failed_nodes"`
	GraphVersion    int                 `json:"graph_version"`
	LastUpdated     time.Time           `json:"last_updated"`
	PartialEvidence map[string]any      `json:"partial_evidence,omitempty"`
}
