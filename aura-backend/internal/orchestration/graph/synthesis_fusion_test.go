package graph_test

import (
	"testing"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/graph"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

func baseRC() graph.RunContext {
	return graph.RunContext{
		TaskID: "t1", IncidentID: "i1", Title: "test", Severity: "P2",
		Symptoms: "spike", Scope: map[string]any{"service": "api-gateway"},
		PolicyVersion: "stub-v1",
	}
}

func TestFuseSynthesis_CommsDisabled_OmitsComms(t *testing.T) {
	reg := registry.DefaultRegistry()
	payload := graph.FuseSynthesis(baseRC(), nil, reg)

	findings, _ := payload["findings"].([]map[string]any)
	for _, f := range findings {
		if f["domain"] == string(orchestration.DomainCommunications) {
			t.Fatal("communications finding should be omitted when agent disabled")
		}
	}
	narr, _ := payload["narrative"].(map[string]any)
	agents, _ := narr["agent_findings"].([]map[string]any)
	if len(agents) != 3 {
		t.Fatalf("narrative agents: got %d want 3", len(agents))
	}
}

func TestFuseSynthesis_CommsEnabled_IncludesComms(t *testing.T) {
	reg := registry.New([]string{"telemetry", "code", "context", "communications"})
	results := map[orchestration.AgentDomain]orchestration.AgentResult{
		orchestration.DomainCommunications: {
			Status: orchestration.AgentSuccess,
			Findings: []orchestration.Finding{
				{FindingID: "f-comms-1", Domain: orchestration.DomainCommunications, Type: "CHANNEL_ALERT_MENTION",
					Description: "channel", Confidence: 0.71, TimelineTS: "2026-05-03T14:31:00Z"},
			},
		},
	}
	payload := graph.FuseSynthesis(baseRC(), results, reg)

	foundComms := false
	findings, _ := payload["findings"].([]map[string]any)
	for _, f := range findings {
		if f["domain"] == string(orchestration.DomainCommunications) {
			foundComms = true
		}
	}
	if !foundComms {
		t.Fatal("expected communications finding in synthesis")
	}
	narr, _ := payload["narrative"].(map[string]any)
	agents, _ := narr["agent_findings"].([]map[string]any)
	if len(agents) != 4 {
		t.Fatalf("narrative agents: got %d want 4", len(agents))
	}
}

func TestFuseSynthesis_TimelineOverlapBoostsConfidence(t *testing.T) {
	reg := registry.New([]string{"telemetry", "code", "context", "communications"})
	results := map[orchestration.AgentDomain]orchestration.AgentResult{
		orchestration.DomainTelemetry: {
			Status: orchestration.AgentSuccess,
			ConnectorSnapshot: map[string]any{
				"payload": map[string]any{
					"panels": []any{
						map[string]any{"spike_window": "2026-05-03T14:28:00Z .. 2026-05-03T14:42:00Z"},
					},
				},
			},
		},
		orchestration.DomainCommunications: {
			Status: orchestration.AgentSuccess,
			Findings: []orchestration.Finding{
				{FindingID: "f-comms-1", Domain: orchestration.DomainCommunications, Type: "ONCALL_PING",
					TimelineTS: "2026-05-03T14:32:00Z", Confidence: 0.73},
			},
		},
	}
	payload := graph.FuseSynthesis(baseRC(), results, reg)
	breakdown, _ := payload["confidence_breakdown"].(map[string]any)
	boost, _ := breakdown["timeline_overlap_boost"].(float64)
	if boost != 0.05 {
		t.Fatalf("timeline_overlap_boost=%v want 0.05", boost)
	}
	score, _ := payload["confidence_score"].(float64)
	basePayload := graph.FuseSynthesis(baseRC(), map[orchestration.AgentDomain]orchestration.AgentResult{
		orchestration.DomainCommunications: results[orchestration.DomainCommunications],
	}, reg)
	baseScore, _ := basePayload["confidence_score"].(float64)
	if score <= baseScore {
		t.Fatalf("score with overlap %v should exceed without telemetry window %v", score, baseScore)
	}
}

func TestFuseSynthesis_AgentAgreementFourAgents(t *testing.T) {
	reg := registry.New([]string{"telemetry", "code", "context", "communications"})
	results := map[orchestration.AgentDomain]orchestration.AgentResult{
		orchestration.DomainTelemetry:      {Status: orchestration.AgentSuccess},
		orchestration.DomainCode:           {Status: orchestration.AgentSuccess},
		orchestration.DomainContext:        {Status: orchestration.AgentSuccess},
		orchestration.DomainCommunications: {Status: orchestration.AgentSuccess},
	}
	payload := graph.FuseSynthesis(baseRC(), results, reg)
	breakdown, _ := payload["confidence_breakdown"].(map[string]any)
	agreement, _ := breakdown["agent_agreement"].(float64)
	if agreement != 1.0 {
		t.Fatalf("agent_agreement=%v want 1.0", agreement)
	}
	_ = time.Now()
}
