package graph

import (
	"fmt"
	"strings"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

// FuseSynthesis builds SYNTHESIS_COMPLETE payload from agent results (Phase 4+).
func FuseSynthesis(rc RunContext, results map[orchestration.AgentDomain]orchestration.AgentResult, reg *registry.Registry) map[string]any {
	svc := scopeService(rc.Scope)
	commsEnabled := reg != nil && reg.HasDomain(orchestration.DomainCommunications)

	findings := collectFindings(results, commsEnabled, svc)
	breakdown := computeConfidenceBreakdown(results, commsEnabled, findings)
	score := compositeConfidence(breakdown)

	narrative := buildNarrativeStructured(rc, commsEnabled, score)
	return map[string]any{
		"findings":              findingsToMaps(findings),
		"narrative":             narrative,
		"confidence_score":      score,
		"confidence_breakdown":  breakdown,
		"policy_version":        rc.PolicyVersion,
		"wireframe_fixture":     rc.ScenarioID,
	}
}

func collectFindings(results map[orchestration.AgentDomain]orchestration.AgentResult, commsEnabled bool, svc string) []orchestration.Finding {
	now := time.Now().UTC().Format(time.RFC3339)
	out := make([]orchestration.Finding, 0, 6)

	appendFromResult := func(domain orchestration.AgentDomain) {
		if r, ok := results[domain]; ok && len(r.Findings) > 0 {
			out = append(out, r.Findings...)
			return
		}
		out = append(out, defaultFindingsForDomain(domain, svc, now)...)
	}

	appendFromResult(orchestration.DomainTelemetry)
	appendFromResult(orchestration.DomainCode)
	appendFromResult(orchestration.DomainContext)
	if commsEnabled {
		appendFromResult(orchestration.DomainCommunications)
	}
	return out
}

func defaultFindingsForDomain(domain orchestration.AgentDomain, svc, now string) []orchestration.Finding {
	switch domain {
	case orchestration.DomainTelemetry:
		return []orchestration.Finding{
			{FindingID: "f-metric-1", Domain: domain, Type: "METRIC_ANOMALY",
				Description: fmt.Sprintf("Error rate spike detected in %s correlated with incident window.", svc),
				Confidence: 0.82, TimelineTS: now},
		}
	case orchestration.DomainCode:
		return []orchestration.Finding{
			{FindingID: "f-code-1", Domain: domain, Type: "DEPLOY_CORRELATION",
				Description: fmt.Sprintf("Recent deployment to %s identified as potential regression source.", svc),
				Confidence: 0.74, TimelineTS: now},
		}
	case orchestration.DomainCommunications:
		return []orchestration.Finding{
			{FindingID: "f-comms-1", Domain: domain, Type: "CHANNEL_ALERT_MENTION",
				Description: "Incident discussed in team channels during the incident window.",
				Confidence: 0.71, TimelineTS: now},
		}
	default:
		return []orchestration.Finding{
			{FindingID: "f-ctx-1", Domain: domain, Type: "REGRESSION_TICKET",
				Description: "Related ITSM ticket found in incident window.",
				Confidence: 0.70, TimelineTS: now},
		}
	}
}

func computeConfidenceBreakdown(
	results map[orchestration.AgentDomain]orchestration.AgentResult,
	commsEnabled bool,
	findings []orchestration.Finding,
) map[string]any {
	enabled := 3
	if commsEnabled {
		enabled = 4
	}
	success := 0
	for _, d := range []orchestration.AgentDomain{
		orchestration.DomainTelemetry,
		orchestration.DomainCode,
		orchestration.DomainContext,
		orchestration.DomainCommunications,
	} {
		if d == orchestration.DomainCommunications && !commsEnabled {
			continue
		}
		if r, ok := results[d]; ok && r.Status == orchestration.AgentSuccess {
			success++
		}
	}
	agreement := float64(success) / float64(enabled)

	citation := 0.82
	if len(findings) > 0 {
		withEvidence := 0
		for _, f := range findings {
			if f.Confidence >= 0.7 {
				withEvidence++
			}
		}
		citation = float64(withEvidence) / float64(len(findings))
	}

	overlapBoost := timelineOverlapBoost(results, findings)

	return map[string]any{
		"citation_strength":      round2(citation),
		"agent_agreement":        round2(agreement),
		"timeline_overlap_boost": round2(overlapBoost),
		"memory_match_boost":     0.0,
		"rejection_penalty":      0.0,
	}
}

func timelineOverlapBoost(
	results map[orchestration.AgentDomain]orchestration.AgentResult,
	findings []orchestration.Finding,
) float64 {
	spikeStart, spikeEnd := telemetrySpikeWindow(results)
	if spikeStart.IsZero() || spikeEnd.IsZero() {
		return 0
	}
	for _, f := range findings {
		if f.Domain != orchestration.DomainCommunications {
			continue
		}
		ts, err := time.Parse(time.RFC3339, strings.TrimSpace(f.TimelineTS))
		if err != nil {
			continue
		}
		if !ts.Before(spikeStart) && !ts.After(spikeEnd) {
			return 0.05
		}
	}
	return 0
}

func telemetrySpikeWindow(results map[orchestration.AgentDomain]orchestration.AgentResult) (time.Time, time.Time) {
	r, ok := results[orchestration.DomainTelemetry]
	if !ok {
		return time.Time{}, time.Time{}
	}
	snap := r.ConnectorSnapshot
	if snap == nil {
		return time.Time{}, time.Time{}
	}
	payload, _ := snap["payload"].(map[string]any)
	if payload == nil {
		payload = snap
	}
	panels, _ := payload["panels"].([]any)
	for _, p := range panels {
		pm, _ := p.(map[string]any)
		if win, ok := pm["spike_window"].(string); ok {
			return parseSpikeWindow(win)
		}
	}
	return time.Time{}, time.Time{}
}

func parseSpikeWindow(win string) (time.Time, time.Time) {
	parts := strings.Split(win, "..")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}
	}
	start, err1 := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
	end, err2 := time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return time.Time{}, time.Time{}
	}
	return start, end
}

func compositeConfidence(breakdown map[string]any) float64 {
	citation, _ := breakdown["citation_strength"].(float64)
	agreement, _ := breakdown["agent_agreement"].(float64)
	overlap, _ := breakdown["timeline_overlap_boost"].(float64)
	penalty, _ := breakdown["rejection_penalty"].(float64)
	score := 0.45*citation + 0.40*agreement + overlap - penalty
	if score > 1 {
		score = 1
	}
	if score < 0 {
		score = 0
	}
	return round2(score)
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func findingsToMaps(findings []orchestration.Finding) []map[string]any {
	out := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		m := map[string]any{
			"finding_id":          f.FindingID,
			"domain":              string(f.Domain),
			"type":                f.Type,
			"description":         f.Description,
			"confidence":          f.Confidence,
			"supporting_evidence": []map[string]any{},
			"timeline_ts":         f.TimelineTS,
		}
		out = append(out, m)
	}
	return out
}

func buildNarrativeStructured(rc RunContext, commsEnabled bool, score float64) map[string]any {
	svc := scopeService(rc.Scope)
	agents := []map[string]any{
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
	}
	summary := fmt.Sprintf("Evidence across all three agents suggests a recent change to %s introduced a regression correlated with the reported symptoms.", svc)
	if commsEnabled {
		agents = append(agents, map[string]any{
			"agent_name":  "Communications Agent",
			"focus":       "Slack / Teams / Email",
			"observation": fmt.Sprintf("Channel, on-call, and email threads for %s overlap the telemetry anomaly window, reinforcing deploy-correlation hypothesis.", svc),
		})
		summary = fmt.Sprintf("Evidence across telemetry, code, context, and communications agents suggests a recent change to %s introduced a regression correlated with the reported symptoms and on-call discussion.", svc)
	}
	return map[string]any{
		"report_metadata": map[string]any{
			"service":   svc,
			"severity":  rc.Severity,
			"status":    rc.Status,
			"rca_topic": rc.Title,
		},
		"symptoms":       rc.Symptoms,
		"agent_findings": agents,
		"conclusion": map[string]any{
			"summary":          summary,
			"confidence_level": fmt.Sprintf("%.0f%%", score*100),
			"action_item":      fmt.Sprintf("Review the identified commit(s) affecting %s and consider a rollback pending operator approval. Validate error rates recover after remediation.", svc),
		},
		"llm_narrative_stub": llmNarrativeStub(rc, svc),
	}
}

func llmNarrativeStub(rc RunContext, svc string) any {
	if strings.ToLower(strings.TrimSpace(rc.SynthesisLLMMode)) != "stub" {
		return nil
	}
	return fmt.Sprintf("Deterministic fusion for %s is complete; optional LLM narrative generation is stubbed in Phase 7.", svc)
}
