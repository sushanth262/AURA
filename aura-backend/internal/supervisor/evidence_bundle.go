package supervisor

import (
	"time"
)

// Evidence bundle JSON mirrors aura-frontend EvidenceBundle (demo stub after synthesis).
func buildEvidenceBundle(inv *Investigation) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return map[string]any{
		"incident_id": inv.IncidentID,
		"task_id":     inv.TaskID,
		"narrative":   "Mock synthesis aligned with Screen 2 evidence progression.",
		"confidence_score": 0.78,
		"confidence_breakdown": map[string]any{
			"citation_strength":    0.82,
			"agent_agreement":      0.76,
			"memory_match_boost":   0.05,
			"rejection_penalty":    0,
		},
		"per_agent_summaries": []map[string]any{
			{"domain": "telemetry", "summary": "Grafana-style mock snapshots correlated rollout window.", "finding_count": 2, "status": "SUCCESS", "execution_duration_ms": 2200},
			{"domain": "code", "summary": "GitHub-style mock revision touches checkout handlers.", "finding_count": 1, "status": "SUCCESS", "execution_duration_ms": 1800},
			{"domain": "context", "summary": "Jira-style mock references related infra ticket.", "finding_count": 1, "status": "PARTIAL", "execution_duration_ms": 900},
		},
		"agent_findings": []map[string]any{
			{
				"finding_id":          "f-metric-1",
				"domain":              "telemetry",
				"type":                "METRIC_ANOMALY",
				"description":         "5xx ratio spike correlated with payments-api rollout window (Grafana-style mock).",
				"confidence":          0.82,
				"supporting_evidence": []map[string]any{},
				"timeline_ts":         now,
			},
			{
				"finding_id":          "f-code-1",
				"domain":              "code",
				"type":                "DEPLOY_CORRELATION",
				"description":         "Revision abc123def456 touches checkout handlers and gateway upstream config (GitHub mock).",
				"confidence":          0.74,
				"supporting_evidence": []map[string]any{},
				"timeline_ts":         now,
			},
		},
		"evidence_refs": []any{},
		"root_cause_candidates": []map[string]any{
			{"candidate_id": "rc-mock-1", "description": "Rollout regression correlated across telemetry and deploy timeline.", "confidence": 0.78, "is_primary": true, "citations": []any{}},
		},
		"recommended_actions": []map[string]any{
			{"action_id": "a-mock-1", "description": "Rollback recent payments-api deploy and validate error rates.", "automation": "Manual", "reversible": true, "risk": "Med", "estimated_duration_seconds": nil, "runbook_ref": nil},
		},
		"prior_incident_matches": []any{},
		"synthesized_at":         now,
		"iteration":              1,
	}
}

func evidenceStillInProgress(status string) bool {
	switch status {
	case "QUEUED", "INTAKE", "PLANNING", "RETRIEVING", "SYNTHESIS", "REPLANNING":
		return true
	default:
		return false
	}
}
