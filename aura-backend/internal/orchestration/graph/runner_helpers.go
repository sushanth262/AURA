package graph

import (
	"context"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

// Runner executes an InvestigationGraph and emits progress events.
type Runner struct {
	Policies    orchestration.Policies
	Registry    *registry.Registry
	Checkpoints CheckpointStore
	Sleep       func(time.Duration)
	Now         func() time.Time

	Emit          func(orchestration.ProgressEvent)
	UpdateStatus  func(taskID string, status orchestration.InvestigationStatus) bool
	FetchSnapshot orchestration.SnapshotFetcher

	AgentWorker orchestration.AgentWorkerClient
	AgentRunner func(ctx context.Context, node Node, rc RunContext) (orchestration.AgentResult, error)
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
	if n.AgentDomain == orchestration.DomainCommunications {
		if spec, ok := registry.BuiltinCatalog(n.AgentDomain); ok {
			connectors = spec.Connectors
		}
	}
	return orchestration.AgentTask{
		IncidentID: rc.IncidentID,
		TaskID:     rc.TaskID,
		Domain:     n.AgentDomain,
		FixtureKey: rc.FixtureKey,
		Connectors: connectors,
		TenantID:   rc.TenantID,
	}
}

type noopSnapshot struct{}

func (noopSnapshot) Fetch(context.Context, string, string) (map[string]any, error) {
	return nil, nil
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
