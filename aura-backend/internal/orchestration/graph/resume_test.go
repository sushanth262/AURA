package graph_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/graph"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

func TestRunner_ResumeSkipsCompletedAgents(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, err := graph.Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}

	cpStore := graph.NewMemoryCheckpointStore()
	cpStore.Save(orchestration.GraphCheckpoint{
		TaskID:        "TSK-R1",
		IncidentID:    "INC-R1",
		CurrentStatus: orchestration.StatusRetrieving,
		GraphVersion:  g.Version,
		CompletedNodes: []string{
			"supervisor_claim", "supervisor_plan_start", "supervisor_plan_done", "graph_planned",
			"telemetry_start", "telemetry_done",
			"code_start", "code_done",
			"context_start", "context_done",
		},
		AgentResults: map[string]orchestration.AgentResult{
			"telemetry": {Domain: orchestration.DomainTelemetry, Status: orchestration.AgentSuccess, FindingCount: 2},
			"code":      {Domain: orchestration.DomainCode, Status: orchestration.AgentSuccess, FindingCount: 1},
			"context":   {Domain: orchestration.DomainContext, Status: orchestration.AgentSuccess, FindingCount: 1},
		},
	})

	var events []orchestration.ProgressEvent
	runner := &graph.Runner{
		Policies:     orchestration.DefaultPolicies(),
		Registry:     reg,
		Checkpoints:  cpStore,
		Sleep:        func(time.Duration) {},
		Emit:         func(ev orchestration.ProgressEvent) { events = append(events, ev) },
		UpdateStatus: func(_ string, _ orchestration.InvestigationStatus) bool { return true },
		AgentRunner: func(_ context.Context, _ graph.Node, _ graph.RunContext) (orchestration.AgentResult, error) {
			t.Fatal("resume should not re-execute completed agents")
			return orchestration.AgentResult{}, nil
		},
	}

	rc := graph.RunContext{
		TaskID: "TSK-R1", IncidentID: "INC-R1", Title: "resume",
		Scope: map[string]any{"service": "api"}, PolicyVersion: "stub-v1",
	}
	if err := runner.Run(context.Background(), g, rc); err != nil {
		t.Fatal(err)
	}

	completeCount, skippedCount, synthCount := 0, 0, 0
	for _, ev := range events {
		switch ev.EventType {
		case "AGENT_COMPLETE":
			completeCount++
		case "AGENT_SKIPPED":
			skippedCount++
		case "SYNTHESIS_COMPLETE":
			synthCount++
		}
	}
	if completeCount > 0 {
		t.Fatalf("resume should not re-emit AGENT_COMPLETE, got %d", completeCount)
	}
	if skippedCount != 3 {
		t.Fatalf("expected 3 AGENT_SKIPPED, got %d", skippedCount)
	}
	if synthCount != 1 {
		t.Fatalf("expected SYNTHESIS_COMPLETE, got %d", synthCount)
	}
}

func TestRunner_CancelMidGraphThenResume(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, err := graph.Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}

	cpStore := graph.NewMemoryCheckpointStore()
	var ran int32
	runner := &graph.Runner{
		Policies:    orchestration.DefaultPolicies(),
		Registry:    reg,
		Checkpoints: cpStore,
		Sleep:       func(time.Duration) {},
		UpdateStatus: func(_ string, _ orchestration.InvestigationStatus) bool { return true },
		AgentRunner: func(ctx context.Context, n graph.Node, _ graph.RunContext) (orchestration.AgentResult, error) {
			nDone := atomic.AddInt32(&ran, 1)
			if n.AgentDomain == orchestration.DomainCode && nDone == 2 {
				return orchestration.AgentResult{}, context.Canceled
			}
			if err := ctx.Err(); err != nil {
				return orchestration.AgentResult{}, err
			}
			return orchestration.AgentResult{
				Domain: n.AgentDomain, Status: orchestration.AgentSuccess, FindingCount: 1,
			}, nil
		},
	}

	rc := graph.RunContext{
		TaskID: "TSK-R2", IncidentID: "INC-R2", Title: "cancel-resume",
		Scope: map[string]any{"service": "api"}, PolicyVersion: "stub-v1",
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runner.Run(ctx, g, rc) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("first run: got %v want context.Canceled", err)
	}

	atomic.StoreInt32(&ran, 0)
	var final orchestration.InvestigationStatus
	runner.Emit = nil
	runner.UpdateStatus = func(_ string, st orchestration.InvestigationStatus) bool {
		final = st
		return true
	}
	if err := runner.Run(context.Background(), g, rc); err != nil {
		t.Fatal(err)
	}
	if final != orchestration.StatusHITLPending {
		t.Fatalf("final status: %s", final)
	}
}

func TestRedisCheckpointStore_RoundTrip(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	store, err := graph.NewRedisCheckpointStore("redis://"+mr.Addr(), "")
	if err != nil {
		t.Fatal(err)
	}
	store.Save(orchestration.GraphCheckpoint{
		TaskID: "TSK-REDIS", IncidentID: "INC-1", CurrentStatus: orchestration.StatusRetrieving,
		CompletedNodes: []string{"telemetry_start"},
		AgentResults: map[string]orchestration.AgentResult{
			"telemetry": {Domain: orchestration.DomainTelemetry, Status: orchestration.AgentSuccess, FindingCount: 2},
		},
	})
	store.MarkNodeComplete("TSK-REDIS", "telemetry_done", orchestration.StatusRetrieving)

	cp, ok := store.Load("TSK-REDIS")
	if !ok {
		t.Fatal("expected checkpoint")
	}
	if len(cp.CompletedNodes) < 2 {
		t.Fatalf("completed nodes: %v", cp.CompletedNodes)
	}
	if cp.AgentResults["telemetry"].FindingCount != 2 {
		t.Fatalf("agent result not persisted")
	}
}

func TestNewCheckpointStore_MemoryDefault(t *testing.T) {
	s, err := graph.NewCheckpointStore(config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.(*graph.MemoryCheckpointStore); !ok {
		t.Fatalf("got %T", s)
	}
}
