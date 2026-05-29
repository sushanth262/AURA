package graph

import "github.com/sushanth262/AURA/aura-backend/internal/orchestration"

// NodeKind classifies graph nodes for execution and checkpointing.
type NodeKind string

const (
	NodeKindSupervisor NodeKind = "supervisor"
	NodeKindAgentStart NodeKind = "agent_start"
	NodeKindAgentDone  NodeKind = "agent_done"
	NodeKindSynthesis    NodeKind = "synthesis"
)

// JoinMode on edges into synthesis (Phase 2+); planner embeds policy on graph.
type JoinMode string

const (
	JoinModeAnySuccess JoinMode = "any_success"
	JoinModeAllRequired JoinMode = "all_required"
)

// Node is a unit of work in the investigation graph.
type Node struct {
	ID             string                    `json:"id"`
	Kind           NodeKind                  `json:"kind"`
	AgentDomain    orchestration.AgentDomain `json:"agent_domain,omitempty"`
	ParallelGroup  string                    `json:"parallel_group,omitempty"`
	Connector      string                    `json:"connector,omitempty"`
	TimelineIndex  int                       `json:"timeline_index,omitempty"`
}

// Edge connects two nodes (serializable for UI / checkpoint).
type Edge struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Join JoinMode `json:"join,omitempty"`
}

// InvestigationGraph is the planned investigation DAG (mostly linear + parallel retrieve).
type InvestigationGraph struct {
	Version      int    `json:"graph_version"`
	Nodes        []Node `json:"nodes"`
	Edges        []Edge `json:"edges"`
	ParallelGroup string `json:"parallel_group,omitempty"`
}

// NodeIDs returns all node IDs in declaration order.
func (g InvestigationGraph) NodeIDs() []string {
	ids := make([]string, len(g.Nodes))
	for i, n := range g.Nodes {
		ids[i] = n.ID
	}
	return ids
}
