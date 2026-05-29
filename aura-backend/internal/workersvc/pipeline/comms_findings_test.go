package pipeline_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/fixturesdata"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	orchregistry "github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
	"github.com/sushanth262/AURA/aura-backend/internal/workersvc/pipeline"
	"gopkg.in/yaml.v3"
)

func TestExecutor_CommunicationsMultiConnectorFindings(t *testing.T) {
	spec, ok := orchregistry.BuiltinCatalog(orchestration.DomainCommunications)
	if !ok {
		t.Fatal("communications catalog missing")
	}
	if len(spec.Connectors) < 3 {
		t.Fatalf("connectors: got %v", spec.Connectors)
	}

	exec := pipeline.Executor{
		MCP: pipeline.FixtureMCP{
			EnabledSources: []string{"slack", "teams", "email"},
			LoadYAML:       loadScenarioYAML,
			ExtractMock:    extractSourceMock,
		},
		RAG:           pipeline.NewRAGClient(config.Config{RAGMode: "stub"}),
		Security:      pipeline.NewSecurityClient(config.Config{SecurityMode: "inline"}),
		DefaultTenant: "demo",
	}

	result, err := exec.Run(context.Background(), orchestration.AgentTask{
		Domain:     orchestration.DomainCommunications,
		FixtureKey: "inc2847_api_gateway",
		Connectors: spec.Connectors,
	}, spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Findings) < 3 {
		t.Fatalf("findings: got %d want >= 3", len(result.Findings))
	}
	types := map[string]bool{}
	for _, f := range result.Findings {
		types[f.Type] = true
	}
	for _, want := range []string{"CHANNEL_ALERT_MENTION", "ONCALL_PING", "EMAIL_THREAD"} {
		if !types[want] {
			t.Fatalf("missing finding type %s", want)
		}
	}
}

func loadScenarioYAML(baseName string) (map[string]any, error) {
	b, err := fixturesdata.FS.ReadFile(baseName + ".yaml")
	if err != nil {
		return nil, err
	}
	body := bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	var root map[string]any
	if err := yaml.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	return root, nil
}

func extractSourceMock(root map[string]any, source string) any {
	sm, _ := root["source_mocks"].(map[string]any)
	if sm == nil {
		return nil
	}
	return sm[source]
}
