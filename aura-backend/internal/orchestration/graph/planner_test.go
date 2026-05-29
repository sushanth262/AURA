package graph

import (
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

func TestPlan_DefaultThreeAgents(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, err := Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	if g.Version != 1 {
		t.Fatalf("version: got %d", g.Version)
	}
	if len(g.Nodes) < 10 {
		t.Fatalf("expected at least 10 nodes, got %d", len(g.Nodes))
	}
	starts := 0
	for _, n := range g.Nodes {
		if n.Kind == NodeKindAgentStart {
			starts++
		}
	}
	if starts != 3 {
		t.Fatalf("agent starts: got %d", starts)
	}
}

func TestPlan_OneAgent(t *testing.T) {
	reg := registry.New([]string{"telemetry"})
	g, err := Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	starts := 0
	for _, n := range g.Nodes {
		if n.Kind == NodeKindAgentStart {
			starts++
		}
	}
	if starts != 1 {
		t.Fatalf("agent starts: got %d", starts)
	}
}

func TestPlan_NoAgentsFails(t *testing.T) {
	reg := registry.New([]string{})
	_, err := Plan(reg, orchestration.DefaultPolicies())
	if err == nil {
		t.Fatal("expected error with no agents")
	}
}
