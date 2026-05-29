package graph

import (
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

func TestBuildGraphManifest_DefaultThreeAgents(t *testing.T) {
	reg := registry.DefaultRegistry()
	g, err := Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	m := BuildGraphManifest(g, reg)
	if len(m.Lanes) != 4 {
		t.Fatalf("lanes: got %d want 4 (supervisor + 3 agents)", len(m.Lanes))
	}
	if m.Lanes[0].Domain != "supervisor" {
		t.Fatalf("first lane: got %s", m.Lanes[0].Domain)
	}
	if m.Lanes[3].Domain != "context" {
		t.Fatalf("last agent lane: got %s", m.Lanes[3].Domain)
	}
}

func TestBuildGraphManifest_WithCommunications(t *testing.T) {
	reg := registry.New([]string{"telemetry", "code", "context", "communications"})
	g, err := Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	m := BuildGraphManifest(g, reg)
	if len(m.Lanes) != 5 {
		t.Fatalf("lanes: got %d want 5", len(m.Lanes))
	}
	if m.Lanes[4].Domain != "communications" {
		t.Fatalf("comms lane: got %s", m.Lanes[4].Domain)
	}
	if m.Lanes[4].Label != "Communications" {
		t.Fatalf("comms label: got %q", m.Lanes[4].Label)
	}
}

func TestPlan_CommunicationsDisabled(t *testing.T) {
	reg := registry.New([]string{"telemetry", "code", "context"})
	g, err := Plan(reg, orchestration.DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range g.Nodes {
		if n.AgentDomain == orchestration.DomainCommunications {
			t.Fatal("communications node should not be planned")
		}
	}
}

func TestPlan_TelemetryOnly(t *testing.T) {
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
