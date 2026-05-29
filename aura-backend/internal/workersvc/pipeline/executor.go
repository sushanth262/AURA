package pipeline

import (
	"context"
	"errors"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	orchregistry "github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

// Stage records pipeline stage names for tests.
type Stage string

const (
	StageMCP      Stage = "mcp"
	StageRAG      Stage = "rag"
	StageSecurity Stage = "security"
)

// MCPClient fetches connector payloads (stub: YAML fixtures).
type MCPClient interface {
	Fetch(ctx context.Context, connectorID, fixtureKey string) (map[string]any, error)
}

// RAGClient retrieves embedding context (stub).
type RAGClient interface {
	Retrieve(ctx context.Context, query orchestration.RAGQuery) ([]map[string]any, error)
}

// SecurityClient redacts payloads before findings (stub: pass-through).
type SecurityClient interface {
	Redact(ctx context.Context, raw map[string]any, source string) (map[string]any, error)
}

// Executor runs MCP → RAG → Security → findings for one agent task.
type Executor struct {
	MCP          MCPClient
	RAG          RAGClient
	Security     SecurityClient
	DefaultTenant string
	Now          func() time.Time
	OnStage      func(Stage)
}

// Run executes the pipeline and returns an AgentResult.
func (e *Executor) Run(ctx context.Context, task orchestration.AgentTask, spec orchregistry.AgentDefinition) (orchestration.AgentResult, error) {
	if e.Now == nil {
		e.Now = func() time.Time { return time.Now().UTC() }
	}
	start := e.Now()

	if err := validateConnectors(task, spec); err != nil {
		return orchestration.AgentResult{
			Domain:      task.Domain,
			TaskID:      task.TaskID,
			Status:      orchestration.AgentFailed,
			CompletedAt: e.Now(),
		}, err
	}

	connectors := task.Connectors
	if len(connectors) == 0 {
		connectors = spec.Connectors
	}

	var connectorSnaps = make(map[string]map[string]any)
	var lastSnap map[string]any
	for _, conn := range connectors {
		if !orchregistry.ConnectorAllowed(spec, conn) {
			continue
		}
		e.stage(StageMCP)
		raw, err := e.MCP.Fetch(ctx, conn, task.FixtureKey)
		if err != nil {
			return orchestration.AgentResult{
				Domain:      task.Domain,
				TaskID:      task.TaskID,
				Status:      orchestration.AgentFailed,
				CompletedAt: e.Now(),
			}, err
		}
		if len(raw) > 0 {
			connectorSnaps[conn] = raw
			lastSnap = raw
		}
	}

	e.stage(StageRAG)
	tenantID := task.TenantID
	if tenantID == "" {
		tenantID = e.DefaultTenant
	}
	if tenantID == "" {
		tenantID = "demo"
	}
	ragHits, _ := e.RAG.Retrieve(ctx, orchestration.RAGQuery{
		Namespaces: spec.RAGNamespaces,
		IncidentID: task.IncidentID,
		TenantID:   tenantID,
		QueryText:  task.Instructions,
	})

	e.stage(StageSecurity)
	redacted := lastSnap
	if lastSnap != nil {
		if e.Security == nil {
			return orchestration.AgentResult{
				Domain: task.Domain, TaskID: task.TaskID,
				Status: orchestration.AgentFailed, CompletedAt: e.Now(),
			}, errSecurityRequired
		}
		var err error
		redacted, err = e.Security.Redact(ctx, lastSnap, primaryConnector(spec))
		if err != nil {
			return orchestration.AgentResult{
				Domain:      task.Domain,
				TaskID:      task.TaskID,
				Status:      orchestration.AgentFailed,
				CompletedAt: e.Now(),
			}, err
		}
	}

	findings := buildFindings(task.Domain, connectorSnaps)
	if len(ragHits) > 0 && len(findings) > 0 {
		evidence := make([]any, len(ragHits))
		for i, hit := range ragHits {
			evidence[i] = hit
		}
		findings[0].SupportingEvidence = evidence
	}
	return orchestration.AgentResult{
		Domain:               task.Domain,
		TaskID:               task.TaskID,
		Status:               orchestration.AgentSuccess,
		Findings:             findings,
		FindingCount:         len(findings),
		ConnectorSnapshot:    redacted,
		ExecutionDurationMS:  e.Now().Sub(start).Milliseconds(),
		CompletedAt:          e.Now(),
	}, nil
}

func (e *Executor) stage(s Stage) {
	if e.OnStage != nil {
		e.OnStage(s)
	}
}

func validateConnectors(task orchestration.AgentTask, spec orchregistry.AgentDefinition) error {
	for _, conn := range task.Connectors {
		if !orchregistry.ConnectorAllowed(spec, conn) {
			return errConnectorDenied{agent: string(task.Domain), connector: conn}
		}
	}
	return nil
}

func primaryConnector(spec orchregistry.AgentDefinition) string {
	if len(spec.Connectors) > 0 {
		return spec.Connectors[0]
	}
	return ""
}

type errConnectorDenied struct {
	agent, connector string
}

var errSecurityRequired = errors.New("security client required")

// ErrConnectorDenied is returned when a task requests a connector outside the agent allowlist.
type ErrConnectorDenied = errConnectorDenied

func (e errConnectorDenied) Error() string {
	return "connector " + e.connector + " not allowed for agent " + e.agent
}

func buildFindings(domain orchestration.AgentDomain, connectorSnaps map[string]map[string]any) []orchestration.Finding {
	now := time.Now().UTC().Format(time.RFC3339)
	switch domain {
	case orchestration.DomainTelemetry:
		return []orchestration.Finding{
			{
				FindingID:   "f-metric-1",
				Domain:      domain,
				Type:        "METRIC_ANOMALY",
				Description: "Error rate spike detected correlated with incident window.",
				Confidence:  0.82,
				TimelineTS:  now,
			},
			{
				FindingID:   "f-metric-2",
				Domain:      domain,
				Type:        "LOG_ERROR_BURST",
				Description: "Elevated error burst in service metrics.",
				Confidence:  0.76,
				TimelineTS:  now,
			},
		}
	case orchestration.DomainCode:
		return []orchestration.Finding{
			{
				FindingID:   "f-code-1",
				Domain:      domain,
				Type:        "DEPLOY_CORRELATION",
				Description: "Recent deployment identified as potential regression source.",
				Confidence:  0.74,
				TimelineTS:  now,
			},
		}
	case orchestration.DomainCommunications:
		return buildCommunicationsFindings(connectorSnaps)
	default:
		return []orchestration.Finding{
			{
				FindingID:   "f-ctx-1",
				Domain:      domain,
				Type:        "REGRESSION_TICKET",
				Description: "Related ITSM ticket found in incident window.",
				Confidence:  0.70,
				TimelineTS:  now,
			},
		}
	}
}
