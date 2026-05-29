package registry

import "github.com/sushanth262/AURA/aura-backend/internal/orchestration"

// AgentDefinition describes a built-in agent and its connectors.
type AgentDefinition struct {
	Domain        orchestration.AgentDomain
	Label         string
	Color         string
	Connectors    []string
	RAGNamespaces []string
	Enabled       bool
}

// SupervisorLane returns UI metadata for the supervisor swimlane.
func SupervisorLane() AgentDefinition {
	return AgentDefinition{
		Domain: orchestration.DomainSupervisor,
		Label:  "Supervisor",
		Color:  "#1B2B65",
	}
}

// builtinAgents is the canonical catalog (Phase 2 adds communications).
var builtinAgents = []AgentDefinition{
	{
		Domain:        orchestration.DomainTelemetry,
		Label:         "Telemetry / RCA",
		Color:         "#3B82F6",
		Connectors:    []string{"grafana"},
		RAGNamespaces: []string{"incident_memory", "runbooks"},
	},
	{
		Domain:        orchestration.DomainCode,
		Label:         "Code / Fix",
		Color:         "#8B5CF6",
		Connectors:    []string{"github"},
		RAGNamespaces: []string{"source_code", "incident_memory"},
	},
	{
		Domain:        orchestration.DomainContext,
		Label:         "Context / Docs",
		Color:         "#10B981",
		Connectors:    []string{"jira"},
		RAGNamespaces: []string{"runbooks", "incident_memory"},
	},
	{
		Domain:        orchestration.DomainCommunications,
		Label:         "Communications",
		Color:         "#F59E0B",
		Connectors:    []string{"slack"},
		RAGNamespaces: []string{"incident_memory"},
	},
}

// catalogOrder is stable lane order for enabled agents.
var catalogOrder = []orchestration.AgentDomain{
	orchestration.DomainTelemetry,
	orchestration.DomainCode,
	orchestration.DomainContext,
	orchestration.DomainCommunications,
}
