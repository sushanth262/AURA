package pipeline_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	orchregistry "github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
	"github.com/sushanth262/AURA/aura-backend/internal/security"
	"github.com/sushanth262/AURA/aura-backend/internal/workersvc/pipeline"
)

func TestExecutor_SecurityRedactsEmailInSnapshot(t *testing.T) {
	spec, ok := orchregistry.BuiltinCatalog(orchestration.DomainCommunications)
	if !ok {
		t.Fatal("communications catalog missing")
	}

	exec := pipeline.Executor{
		MCP: pipeline.FixtureMCP{
			EnabledSources: []string{"email"},
			LoadYAML: func(string) (map[string]any, error) {
				return map[string]any{
					"source_mocks": map[string]any{
						"email": map[string]any{
							"threads": []any{
								map[string]any{
									"messages": []any{
										map[string]any{
											"from": "leak@customer.com",
											"body": "Outage details sent to ops@example.com",
										},
									},
								},
							},
						},
					},
				}, nil
			},
			ExtractMock: func(root map[string]any, source string) any {
				sm, _ := root["source_mocks"].(map[string]any)
				return sm[source]
			},
		},
		RAG:           pipeline.NewRAGClient(config.Config{RAGMode: "stub"}),
		Security:      pipeline.NewSecurityClient(config.Config{SecurityMode: "inline"}),
		DefaultTenant: "demo",
	}

	result, err := exec.Run(context.Background(), orchestration.AgentTask{
		Domain:     orchestration.DomainCommunications,
		FixtureKey: "test",
		Connectors: []string{"email"},
		TenantID:   "demo",
	}, spec)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(result.ConnectorSnapshot)
	if security.ContainsRawEmail(string(b)) {
		t.Fatalf("snapshot still contains email: %s", b)
	}
}
