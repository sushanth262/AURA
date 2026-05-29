package graph

import "github.com/sushanth262/AURA/aura-backend/internal/orchestration"

func isResumable(cp orchestration.GraphCheckpoint) bool {
	switch cp.CurrentStatus {
	case orchestration.StatusHITLPending,
		orchestration.StatusFailed,
		orchestration.StatusQueued:
		return false
	default:
		return len(cp.CompletedNodes) > 0
	}
}

func nodeCompleted(cp orchestration.GraphCheckpoint, nodeID string) bool {
	for _, id := range cp.CompletedNodes {
		if id == nodeID {
			return true
		}
	}
	return false
}

func countSuccessfulAgents(results map[orchestration.AgentDomain]orchestration.AgentResult) int {
	n := 0
	for _, r := range results {
		if r.Status == orchestration.AgentSuccess {
			n++
		}
	}
	return n
}

func restoreAgentResults(cp orchestration.GraphCheckpoint) map[orchestration.AgentDomain]orchestration.AgentResult {
	out := make(map[orchestration.AgentDomain]orchestration.AgentResult, len(cp.AgentResults))
	for k, v := range cp.AgentResults {
		out[orchestration.AgentDomain(k)] = v
	}
	return out
}
