package graph

import (
	"fmt"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

const parallelRetrieve = "retrieve"

// Plan builds an InvestigationGraph from the agent registry and policies.
func Plan(reg *registry.Registry, policies orchestration.Policies) (InvestigationGraph, error) {
	if reg == nil {
		return InvestigationGraph{}, fmt.Errorf("registry is required")
	}
	enabled := reg.EnabledAgents()
	if len(enabled) == 0 {
		return InvestigationGraph{}, fmt.Errorf("at least one agent must be enabled")
	}

	nodes := []Node{
		{ID: "supervisor_claim", Kind: NodeKindSupervisor, AgentDomain: orchestration.DomainSupervisor, TimelineIndex: 0},
		{ID: "supervisor_plan_start", Kind: NodeKindSupervisor, AgentDomain: orchestration.DomainSupervisor, TimelineIndex: 1},
		{ID: "supervisor_plan_done", Kind: NodeKindSupervisor, AgentDomain: orchestration.DomainSupervisor, TimelineIndex: 2},
	}
	edges := []Edge{
		{From: "supervisor_claim", To: "supervisor_plan_start"},
		{From: "supervisor_plan_start", To: "supervisor_plan_done"},
	}

	prev := "supervisor_plan_done"
	for i, def := range enabled {
		startID := fmt.Sprintf("%s_start", def.Domain)
		doneID := fmt.Sprintf("%s_done", def.Domain)
		nodes = append(nodes,
			Node{
				ID:            startID,
				Kind:          NodeKindAgentStart,
				AgentDomain:   def.Domain,
				ParallelGroup: parallelRetrieve,
				Connector:     primaryConnector(def),
				TimelineIndex: domainTimelineIndex(def.Domain),
			},
			Node{
				ID:            doneID,
				Kind:          NodeKindAgentDone,
				AgentDomain:   def.Domain,
				ParallelGroup: parallelRetrieve,
				Connector:     primaryConnector(def),
			},
		)
		edges = append(edges,
			Edge{From: prev, To: startID},
			Edge{From: startID, To: doneID},
		)
		if i == 0 {
			prev = doneID
		} else {
			edges = append(edges, Edge{From: fmt.Sprintf("%s_done", enabled[i-1].Domain), To: startID})
		}
	}

	lastDone := fmt.Sprintf("%s_done", enabled[len(enabled)-1].Domain)
	synthesisID := "synthesis"
	nodes = append(nodes, Node{ID: synthesisID, Kind: NodeKindSynthesis, AgentDomain: orchestration.DomainSupervisor})
	edges = append(edges, Edge{
		From: lastDone,
		To:   synthesisID,
		Join: joinModeFromPolicy(policies),
	})

	return InvestigationGraph{
		Version:       1,
		Nodes:         nodes,
		Edges:         edges,
		ParallelGroup: parallelRetrieve,
	}, nil
}

func primaryConnector(def registry.AgentDefinition) string {
	if len(def.Connectors) > 0 {
		return def.Connectors[0]
	}
	return ""
}

func joinModeFromPolicy(p orchestration.Policies) JoinMode {
	if p.SynthesisJoin == orchestration.JoinAllRequired {
		return JoinModeAllRequired
	}
	return JoinModeAnySuccess
}
