package supervisor

import (
	"context"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

// RunMockScenario emits WS timeline aligned with WIREFRAMES.md Screen 2 ordering (stub agents).
// wf optionally enriches AGENT_COMPLETE steps with aura-worker fixture snapshots (grafana, github, jira).
func RunMockScenario(cfg config.Config, st *Store, hub *WSClients, inv *Investigation, wf SnapshotFetcher) {
	if wf == nil {
		wf = noopFetcher{}
	}
	go func() {
		taskID := inv.TaskID
		incidentID := inv.IncidentID
		seqN := 0
		seq := func() int {
			seqN++
			return seqN
		}
		emit := func(delay time.Duration, eventType string, domain string, payload map[string]any) {
			time.Sleep(delay)
			n := seq()
			ev := map[string]any{
				"task_id":      taskID,
				"incident_id":  incidentID,
				"event_type":   eventType,
				"payload":      payload,
				"timestamp":    time.Now().UTC().Format(time.RFC3339Nano),
				"sequence_num": n,
			}
			if domain != "" {
				ev["agent_domain"] = domain
			}
			hub.Broadcast(taskID, ev)
		}

		labels := inv.Fixture.TimelineLabels
		idx := func(i int) string {
			if i >= 0 && i < len(labels) {
				return labels[i]
			}
			return "timeline_step"
		}

		st.UpdateStatus(taskID, "INTAKE")
		emit(80*time.Millisecond, "TASK_CLAIMED", "", map[string]any{"message": idx(0)})
		emit(120*time.Millisecond, "AGENT_STARTED", "", map[string]any{"message": idx(1)})
		emit(100*time.Millisecond, "AGENT_COMPLETE", "", map[string]any{"message": idx(2)})

		st.UpdateStatus(taskID, "RETRIEVING")
		emit(150*time.Millisecond, "AGENT_STARTED", "telemetry", map[string]any{"progress_pct": 52, "message": idx(3)})
		emit(180*time.Millisecond, "AGENT_STARTED", "code", map[string]any{"progress_pct": 21, "message": idx(4)})
		emit(160*time.Millisecond, "AGENT_STARTED", "context", map[string]any{"progress_pct": 68, "message": idx(5)})

		ctx := context.Background()
		key := inv.FixtureKey
		telPayload := map[string]any{"finding_count": 2}
		if snap, err := wf.Fetch(ctx, "grafana", key); err == nil && len(snap) > 0 {
			telPayload["connector_snapshot"] = snap
		}
		codePayload := map[string]any{"finding_count": 1}
		if snap, err := wf.Fetch(ctx, "github", key); err == nil && len(snap) > 0 {
			codePayload["connector_snapshot"] = snap
		}
		ctxPayload := map[string]any{"finding_count": 1}
		if snap, err := wf.Fetch(ctx, "jira", key); err == nil && len(snap) > 0 {
			ctxPayload["connector_snapshot"] = snap
		}

		emit(400*time.Millisecond, "AGENT_COMPLETE", "telemetry", telPayload)
		emit(120*time.Millisecond, "AGENT_COMPLETE", "code", codePayload)
		emit(100*time.Millisecond, "AGENT_COMPLETE", "context", ctxPayload)

		st.UpdateStatus(taskID, "SYNTHESIS")
		findings := []map[string]any{
			{
				"finding_id":          "f-metric-1",
				"domain":              "telemetry",
				"type":                "METRIC_ANOMALY",
				"description":         "5xx ratio spike correlated with payments-api rollout window (Grafana-style mock).",
				"confidence":          0.82,
				"supporting_evidence": []map[string]any{},
				"timeline_ts":         time.Now().UTC().Format(time.RFC3339),
			},
			{
				"finding_id":          "f-code-1",
				"domain":              "code",
				"type":                "DEPLOY_CORRELATION",
				"description":         "Revision abc123def456 touches checkout handlers and gateway upstream config (GitHub mock).",
				"confidence":          0.74,
				"supporting_evidence": []map[string]any{},
				"timeline_ts":         time.Now().UTC().Format(time.RFC3339),
			},
		}
		emit(200*time.Millisecond, "SYNTHESIS_COMPLETE", "", map[string]any{
			"findings":          findings,
			"narrative":         "Mock synthesis aligned with Screen 2 evidence progression.",
			"confidence_score":  0.78,
			"policy_version":    cfg.PolicyVersion,
			"wireframe_fixture": inv.Fixture.ScenarioID,
		})

		st.UpdateStatus(taskID, "HITL_PENDING")
	}()
}
