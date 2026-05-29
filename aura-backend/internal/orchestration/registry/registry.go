package registry

import (
	"fmt"
	"strings"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// AgentDefinition describes a built-in agent and its connectors.
type AgentDefinition struct {
	Domain      orchestration.AgentDomain
	Connectors  []string
	RAGNamespaces []string
	Enabled     bool
}

// Registry holds the agents available for graph planning.
type Registry struct {
	agents map[orchestration.AgentDomain]AgentDefinition
}

// builtinAgents is the canonical catalog (Phase 2 adds communications).
var builtinAgents = []AgentDefinition{
	{
		Domain:     orchestration.DomainTelemetry,
		Connectors: []string{"grafana"},
		RAGNamespaces: []string{"incident_memory", "runbooks"},
		Enabled:    true,
	},
	{
		Domain:     orchestration.DomainCode,
		Connectors: []string{"github"},
		RAGNamespaces: []string{"source_code", "incident_memory"},
		Enabled:    true,
	},
	{
		Domain:     orchestration.DomainContext,
		Connectors: []string{"jira"},
		RAGNamespaces: []string{"runbooks", "incident_memory"},
		Enabled:    true,
	},
}

// DefaultRegistry enables telemetry, code, and context.
func DefaultRegistry() *Registry {
	enabled := []string{
		string(orchestration.DomainTelemetry),
		string(orchestration.DomainCode),
		string(orchestration.DomainContext),
	}
	return New(enabled)
}

// New builds a registry from a list of enabled agent domain names.
func New(enabledDomains []string) *Registry {
	enableSet := make(map[string]struct{}, len(enabledDomains))
	for _, d := range enabledDomains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "" || d == string(orchestration.DomainSupervisor) {
			continue
		}
		enableSet[d] = struct{}{}
	}

	agents := make(map[orchestration.AgentDomain]AgentDefinition, len(builtinAgents))
	for _, def := range builtinAgents {
		copy := def
		if _, ok := enableSet[string(def.Domain)]; ok {
			copy.Enabled = true
		} else {
			copy.Enabled = false
		}
		agents[def.Domain] = copy
	}
	return &Registry{agents: agents}
}

// EnabledAgents returns definitions with Enabled=true in stable domain order.
func (r *Registry) EnabledAgents() []AgentDefinition {
	order := []orchestration.AgentDomain{
		orchestration.DomainTelemetry,
		orchestration.DomainCode,
		orchestration.DomainContext,
	}
	out := make([]AgentDefinition, 0, len(order))
	for _, d := range order {
		if def, ok := r.agents[d]; ok && def.Enabled {
			out = append(out, def)
		}
	}
	return out
}

// Lookup returns a definition by domain.
func (r *Registry) Lookup(domain orchestration.AgentDomain) (AgentDefinition, bool) {
	def, ok := r.agents[domain]
	return def, ok && def.Enabled
}

// ValidateEnabledDomains returns an error if any name is unknown.
func ValidateEnabledDomains(domains []string) error {
	known := map[string]struct{}{
		string(orchestration.DomainTelemetry): {},
		string(orchestration.DomainCode):      {},
		string(orchestration.DomainContext):   {},
	}
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d == "" || d == string(orchestration.DomainSupervisor) {
			continue
		}
		if _, ok := known[d]; !ok {
			return fmt.Errorf("unknown agent domain %q", d)
		}
	}
	return nil
}
