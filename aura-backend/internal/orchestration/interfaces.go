package orchestration

import "context"

// ConnectorRuntime invokes MCP-style tools against external systems.
type ConnectorRuntime interface {
	Invoke(ctx context.Context, call ConnectorCall) (map[string]any, error)
}

// ConnectorCall describes a single connector invocation.
type ConnectorCall struct {
	ConnectorID string
	ScenarioKey string
	Query       map[string]any
}

// RAGClient retrieves embedding-backed context for an agent.
type RAGClient interface {
	Retrieve(ctx context.Context, query RAGQuery) ([]map[string]any, error)
}

// RAGQuery scopes retrieval to namespaces and incident context.
type RAGQuery struct {
	Namespaces  []string
	IncidentID  string
	QueryText   string
}

// SecurityClient redacts raw connector payloads before synthesis/LLM.
type SecurityClient interface {
	Redact(ctx context.Context, raw map[string]any, source string) (map[string]any, error)
}

// AgentExecutor runs one domain agent through MCP → RAG → security (Phase 3+).
type AgentExecutor interface {
	Domain() AgentDomain
	RequiredConnectors() []string
	Execute(ctx context.Context, task AgentTask) (AgentResult, error)
}

// AgentWorkerClient dispatches AgentTask to aura-worker (Phase 3+).
type AgentWorkerClient interface {
	Execute(ctx context.Context, domain AgentDomain, task AgentTask) (AgentResult, error)
}

// SnapshotFetcher pulls fixture connector payloads (supervisor → worker today).
type SnapshotFetcher interface {
	Fetch(ctx context.Context, source string, scenarioKey string) (map[string]any, error)
}
