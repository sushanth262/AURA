package graph

import (
	"context"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

// Runner executes an InvestigationGraph and emits progress events.
type Runner struct {
	Policies   orchestration.Policies
	Registry   *registry.Registry
	Checkpoints CheckpointStore
	Sleep      func(time.Duration)
	Now        func() time.Time

	Emit         func(orchestration.ProgressEvent)
	UpdateStatus func(taskID string, status orchestration.InvestigationStatus) bool
	FetchSnapshot orchestration.SnapshotFetcher

	// AgentWorker dispatches tasks to aura-worker when AGENT_EXECUTION_MODE=worker.
	AgentWorker orchestration.AgentWorkerClient

	// AgentRunner optional hook for tests; nil uses stub snapshot path.
	AgentRunner func(ctx context.Context, node Node, rc RunContext) (orchestration.AgentResult, error)
}

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
		cp, _ := r.Checkpoints.Load(rc.TaskID)
		cp.TaskID = rc.TaskID
		cp.IncidentID = rc.IncidentID
		cp.CurrentStatus = st
		cp.GraphVersion = g.Version
		for _, id := range completed {
			cp.CompletedNodes = appendUnique(cp.CompletedNodes, id)
		}
		r.Checkpoints.Save(cp)
	}

	r.Checkpoints.Save(orchestration.GraphCheckpoint{
		TaskID:        rc.TaskID,
		IncidentID:    rc.IncidentID,
		CurrentStatus: orchestration.StatusQueued,
		GraphVersion:  g.Version,
	})

	setStatus(orchestration.StatusIntake)
	emit(supervisorDelay("claim"), "TASK_CLAIMED", orchestration.DomainSupervisor,
		map[string]any{"message": rc.timelineMessage(0)})
	saveCP(orchestration.StatusIntake, "supervisor_claim")

	emit(supervisorDelay("plan_start"), "AGENT_STARTED", orchestration.DomainSupervisor,
		map[string]any{"message": rc.timelineMessage(1)})
	saveCP(orchestration.StatusPlanning, "supervisor_plan_start")

	emit(supervisorDelay("plan_done"), "AGENT_COMPLETE", orchestration.DomainSupervisor,
		map[string]any{"message": rc.timelineMessage(2)})
	setStatus(orchestration.StatusPlanning)
	saveCP(orchestration.StatusPlanning, "supervisor_plan_done")

	emit(0, "GRAPH_PLANNED", orchestration.DomainSupervisor, manifestPayload(g, r.Registry))
	saveCP(orchestration.StatusPlanning, "graph_planned")

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

	successCount := 0
	enabledCount := len(startNodes)

	for _, n := range startNodes {
		msg := rc.timelineMessage(n.TimelineIndex)
		emit(agentStartDelay(n.AgentDomain), "AGENT_STARTED", n.AgentDomain, map[string]any{
			"progress_pct": DefaultStartProgressPct(n.AgentDomain),
			"message":      msg,
		})
		r.Checkpoints.MarkNodeComplete(rc.TaskID, n.ID, orchestration.StatusRetrieving)
	}

	results := make(map[orchestration.AgentDomain]orchestration.AgentResult, len(doneNodes))
	for _, n := range doneNodes {
		var result orchestration.AgentResult
		var err error
		if r.AgentRunner != nil {
			result, err = r.AgentRunner(ctx, n, rc)
		} else {
			result, err = r.stubAgentResult(ctx, n, rc)
		}
		if err != nil || result.Status == orchestration.AgentFailed {
			r.Checkpoints.MarkNodeFailed(rc.TaskID, n.ID)
			continue
		}
		successCount++
		results[n.AgentDomain] = result

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

	emit(200*time.Millisecond, "SYNTHESIS_COMPLETE", "", BuildSynthesisPayload(rc))
	saveCP(orchestration.StatusSynthesis, "synthesis")

	setStatus(orchestration.StatusHITLPending)
	saveCP(orchestration.StatusHITLPending, "synthesis_complete")
	return nil
}

func (r *Runner) stubAgentResult(ctx context.Context, n Node, rc RunContext) (orchestration.AgentResult, error) {
	if r.AgentRunner != nil {
		return r.AgentRunner(ctx, n, rc)
	}
	if r.AgentWorker != nil {
		task := agentTaskFromNode(n, rc)
		result, err := r.AgentWorker.Execute(ctx, n.AgentDomain, task)
		if err != nil {
			return orchestration.AgentResult{Domain: n.AgentDomain, TaskID: rc.TaskID, Status: orchestration.AgentFailed}, err
		}
		if result.FindingCount == 0 && len(result.Findings) > 0 {
			result.FindingCount = len(result.Findings)
		}
		if result.FindingCount == 0 {
			result.FindingCount = DefaultFindingCount(n.AgentDomain)
		}
		return result, nil
	}
	payload := map[string]any{}
	if n.Connector != "" && rc.FixtureKey != "" {
		if snap, err := r.FetchSnapshot.Fetch(ctx, n.Connector, rc.FixtureKey); err == nil && len(snap) > 0 {
			payload = snap
		}
	}
	return orchestration.AgentResult{
		Domain:            n.AgentDomain,
		TaskID:            rc.TaskID,
		Status:            orchestration.AgentSuccess,
		FindingCount:      DefaultFindingCount(n.AgentDomain),
		ConnectorSnapshot: payload,
		CompletedAt:       r.Now(),
	}, nil
}

func agentTaskFromNode(n Node, rc RunContext) orchestration.AgentTask {
	connectors := []string{}
	if n.Connector != "" {
		connectors = []string{n.Connector}
	}
	return orchestration.AgentTask{
		IncidentID: rc.IncidentID,
		TaskID:     rc.TaskID,
		Domain:     n.AgentDomain,
		FixtureKey: rc.FixtureKey,
		Connectors: connectors,
	}
}

type noopSnapshot struct{}

func (noopSnapshot) Fetch(context.Context, string, string) (map[string]any, error) {
	return nil, nil
}

func domainTimelineIndex(d orchestration.AgentDomain) int {
	switch d {
	case orchestration.DomainTelemetry:
		return 3
	case orchestration.DomainCode:
		return 4
	case orchestration.DomainContext:
		return 5
	default:
		return 0
	}
}

func supervisorDelay(step string) time.Duration {
	switch step {
	case "claim":
		return 80 * time.Millisecond
	case "plan_start":
		return 120 * time.Millisecond
	case "plan_done":
		return 100 * time.Millisecond
	default:
		return 0
	}
}

func agentStartDelay(d orchestration.AgentDomain) time.Duration {
	switch d {
	case orchestration.DomainTelemetry:
		return 150 * time.Millisecond
	case orchestration.DomainCode:
		return 180 * time.Millisecond
	case orchestration.DomainContext:
		return 160 * time.Millisecond
	case orchestration.DomainCommunications:
		return 140 * time.Millisecond
	default:
		return 100 * time.Millisecond
	}
}

func agentCompleteDelay(d orchestration.AgentDomain) time.Duration {
	switch d {
	case orchestration.DomainTelemetry:
		return 400 * time.Millisecond
	case orchestration.DomainCode:
		return 120 * time.Millisecond
	case orchestration.DomainContext:
		return 100 * time.Millisecond
	case orchestration.DomainCommunications:
		return 90 * time.Millisecond
	default:
		return 100 * time.Millisecond
	}
}

// PlanAndRun builds the graph from registry policies and executes it.
func PlanAndRun(ctx context.Context, r *Runner, reg *registry.Registry, policies orchestration.Policies, rc RunContext) error {
	g, err := Plan(reg, policies)
	if err != nil {
		return err
	}
	return r.Run(ctx, g, rc)
}
