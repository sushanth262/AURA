package graph

import (
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
)

// GraphLane is one swimlane in the investigation UI.
type GraphLane struct {
	Domain string `json:"domain"`
	Label  string `json:"label"`
	Color  string `json:"color"`
}

// GraphManifest is emitted on GRAPH_PLANNED for dynamic UI swimlanes.
type GraphManifest struct {
	GraphVersion int         `json:"graph_version"`
	Lanes        []GraphLane `json:"lanes"`
	NodeIDs      []string    `json:"node_ids,omitempty"`
}

// BuildGraphManifest builds lane metadata from the planned graph and registry.
func BuildGraphManifest(g InvestigationGraph, reg *registry.Registry) GraphManifest {
	lanes := []GraphLane{
		{
			Domain: string(orchestration.DomainSupervisor),
			Label:  registry.SupervisorLane().Label,
			Color:  registry.SupervisorLane().Color,
		},
	}
	if reg != nil {
		for _, def := range reg.EnabledAgents() {
			lanes = append(lanes, GraphLane{
				Domain: string(def.Domain),
				Label:  def.Label,
				Color:  def.Color,
			})
		}
	}
	return GraphManifest{
		GraphVersion: g.Version,
		Lanes:        lanes,
		NodeIDs:      g.NodeIDs(),
	}
}

func manifestPayload(g InvestigationGraph, reg *registry.Registry) map[string]any {
	m := BuildGraphManifest(g, reg)
	return map[string]any{
		"graph_manifest": m,
	}
}
