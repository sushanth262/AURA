package graph

import (
	"fmt"
	"time"

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

// BuildSynthesisPayload mirrors supervisor mocktimeline SYNTHESIS_COMPLETE payload.
func BuildSynthesisPayload(rc RunContext) map[string]any {
	svc := scopeService(rc.Scope)
	now := time.Now().UTC().Format(time.RFC3339)
	findings := []map[string]any{
		{
			"finding_id":          "f-metric-1",
			"domain":              orchestration.DomainTelemetry,
			"type":                "METRIC_ANOMALY",
			"description":         fmt.Sprintf("Error rate spike detected in %s correlated with incident window.", svc),
			"confidence":          0.82,
			"supporting_evidence": []map[string]any{},
			"timeline_ts":         now,
		},
		{
			"finding_id":          "f-code-1",
			"domain":              orchestration.DomainCode,
			"type":                "DEPLOY_CORRELATION",
			"description":         fmt.Sprintf("Recent deployment to %s identified as potential regression source.", svc),
			"confidence":          0.74,
			"supporting_evidence": []map[string]any{},
			"timeline_ts":         now,
		},
	}
	return map[string]any{
		"findings":           findings,
		"narrative":          buildNarrativeStructured(rc),
		"confidence_score":   0.78,
		"policy_version":     rc.PolicyVersion,
		"wireframe_fixture":  rc.ScenarioID,
	}
}

func buildNarrativeStructured(rc RunContext) map[string]any {
	svc := scopeService(rc.Scope)
	return map[string]any{
		"report_metadata": map[string]any{
			"service":   svc,
			"severity":  rc.Severity,
			"status":    rc.Status,
			"rca_topic": rc.Title,
		},
		"symptoms": rc.Symptoms,
		"agent_findings": []map[string]any{
			{
				"agent_name":  "Telemetry Agent",
				"focus":       "Metrics & Logs",
				"observation": fmt.Sprintf("Detected anomalous error-rate and latency patterns in %s during the incident window. Elevated 5xx ratios and p99 latency spikes are consistent with the reported symptoms.", svc),
			},
			{
				"agent_name":  "Code Agent",
				"focus":       "Deployments & Commits",
				"observation": fmt.Sprintf("Correlated recent deployments and commits touching %s with the degradation timeline. Identified changes that align with the onset of the incident.", svc),
			},
			{
				"agent_name":  "Context Agent",
				"focus":       "Tickets & Runbooks",
				"observation": fmt.Sprintf("Cross-referenced ITSM tickets and runbooks for %s. Distinguished intentional operational changes from potential regressions introduced by recent releases.", svc),
			},
		},
		"conclusion": map[string]any{
			"summary":          fmt.Sprintf("Evidence across all three agents suggests a recent change to %s introduced a regression correlated with the reported symptoms.", svc),
			"confidence_level": "78%",
			"action_item":      fmt.Sprintf("Review the identified commit(s) affecting %s and consider a rollback pending operator approval. Validate error rates recover after remediation.", svc),
		},
	}
}

// DefaultFindingCount for stub AGENT_COMPLETE payloads.
func DefaultFindingCount(domain orchestration.AgentDomain) int {
	switch domain {
	case orchestration.DomainTelemetry:
		return 2
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
