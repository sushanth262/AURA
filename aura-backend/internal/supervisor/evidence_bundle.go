package supervisor

import (
	"fmt"
	"time"
)

func scopeService(scope map[string]any) string {
	if s, ok := scope["service"].(string); ok && s != "" {
		return s
	}
	return "unknown-service"
}

func buildNarrativeStructured(inv *Investigation) map[string]any {
	svc := scopeService(inv.Scope)
	return map[string]any{
		"report_metadata": map[string]any{
			"service":    svc,
			"severity":   inv.Severity,
			"status":     inv.Status,
			"rca_topic":  inv.Title,
		},
		"symptoms": inv.Symptoms,
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

// Evidence bundle JSON mirrors aura-frontend EvidenceBundle (demo stub after synthesis).
func buildEvidenceBundle(inv *Investigation) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	svc := scopeService(inv.Scope)
	return map[string]any{
		"incident_id": inv.IncidentID,
		"task_id":     inv.TaskID,
		"narrative":   buildNarrativeStructured(inv),
		"confidence_score": 0.78,
		"confidence_breakdown": map[string]any{
			"citation_strength":    0.82,
			"agent_agreement":      0.76,
			"memory_match_boost":   0.05,
			"rejection_penalty":    0,
		},
		"per_agent_summaries": []map[string]any{
			{"domain": "telemetry", "summary": fmt.Sprintf("Detected metric anomalies in %s during incident window.", svc), "finding_count": 2, "status": "SUCCESS", "execution_duration_ms": 2200},
			{"domain": "code", "summary": fmt.Sprintf("Correlated recent commits/deploys to %s with degradation timeline.", svc), "finding_count": 1, "status": "SUCCESS", "execution_duration_ms": 1800},
			{"domain": "context", "summary": fmt.Sprintf("Cross-referenced ITSM tickets and runbooks for %s.", svc), "finding_count": 1, "status": "PARTIAL", "execution_duration_ms": 900},
		},
		"agent_findings": []map[string]any{
			{
				"finding_id":          "f-metric-1",
				"domain":              "telemetry",
				"type":                "METRIC_ANOMALY",
				"description":         fmt.Sprintf("Error rate spike detected in %s correlated with incident window.", svc),
				"confidence":          0.82,
				"supporting_evidence": []map[string]any{},
				"timeline_ts":         now,
			},
			{
				"finding_id":          "f-code-1",
				"domain":              "code",
				"type":                "DEPLOY_CORRELATION",
				"description":         fmt.Sprintf("Recent deployment to %s identified as potential regression source.", svc),
				"confidence":          0.74,
				"supporting_evidence": []map[string]any{},
				"timeline_ts":         now,
			},
		},
		"evidence_refs": []any{},
		"root_cause_candidates": []map[string]any{
			{"candidate_id": "rc-mock-1", "description": fmt.Sprintf("Recent change to %s correlated with reported symptoms across telemetry and deploy timeline.", svc), "confidence": 0.78, "is_primary": true, "citations": []any{}},
		},
		"recommended_actions": []map[string]any{
			{"action_id": "a-mock-1", "description": fmt.Sprintf("Rollback recent %s deployment and validate error rates recover.", svc), "automation": "Manual", "reversible": true, "risk": "Med", "estimated_duration_seconds": nil, "runbook_ref": nil},
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
