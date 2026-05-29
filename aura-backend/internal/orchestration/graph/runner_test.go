package graph

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

func TestRunner_GoldenEventSequence(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, err := Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}

	var events []orchestration.ProgressEvent
	var statuses []orchestration.InvestigationStatus
	runner := &Runner{
		Policies:    orchestration.DefaultPolicies(),
		Registry:    reg,
		Checkpoints: NewMemoryCheckpointStore(),
		Sleep:       func(time.Duration) {},
		Emit: func(ev orchestration.ProgressEvent) {
			events = append(events, ev)
		},
		UpdateStatus: func(_ string, st orchestration.InvestigationStatus) bool {
			statuses = append(statuses, st)
			return true
		},
	}

	rc := RunContext{
		TaskID:         "TSK-TEST",
		IncidentID:     "INC-TEST",
		Title:          "Test",
		Severity:       "P2",
		Scope:          map[string]any{"service": "api-gateway"},
		FixtureKey:     "inc2847_api_gateway",
		TimelineLabels: []string{"a", "b", "c", "d", "e", "f"},
		ScenarioID:     "inc2847",
		PolicyVersion:  "stub-v1",
	}

	if err := runner.Run(context.Background(), g, rc); err != nil {
		t.Fatal(err)
	}

	wantTypes := []string{
		"TASK_CLAIMED",
		"AGENT_STARTED",
		"AGENT_COMPLETE",
		"GRAPH_PLANNED",
		"AGENT_STARTED",
		"AGENT_STARTED",
		"AGENT_STARTED",
		"AGENT_COMPLETE",
		"AGENT_COMPLETE",
		"AGENT_COMPLETE",
		"SYNTHESIS_COMPLETE",
	}
	if len(events) != len(wantTypes) {
		t.Fatalf("event count: got %d want %d", len(events), len(wantTypes))
	}
	for i, want := range wantTypes {
		if events[i].EventType != want {
			t.Fatalf("event[%d]: got %s want %s", i, events[i].EventType, want)
		}
	}

	// All retrieve starts before any retrieve complete.
	firstComplete := -1
	lastStart := -1
	for i, ev := range events {
		if ev.EventType == "AGENT_STARTED" && ev.AgentDomain == orchestration.DomainTelemetry {
			lastStart = i
		}
		if ev.EventType == "AGENT_COMPLETE" && ev.AgentDomain == orchestration.DomainTelemetry && firstComplete < 0 {
			firstComplete = i
		}
	}
	for i, ev := range events {
		if ev.EventType == "AGENT_STARTED" && (ev.AgentDomain == orchestration.DomainCode || ev.AgentDomain == orchestration.DomainContext) {
			if i > lastStart {
				lastStart = i
			}
		}
	}
	if firstComplete <= lastStart {
		t.Fatalf("telemetry complete at %d should be after last agent start at %d", firstComplete, lastStart)
	}

	last := statuses[len(statuses)-1]
	if last != orchestration.StatusHITLPending {
		t.Fatalf("final status: got %s", last)
	}

	cp, ok := runner.Checkpoints.Load("TSK-TEST")
	if !ok {
		t.Fatal("missing checkpoint")
	}
	if len(cp.CompletedNodes) < 7 {
		t.Fatalf("expected completed nodes, got %v", cp.CompletedNodes)
	}

	// GRAPH_PLANNED carries swimlane manifest.
	var planned *orchestration.ProgressEvent
	for i := range events {
		if events[i].EventType == "GRAPH_PLANNED" {
			planned = &events[i]
			break
		}
	}
	if planned == nil {
		t.Fatal("missing GRAPH_PLANNED event")
	}
	manifest, ok := planned.Payload["graph_manifest"].(GraphManifest)
	if !ok {
		// payload may be map from JSON round-trip in other tests
		if m, ok2 := planned.Payload["graph_manifest"].(map[string]any); ok2 {
			lanes, _ := m["lanes"].([]any)
			if len(lanes) != 4 {
				t.Fatalf("manifest lanes via map: got %d", len(lanes))
			}
		} else {
			t.Fatalf("graph_manifest type: %T", planned.Payload["graph_manifest"])
		}
	} else if len(manifest.Lanes) != 4 {
		t.Fatalf("manifest lanes: got %d", len(manifest.Lanes))
	}
}

func TestRunner_PartialFailureStillSynthesizes(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, _ := Plan(reg, orchestration.DefaultPolicies())

	var final orchestration.InvestigationStatus
	runner := &Runner{
		Policies:    orchestration.DefaultPolicies(),
		Checkpoints: NewMemoryCheckpointStore(),
		Sleep:       func(time.Duration) {},
		UpdateStatus: func(_ string, st orchestration.InvestigationStatus) bool {
			final = st
			return true
		},
		AgentRunner: func(_ context.Context, n Node, _ RunContext) (orchestration.AgentResult, error) {
			if n.AgentDomain == orchestration.DomainCode || n.AgentDomain == orchestration.DomainContext {
				return orchestration.AgentResult{Status: orchestration.AgentFailed}, errors.New("fail")
			}
			return orchestration.AgentResult{Status: orchestration.AgentSuccess, FindingCount: 2}, nil
		},
	}

	rc := RunContext{TaskID: "TSK-P", IncidentID: "INC-P", Scope: map[string]any{"service": "x"}}
	if err := runner.Run(context.Background(), g, rc); err != nil {
		t.Fatal(err)
	}
	if final != orchestration.StatusHITLPending {
		t.Fatalf("final: got %s want HITL_PENDING", final)
	}
}

func TestRunner_AllAgentsFail(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, _ := Plan(reg, orchestration.Policies{MinAgentsForSynthesis: 1, SynthesisJoin: orchestration.JoinAnySuccess})

	var final orchestration.InvestigationStatus
	runner := &Runner{
		Policies:    orchestration.Policies{MinAgentsForSynthesis: 1, SynthesisJoin: orchestration.JoinAnySuccess},
		Checkpoints: NewMemoryCheckpointStore(),
		Sleep:       func(time.Duration) {},
		UpdateStatus: func(_ string, st orchestration.InvestigationStatus) bool {
			final = st
			return true
		},
		AgentRunner: func(context.Context, Node, RunContext) (orchestration.AgentResult, error) {
			return orchestration.AgentResult{Status: orchestration.AgentFailed}, errors.New("fail")
		},
	}

	rc := RunContext{TaskID: "TSK-F", IncidentID: "INC-F"}
	_ = runner.Run(context.Background(), g, rc)
	if final != orchestration.StatusFailed {
		t.Fatalf("final: got %s want FAILED", final)
	}
}

func TestRunner_SequenceNumbersMonotonic(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, _ := Plan(reg, orchestration.DefaultPolicies())
	var events []orchestration.ProgressEvent
	runner := &Runner{
		Policies:    orchestration.DefaultPolicies(),
		Checkpoints: NewMemoryCheckpointStore(),
		Sleep:       func(time.Duration) {},
		Emit:        func(ev orchestration.ProgressEvent) { events = append(events, ev) },
	}
	_ = runner.Run(context.Background(), g, RunContext{TaskID: "TSK-S", IncidentID: "INC-S", Scope: map[string]any{"service": "s"}})
	prev := 0
	for _, ev := range events {
		if ev.SequenceNum <= prev {
			t.Fatalf("sequence not monotonic: %d after %d", ev.SequenceNum, prev)
		}
		prev = ev.SequenceNum
	}
}
