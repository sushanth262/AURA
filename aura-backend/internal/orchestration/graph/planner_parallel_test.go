package graph_test

import (
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/graph"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

func TestPlan_FourParallelRetrieveNodes(t *testing.T) {
	reg := registry.New([]string{"telemetry", "code", "context", "communications"})
	g, err := graph.Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	starts := 0
	for _, n := range g.Nodes {
		if n.Kind == graph.NodeKindAgentStart && n.ParallelGroup == "retrieve" {
			starts++
		}
	}
	if starts != 4 {
		t.Fatalf("parallel retrieve starts: got %d want 4", starts)
	}
}
