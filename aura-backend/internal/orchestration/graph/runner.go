package graph

import (
	"context"
	"errors"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// Run executes the graph synchronously (caller typically runs in a goroutine).
func (r *Runner) Run(ctx context.Context, g InvestigationGraph, rc RunContext) error {
	if r.Sleep == nil {
		r.Sleep = time.Sleep
	}
	if r.Now == nil {
		r.Now = func() time.Time { return time.Now().UTC() }
	}
	if r.Checkpoints == nil {
		r.Checkpoints = NewMemoryCheckpointStore()
	}
	if r.FetchSnapshot == nil {
		r.FetchSnapshot = noopSnapshot{}
	}

	cp, hasCP := r.Checkpoints.Load(rc.TaskID)
	resume := hasCP && isResumable(cp)
	results := restoreAgentResults(cp)

	seq := 0
	nextSeq := func() int {
		seq++
		return seq
	}
	emit := func(delay time.Duration, eventType string, domain orchestration.AgentDomain, payload map[string]any) {
		if delay > 0 {
			r.Sleep(delay)
		}
		ev := orchestration.ProgressEvent{
			TaskID:      rc.TaskID,
			IncidentID:  rc.IncidentID,
			EventType:   eventType,
			AgentDomain: domain,
			Payload:     payload,
			Timestamp:   r.Now().Format(time.RFC3339Nano),
			SequenceNum: nextSeq(),
		}
		if r.Emit != nil {
			r.Emit(ev)
		}
	}
	setStatus := func(st orchestration.InvestigationStatus) {
		rc.Status = string(st)
		if r.UpdateStatus != nil {
			r.UpdateStatus(rc.TaskID, st)
		}
	}
	saveCP := func(st orchestration.InvestigationStatus, completed ...string) {
		loaded, _ := r.Checkpoints.Load(rc.TaskID)
		loaded.TaskID = rc.TaskID
		loaded.IncidentID = rc.IncidentID
		loaded.CurrentStatus = st
		loaded.GraphVersion = g.Version
		loaded.AgentResults = agentResultsToMap(results)
		for _, id := range completed {
			loaded.CompletedNodes = appendUnique(loaded.CompletedNodes, id)
		}
		r.Checkpoints.Save(loaded)
		cp = loaded
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if resume && nodeCompleted(cp, "synthesis_complete") {
		setStatus(orchestration.StatusHITLPending)
		return nil
	}

	if !resume {
		r.Checkpoints.Save(orchestration.GraphCheckpoint{
			TaskID:        rc.TaskID,
			IncidentID:    rc.IncidentID,
			CurrentStatus: orchestration.StatusQueued,
			GraphVersion:  g.Version,
		})
	}

	if !resume || !nodeCompleted(cp, "supervisor_claim") {
		setStatus(orchestration.StatusIntake)
		emit(supervisorDelay("claim"), "TASK_CLAIMED", orchestration.DomainSupervisor,
			map[string]any{"message": rc.timelineMessage(0)})
		saveCP(orchestration.StatusIntake, "supervisor_claim")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if !resume || !nodeCompleted(cp, "supervisor_plan_start") {
		emit(supervisorDelay("plan_start"), "AGENT_STARTED", orchestration.DomainSupervisor,
			map[string]any{"message": rc.timelineMessage(1)})
		saveCP(orchestration.StatusPlanning, "supervisor_plan_start")
	}

	if !resume || !nodeCompleted(cp, "supervisor_plan_done") {
		emit(supervisorDelay("plan_done"), "AGENT_COMPLETE", orchestration.DomainSupervisor,
			map[string]any{"message": rc.timelineMessage(2)})
		setStatus(orchestration.StatusPlanning)
		saveCP(orchestration.StatusPlanning, "supervisor_plan_done")
	}

	if !resume || !nodeCompleted(cp, "graph_planned") {
		emit(0, "GRAPH_PLANNED", orchestration.DomainSupervisor, manifestPayload(g, r.Registry))
		saveCP(orchestration.StatusPlanning, "graph_planned")
	}

	setStatus(orchestration.StatusRetrieving)

	var startNodes, doneNodes []Node
	for _, n := range g.Nodes {
		switch n.Kind {
		case NodeKindAgentStart:
			startNodes = append(startNodes, n)
		case NodeKindAgentDone:
			doneNodes = append(doneNodes, n)
		}
	}
	enabledCount := len(startNodes)

	for _, n := range startNodes {
		if err := ctx.Err(); err != nil {
			return err
		}
		if resume && nodeCompleted(cp, n.ID) {
			continue
		}
		msg := rc.timelineMessage(n.TimelineIndex)
		emit(agentStartDelay(n.AgentDomain), "AGENT_STARTED", n.AgentDomain, map[string]any{
			"progress_pct": DefaultStartProgressPct(n.AgentDomain),
			"message":      msg,
		})
		r.Checkpoints.MarkNodeComplete(rc.TaskID, n.ID, orchestration.StatusRetrieving)
	}

	successCount := countSuccessfulAgents(results)

	for _, n := range doneNodes {
		if err := ctx.Err(); err != nil {
			return err
		}
		if resume && nodeCompleted(cp, n.ID) {
			if res, ok := results[n.AgentDomain]; ok && res.Status == orchestration.AgentSuccess {
				emit(0, "AGENT_SKIPPED", n.AgentDomain, map[string]any{
					"reason":        "checkpoint_resume",
					"finding_count": res.FindingCount,
				})
				continue
			}
		}

		var result orchestration.AgentResult
		var err error
		if r.AgentRunner != nil {
			result, err = r.AgentRunner(ctx, n, rc)
		} else {
			result, err = r.stubAgentResult(ctx, n, rc)
		}
		if err != nil || result.Status == orchestration.AgentFailed {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			r.Checkpoints.MarkNodeFailed(rc.TaskID, n.ID)
			continue
		}
		successCount++
		results[n.AgentDomain] = result
		r.Checkpoints.SaveAgentResult(rc.TaskID, n.AgentDomain, result)

		payload := map[string]any{"finding_count": result.FindingCount}
		if len(result.ConnectorSnapshot) > 0 {
			payload["connector_snapshot"] = result.ConnectorSnapshot
		}
		emit(agentCompleteDelay(n.AgentDomain), "AGENT_COMPLETE", n.AgentDomain, payload)
		r.Checkpoints.MarkNodeComplete(rc.TaskID, n.ID, orchestration.StatusRetrieving)
	}

	if !r.Policies.CanSynthesize(successCount, enabledCount) {
		setStatus(orchestration.StatusFailed)
		return nil
	}

	if successCount < enabledCount {
		setStatus(orchestration.StatusPartialEvidence)
	} else {
		setStatus(orchestration.StatusSynthesis)
	}

	if resume && nodeCompleted(cp, "synthesis") {
		setStatus(orchestration.StatusHITLPending)
		saveCP(orchestration.StatusHITLPending, "synthesis_complete")
		return nil
	}

	emit(200*time.Millisecond, "SYNTHESIS_COMPLETE", "", FuseSynthesis(rc, results, r.Registry))
	saveCP(orchestration.StatusSynthesis, "synthesis")

	setStatus(orchestration.StatusHITLPending)
	saveCP(orchestration.StatusHITLPending, "synthesis_complete")
	return nil
}

func agentResultsToMap(results map[orchestration.AgentDomain]orchestration.AgentResult) map[string]orchestration.AgentResult {
	if len(results) == 0 {
		return nil
	}
	out := make(map[string]orchestration.AgentResult, len(results))
	for d, r := range results {
		out[string(d)] = r
	}
	return out
}
