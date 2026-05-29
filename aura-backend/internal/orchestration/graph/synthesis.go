package graph

import (
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)
// RunContext carries incident fields needed for stub synthesis payloads.
type RunContext struct {
	TaskID         string
	IncidentID     string
	Title          string
	Severity       string
	Symptoms       string
	Scope          map[string]any
	FixtureKey     string
	TimelineLabels []string
	ScenarioID     string
	PolicyVersion  string
	Status         string
	SynthesisLLMMode string
	TenantID       string
}

func scopeService(scope map[string]any) string {
	if s, ok := scope["service"].(string); ok && s != "" {
		return s
	}
	return "unknown-service"
}

func (rc RunContext) timelineMessage(index int) string {
	if index >= 0 && index < len(rc.TimelineLabels) {
		return rc.TimelineLabels[index]
	}
	return "timeline_step"
}

// BuildSynthesisPayload mirrors supervisor mocktimeline SYNTHESIS_COMPLETE payload (legacy stub).
// Prefer FuseSynthesis when agent results are available.
func BuildSynthesisPayload(rc RunContext) map[string]any {
	return FuseSynthesis(rc, nil, nil)
}

// DefaultFindingCount for stub AGENT_COMPLETE payloads.
func DefaultFindingCount(domain orchestration.AgentDomain) int {
	switch domain {
	case orchestration.DomainTelemetry:
		return 2
	case orchestration.DomainCommunications:
		return 3
	default:
		return 1
	}
}

// DefaultStartProgressPct for stub AGENT_STARTED payloads.
func DefaultStartProgressPct(domain orchestration.AgentDomain) int {
	switch domain {
	case orchestration.DomainTelemetry:
		return 52
	case orchestration.DomainCode:
		return 21
	case orchestration.DomainContext:
		return 68
	case orchestration.DomainCommunications:
		return 45
	default:
		return 10
	}
}
