package config

import (
	"os"
	"testing"
)

func TestLoad_GraphEngineDefaults(t *testing.T) {
	os.Unsetenv("GRAPH_ENGINE_MODE")
	os.Unsetenv("ENABLED_AGENTS")
	c := Load()
	if c.GraphEngineMode != "engine" {
		t.Fatalf("GraphEngineMode: got %q", c.GraphEngineMode)
	}
	if len(c.EnabledAgents) != 3 {
		t.Fatalf("EnabledAgents: got %v", c.EnabledAgents)
	}
}

func TestLoad_ConnectorGrafanaDefaults(t *testing.T) {
	os.Unsetenv("CONNECTOR_GRAFANA_MODE")
	os.Unsetenv("GRAFANA_URL")
	c := Load()
	if c.ConnectorGrafanaMode != "fixture" {
		t.Fatalf("ConnectorGrafanaMode: got %q", c.ConnectorGrafanaMode)
	}
}

func TestLoad_CheckpointBackendDefaults(t *testing.T) {
	os.Unsetenv("CHECKPOINT_BACKEND")
	os.Unsetenv("REDIS_URL")
	c := Load()
	if c.CheckpointBackend != "memory" {
		t.Fatalf("CheckpointBackend: got %q", c.CheckpointBackend)
	}
}

func TestLoad_RAGSecurityDefaults(t *testing.T) {
	os.Unsetenv("RAG_MODE")
	os.Unsetenv("RAG_SERVICE_URL")
	os.Unsetenv("SECURITY_MODE")
	os.Unsetenv("SECURITY_SERVICE_URL")
	os.Unsetenv("DEFAULT_TENANT_ID")
	os.Unsetenv("SYNTHESIS_LLM_MODE")
	c := Load()
	if c.RAGMode != "stub" || c.SecurityMode != "inline" {
		t.Fatalf("RAG/Security: got rag=%q security=%q", c.RAGMode, c.SecurityMode)
	}
	if c.DefaultTenantID != "demo" || c.SynthesisLLMMode != "off" {
		t.Fatalf("tenant/llm: got tenant=%q llm=%q", c.DefaultTenantID, c.SynthesisLLMMode)
	}
}

func TestOrchestrationPolicies(t *testing.T) {
	c := Config{MinAgentsForSynthesis: 2, SynthesisJoin: "all_required"}
	min, join := c.OrchestrationPolicies()
	if min != 2 || join != "all_required" {
		t.Fatalf("got min=%d join=%q", min, join)
	}
}
