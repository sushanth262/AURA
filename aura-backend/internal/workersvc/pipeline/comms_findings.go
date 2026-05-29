package pipeline

import (
	"strings"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

func buildCommunicationsFindings(snaps map[string]map[string]any) []orchestration.Finding {
	now := time.Now().UTC().Format(time.RFC3339)
	findings := make([]orchestration.Finding, 0, 3)

	if snap, ok := snaps["slack"]; ok && len(snap) > 0 {
		ts := commsMessageTS(snap, "2026-05-03T14:31:00Z")
		findings = append(findings, orchestration.Finding{
			FindingID:   "f-comms-slack",
			Domain:      orchestration.DomainCommunications,
			Type:        "CHANNEL_ALERT_MENTION",
			Description: "Incident discussed in #incidents during the outage window.",
			Confidence:  0.71,
			TimelineTS:  ts,
		})
	}
	if snap, ok := snaps["teams"]; ok && len(snap) > 0 {
		ts := commsMessageTS(snap, "2026-05-03T14:32:00Z")
		findings = append(findings, orchestration.Finding{
			FindingID:   "f-comms-teams",
			Domain:      orchestration.DomainCommunications,
			Type:        "ONCALL_PING",
			Description: "Pager/on-call mention correlates with telemetry anomaly onset.",
			Confidence:  0.73,
			TimelineTS:  ts,
		})
	}
	if snap, ok := snaps["email"]; ok && len(snap) > 0 {
		ts := commsMessageTS(snap, "2026-05-03T14:30:00Z")
		findings = append(findings, orchestration.Finding{
			FindingID:   "f-comms-email",
			Domain:      orchestration.DomainCommunications,
			Type:        "EMAIL_THREAD",
			Description: "Related outage email thread found in incident window.",
			Confidence:  0.69,
			TimelineTS:  ts,
		})
	}
	if len(findings) == 0 {
		findings = append(findings, orchestration.Finding{
			FindingID:   "f-comms-1",
			Domain:      orchestration.DomainCommunications,
			Type:        "CHANNEL_ALERT_MENTION",
			Description: "Incident discussed in team channels during the incident window.",
			Confidence:  0.71,
			TimelineTS:  now,
		})
	}
	return findings
}

func commsMessageTS(snap map[string]any, fallback string) string {
	payload, _ := snap["payload"].(map[string]any)
	if payload == nil {
		payload = snap
	}
	if ts := firstMessageTS(payload); ts != "" {
		return ts
	}
	return fallback
}

func firstMessageTS(payload map[string]any) string {
	if chans, ok := payload["channels"].([]any); ok {
		for _, ch := range chans {
			cm, _ := ch.(map[string]any)
			if ts := messageTSFromContainer(cm); ts != "" {
				return ts
			}
		}
	}
	if threads, ok := payload["threads"].([]any); ok {
		for _, th := range threads {
			tm, _ := th.(map[string]any)
			if ts := messageTSFromContainer(tm); ts != "" {
				return ts
			}
		}
	}
	return ""
}

func messageTSFromContainer(m map[string]any) string {
	if m == nil {
		return ""
	}
	if msgs, ok := m["messages"].([]any); ok && len(msgs) > 0 {
		if msg, ok := msgs[0].(map[string]any); ok {
			if ts, ok := msg["ts"].(string); ok {
				return strings.TrimSpace(ts)
			}
		}
	}
	return ""
}
