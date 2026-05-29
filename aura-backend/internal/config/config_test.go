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

func TestOrchestrationPolicies(t *testing.T) {
	c := Config{MinAgentsForSynthesis: 2, SynthesisJoin: "all_required"}
	min, join := c.OrchestrationPolicies()
	if min != 2 || join != "all_required" {
		t.Fatalf("got min=%d join=%q", min, join)
	}
}
