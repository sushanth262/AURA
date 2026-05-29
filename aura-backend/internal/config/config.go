package config

import (
	"os"
	"strconv"
	"strings"
)

// OrchestrationPolicies builds graph policies from env-backed config.
func (c Config) OrchestrationPolicies() (minAgents int, join string) {
	minAgents = c.MinAgentsForSynthesis
	if minAgents < 1 {
		minAgents = 1
	}
	join = c.SynthesisJoin
	if join == "" {
		join = "any_success"
	}
	return minAgents, join
}

// Config is loaded from environment (12-factor). Defaults suit local dev + docker-compose.
// Each binary reads only the fields it needs; shared JWT/policy/env vars stay aligned across services.
type Config struct {
	HTTPAddr string

	// AUTH_DEV_MOCK=true → HS256 bearer tokens signed with AUTH_DEV_JWT_SECRET (see /v1/auth/dev-token).
	AuthDevMock       bool
	AuthDevJWTSecret  string
	AuthIssuer        string
	AuthAudience      string
	AuthJWKSURL       string

	CORSAllowedOrigins []string

	PolicyVersion string

	// BFF → aura-authz (HTTP).
	AuthzURL string

	// BFF → aura-supervisor (HTTP + WS proxy target base URL).
	SupervisorURL string

	// Optional shared secret for BFF→supervisor internal routes (leave unset only on trusted networks).
	InternalSharedSecret string

	// Supervisor → aura-worker (connector mocks).
	WorkerURL string
	// Supervisor: comma-separated enabled connector ids matching worker routes (grafana, github, jira).
	WorkerSources []string

	// aura-worker: which connectors this process exposes (subset); empty defaults to all three mocks.
	EnabledSources []string

	// Supervisor graph engine: "engine" (default) or "legacy" (RunMockScenario).
	GraphEngineMode string
	// Comma-separated agent domains for graph planning (telemetry, code, context).
	EnabledAgents []string
	MinAgentsForSynthesis int
	SynthesisJoin         string

	// Supervisor agent dispatch: "inline" (snapshot fetcher) or "worker" (POST execute).
	AgentExecutionMode string
}

func Load() Config {
	c := Config{
		HTTPAddr:              getenv("HTTP_ADDR", ":8080"),
		AuthDevMock:           getenvBool("AUTH_DEV_MOCK", true),
		AuthDevJWTSecret:      getenv("AUTH_DEV_JWT_SECRET", "aura-dev-secret-change-me-use-32chars-minimum!!"),
		AuthIssuer:            getenv("AUTH_ISSUER", ""),
		AuthAudience:          getenv("AUTH_AUDIENCE", ""),
		AuthJWKSURL:           getenv("AUTH_JWKS_URL", ""),
		CORSAllowedOrigins:    splitCSV(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:8081,http://localhost:19006")),
		PolicyVersion:         getenv("POLICY_VERSION", "stub-v1"),
		AuthzURL:              getenv("AUTHZ_URL", "http://127.0.0.1:8081"),
		SupervisorURL:         getenv("SUPERVISOR_URL", "http://127.0.0.1:8082"),
		InternalSharedSecret:  getenv("INTERNAL_SHARED_SECRET", ""),
		WorkerURL:             getenv("WORKER_URL", ""),
		WorkerSources:         normalizeSourceList(splitCSV(getenv("WORKER_SOURCES", "grafana,github,jira"))),
		EnabledSources:        normalizeSourceList(splitCSV(getenv("WORKER_ENABLED_SOURCES", "grafana,github,jira"))),
		GraphEngineMode:       strings.ToLower(strings.TrimSpace(getenv("GRAPH_ENGINE_MODE", "engine"))),
		EnabledAgents:         normalizeAgentList(splitCSV(getenv("ENABLED_AGENTS", "telemetry,code,context"))),
		MinAgentsForSynthesis: getenvInt("MIN_AGENTS_FOR_SYNTHESIS", 1),
		SynthesisJoin:         strings.ToLower(strings.TrimSpace(getenv("SYNTHESIS_JOIN", "any_success"))),
		AgentExecutionMode:    strings.ToLower(strings.TrimSpace(getenv("AGENT_EXECUTION_MODE", "inline"))),
	}
	return c
}

func normalizeAgentList(in []string) []string {
	return normalizeSourceList(in)
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func normalizeSourceList(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// SourceEnabled reports whether name is in the normalized list (case-insensitive).
func SourceEnabled(list []string, name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, x := range list {
		if x == name {
			return true
		}
	}
	return false
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvBool(k string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
