package registry

import "github.com/sushanth262/AURA/aura-backend/internal/orchestration"

// BuiltinCatalog returns a built-in agent definition by domain (ignores enabled flag).
func BuiltinCatalog(domain orchestration.AgentDomain) (AgentDefinition, bool) {
	for _, def := range builtinAgents {
		if def.Domain == domain {
			return def, true
		}
	}
	return AgentDefinition{}, false
}

// ConnectorAllowed reports whether conn is listed on the agent definition.
func ConnectorAllowed(def AgentDefinition, conn string) bool {
	for _, c := range def.Connectors {
		if c == conn {
			return true
		}
	}
	return false
}
