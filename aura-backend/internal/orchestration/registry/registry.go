package registry

import (
	"fmt"
	"strings"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// Registry holds the agents available for graph planning.
type Registry struct {
	agents map[orchestration.AgentDomain]AgentDefinition
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
	out := make([]AgentDefinition, 0, len(catalogOrder))
	for _, d := range catalogOrder {
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

// KnownDomains returns all built-in agent domain names (excluding supervisor).
func KnownDomains() []string {
	out := make([]string, len(catalogOrder))
	for i, d := range catalogOrder {
		out[i] = string(d)
	}
	return out
}

// ValidateEnabledDomains returns an error if any name is unknown.
func ValidateEnabledDomains(domains []string) error {
	known := make(map[string]struct{}, len(catalogOrder))
	for _, d := range catalogOrder {
		known[string(d)] = struct{}{}
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
