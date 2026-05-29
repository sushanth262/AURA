package supervisor

import (
	"context"
	"strings"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/graph"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

// StartInvestigation runs the mock investigation timeline (graph engine or legacy).
func StartInvestigation(cfg config.Config, st *Store, hub *WSClients, inv *Investigation, wf SnapshotFetcher) {
	if strings.ToLower(strings.TrimSpace(cfg.GraphEngineMode)) == "legacy" {
		RunMockScenario(cfg, st, hub, inv, wf)
		return
	}
	runGraphEngine(cfg, st, hub, inv, wf)
}

func runGraphEngine(cfg config.Config, st *Store, hub *WSClients, inv *Investigation, wf SnapshotFetcher) {
	if wf == nil {
		wf = noopFetcher{}
	}
	enabled := cfg.EnabledAgents
	if len(enabled) == 0 {
		enabled = []string{"telemetry", "code", "context"}
	}
	if err := registry.ValidateEnabledDomains(enabled); err != nil {
		// Skip unknown domains; supervisor still runs with valid subset.
		filtered := make([]string, 0, len(enabled))
		known := map[string]struct{}{}
		for _, d := range registry.KnownDomains() {
			known[d] = struct{}{}
		}
		for _, d := range enabled {
			d = strings.ToLower(strings.TrimSpace(d))
			if _, ok := known[d]; ok {
				filtered = append(filtered, d)
			}
		}
		if len(filtered) == 0 {
			filtered = []string{"telemetry", "code", "context"}
		}
		enabled = filtered
	}
	reg := registry.New(enabled)
	minAgents, join := cfg.OrchestrationPolicies()
	policies := orchestration.PoliciesFromConfig(minAgents, join)

	rc := graph.RunContext{
		TaskID:         inv.TaskID,
		IncidentID:     inv.IncidentID,
		Title:          inv.Title,
		Severity:       inv.Severity,
		Symptoms:       inv.Symptoms,
		Scope:          inv.Scope,
		FixtureKey:     inv.FixtureKey,
		TimelineLabels: inv.Fixture.TimelineLabels,
		ScenarioID:     inv.Fixture.ScenarioID,
		PolicyVersion:  cfg.PolicyVersion,
		Status:         inv.Status,
	}

	runner := &graph.Runner{
		Policies:      policies,
		Registry:      reg,
		Checkpoints:   graph.NewMemoryCheckpointStore(),
		Sleep:         nil,
		FetchSnapshot: wf,
		Emit: func(ev orchestration.ProgressEvent) {
			hub.Broadcast(ev.TaskID, progressEventToMap(ev))
		},
		UpdateStatus: func(taskID string, status orchestration.InvestigationStatus) bool {
			return st.UpdateStatus(taskID, string(status))
		},
	}

	go func() {
		_ = graph.PlanAndRun(context.Background(), runner, reg, policies, rc)
	}()
}

func progressEventToMap(ev orchestration.ProgressEvent) map[string]any {
	m := map[string]any{
		"task_id":      ev.TaskID,
		"incident_id":  ev.IncidentID,
		"event_type":   ev.EventType,
		"payload":      ev.Payload,
		"timestamp":    ev.Timestamp,
		"sequence_num": ev.SequenceNum,
	}
	if ev.AgentDomain != "" {
		m["agent_domain"] = string(ev.AgentDomain)
	}
	return m
}
