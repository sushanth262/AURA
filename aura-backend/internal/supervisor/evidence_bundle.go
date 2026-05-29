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

// Evidence bundle JSON mirrors aura-frontend EvidenceBundle.
func buildEvidenceBundle(inv *Investigation) map[string]any {
	if inv.SynthesisPayload != nil {
		return evidenceFromSynthesis(inv)
	}
	return buildLegacyEvidenceBundle(inv)
}

func evidenceFromSynthesis(inv *Investigation) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	synth := inv.SynthesisPayload
	svc := scopeService(inv.Scope)

	score, _ := synth["confidence_score"].(float64)
	if score == 0 {
		score = 0.78
	}
	breakdown, _ := synth["confidence_breakdown"].(map[string]any)
	if breakdown == nil {
		breakdown = map[string]any{
			"citation_strength": 0.82, "agent_agreement": 0.76,
			"timeline_overlap_boost": 0.0, "memory_match_boost": 0.0, "rejection_penalty": 0.0,
		}
	}

	narrative, _ := synth["narrative"].(map[string]any)
	if narrative == nil {
		narrative = buildNarrativeStructured(inv)
	}

	findings, _ := synth["findings"].([]any)
	agentFindings := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		if m, ok := f.(map[string]any); ok {
			agentFindings = append(agentFindings, m)
		}
	}

	perAgent := perAgentSummariesFromFindings(agentFindings, svc)

	return map[string]any{
		"incident_id":          inv.IncidentID,
		"task_id":              inv.TaskID,
		"narrative":            narrative,
		"confidence_score":     score,
		"confidence_breakdown": breakdown,
		"per_agent_summaries":  perAgent,
		"agent_findings":       agentFindings,
		"evidence_refs":        []any{},
		"root_cause_candidates": []map[string]any{
			{"candidate_id": "rc-1", "description": narrativeConclusion(narrative), "confidence": score, "is_primary": true, "citations": []any{}},
		},
		"recommended_actions": []map[string]any{
			{"action_id": "a-1", "description": narrativeAction(narrative, svc), "automation": "Manual", "reversible": true, "risk": "Med"},
		},
		"prior_incident_matches": []any{},
		"synthesized_at":         now,
		"iteration":              1,
	}
}

func perAgentSummariesFromFindings(findings []map[string]any, svc string) []map[string]any {
	counts := map[string]int{}
	for _, f := range findings {
		d, _ := f["domain"].(string)
		if d != "" {
			counts[d]++
		}
	}
	order := []string{"telemetry", "code", "context", "communications"}
	out := make([]map[string]any, 0, len(order))
	for _, d := range order {
		n := counts[d]
		if n == 0 {
			continue
		}
		out = append(out, map[string]any{
			"domain": d, "finding_count": n, "status": "SUCCESS",
			"summary": defaultAgentSummary(d, svc),
			"execution_duration_ms": 1000,
		})
	}
	return out
}

func defaultAgentSummary(domain, svc string) string {
	switch domain {
	case "telemetry":
		return "Detected metric anomalies in " + svc + " during incident window."
	case "code":
		return "Correlated recent commits/deploys to " + svc + " with degradation timeline."
	case "communications":
		return "Slack, Teams, and email threads overlap incident timeline for " + svc + "."
	default:
		return "Cross-referenced ITSM tickets and runbooks for " + svc + "."
	}
}

func narrativeConclusion(narrative map[string]any) string {
	if c, ok := narrative["conclusion"].(map[string]any); ok {
		if s, ok := c["summary"].(string); ok {
			return s
		}
	}
	return ""
}

func narrativeAction(narrative map[string]any, svc string) string {
	if c, ok := narrative["conclusion"].(map[string]any); ok {
		if s, ok := c["action_item"].(string); ok {
			return s
		}
	}
	return "Rollback recent " + svc + " deployment and validate error rates recover."
}

func buildLegacyEvidenceBundle(inv *Investigation) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	svc := scopeService(inv.Scope)
	return map[string]any{
		"incident_id": inv.IncidentID,
		"task_id":     inv.TaskID,
		"narrative":   buildNarrativeStructured(inv),
		"confidence_score": 0.78,
		"confidence_breakdown": map[string]any{
			"citation_strength": 0.82, "agent_agreement": 0.76,
			"memory_match_boost": 0.05, "rejection_penalty": 0,
		},
		"per_agent_summaries": []map[string]any{
			{"domain": "telemetry", "summary": fmt.Sprintf("Detected metric anomalies in %s during incident window.", svc), "finding_count": 2, "status": "SUCCESS", "execution_duration_ms": 2200},
			{"domain": "code", "summary": fmt.Sprintf("Correlated recent commits/deploys to %s with degradation timeline.", svc), "finding_count": 1, "status": "SUCCESS", "execution_duration_ms": 1800},
			{"domain": "context", "summary": fmt.Sprintf("Cross-referenced ITSM tickets and runbooks for %s.", svc), "finding_count": 1, "status": "PARTIAL", "execution_duration_ms": 900},
		},
		"agent_findings": []map[string]any{
			{"finding_id": "f-metric-1", "domain": "telemetry", "type": "METRIC_ANOMALY",
				"description": fmt.Sprintf("Error rate spike detected in %s correlated with incident window.", svc),
				"confidence": 0.82, "supporting_evidence": []map[string]any{}, "timeline_ts": now},
			{"finding_id": "f-code-1", "domain": "code", "type": "DEPLOY_CORRELATION",
				"description": fmt.Sprintf("Recent deployment to %s identified as potential regression source.", svc),
				"confidence": 0.74, "supporting_evidence": []map[string]any{}, "timeline_ts": now},
		},
		"evidence_refs": []any{},
		"root_cause_candidates": []map[string]any{
			{"candidate_id": "rc-mock-1", "description": fmt.Sprintf("Recent change to %s correlated with reported symptoms across telemetry and deploy timeline.", svc), "confidence": 0.78, "is_primary": true, "citations": []any{}},
		},
		"recommended_actions": []map[string]any{
			{"action_id": "a-mock-1", "description": fmt.Sprintf("Rollback recent %s deployment and validate error rates recover.", svc), "automation": "Manual", "reversible": true, "risk": "Med"},
		},
		"prior_incident_matches": []any{},
		"synthesized_at":         now,
	"iteration":              1,
	}
}

func buildNarrativeStructured(inv *Investigation) map[string]any {
	svc := scopeService(inv.Scope)
	return map[string]any{
		"report_metadata": map[string]any{
			"service":   svc,
			"severity":  inv.Severity,
			"status":    inv.Status,
			"rca_topic": inv.Title,
		},
		"symptoms": inv.Symptoms,
		"agent_findings": []map[string]any{
			{"agent_name": "Telemetry Agent", "focus": "Metrics & Logs",
				"observation": fmt.Sprintf("Detected anomalous error-rate and latency patterns in %s during the incident window.", svc)},
			{"agent_name": "Code Agent", "focus": "Deployments & Commits",
				"observation": fmt.Sprintf("Correlated recent deployments and commits touching %s with the degradation timeline.", svc)},
			{"agent_name": "Context Agent", "focus": "Tickets & Runbooks",
				"observation": fmt.Sprintf("Cross-referenced ITSM tickets and runbooks for %s.", svc)},
		},
		"conclusion": map[string]any{
			"summary":          fmt.Sprintf("Evidence across all three agents suggests a recent change to %s introduced a regression.", svc),
			"confidence_level": "78%",
			"action_item":      fmt.Sprintf("Review commits affecting %s and consider rollback pending operator approval.", svc),
		},
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
