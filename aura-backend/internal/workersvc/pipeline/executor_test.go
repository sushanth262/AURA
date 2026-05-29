package pipeline_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	orchregistry "github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
	"github.com/sushanth262/AURA/aura-backend/internal/workersvc/pipeline"
)

type stageRecorder struct {
	stages []pipeline.Stage
}

func (s *stageRecorder) record(stage pipeline.Stage) {
	s.stages = append(s.stages, stage)
}

type recordingMCP struct {
	called bool
}

func (m *recordingMCP) Fetch(context.Context, string, string) (map[string]any, error) {
	m.called = true
	return map[string]any{"mock": true}, nil
}

type recordingRAG struct {
	called   bool
	afterMCP bool
	mcp      *recordingMCP
}

func (r *recordingRAG) Retrieve(context.Context, orchestration.RAGQuery) ([]map[string]any, error) {
	r.called = true
	r.afterMCP = r.mcp.called
	return nil, nil
}

type recordingSecurity struct {
	called    bool
	afterRAG  bool
	rag       *recordingRAG
}

func (s *recordingSecurity) Redact(context.Context, map[string]any, string) (map[string]any, error) {
	s.called = true
	s.afterRAG = s.rag.called
	return map[string]any{"redacted": true}, nil
}

func TestExecutor_PipelineOrder(t *testing.T) {
	spec, ok := orchregistry.BuiltinCatalog(orchestration.DomainTelemetry)
	if !ok {
		t.Fatal("telemetry catalog missing")
	}
	rec := &stageRecorder{}
	mcp := &recordingMCP{}
	rag := &recordingRAG{mcp: mcp}
	sec := &recordingSecurity{rag: rag}

	exec := pipeline.Executor{
		MCP:      mcp,
		RAG:      rag,
		Security: sec,
		Now:      func() time.Time { return time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC) },
		OnStage:  rec.record,
	}

	result, err := exec.Run(context.Background(), orchestration.AgentTask{
		IncidentID: "inc-1",
		TaskID:     "task-1",
		Domain:     orchestration.DomainTelemetry,
		FixtureKey: "inc2847_api_gateway",
	}, spec)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Status != orchestration.AgentSuccess {
		t.Fatalf("status=%q", result.Status)
	}
	if !mcp.called || !rag.called || !sec.called {
		t.Fatalf("stage clients: mcp=%v rag=%v sec=%v", mcp.called, rag.called, sec.called)
	}
	if !rag.afterMCP {
		t.Fatal("RAG ran before MCP completed")
	}
	if !sec.afterRAG {
		t.Fatal("Security ran before RAG completed")
	}
	want := []pipeline.Stage{pipeline.StageMCP, pipeline.StageRAG, pipeline.StageSecurity}
	if len(rec.stages) != len(want) {
		t.Fatalf("stages=%v want %v", rec.stages, want)
	}
	for i := range want {
		if rec.stages[i] != want[i] {
			t.Fatalf("stage[%d]=%q want %q", i, rec.stages[i], want[i])
		}
	}
}

func TestExecutor_TelemetryCannotUseGitHubConnector(t *testing.T) {
	spec, ok := orchregistry.BuiltinCatalog(orchestration.DomainTelemetry)
	if !ok {
		t.Fatal("telemetry catalog missing")
	}
	exec := pipeline.Executor{
		MCP:      &recordingMCP{},
		RAG:      pipeline.StubRAG{},
		Security: pipeline.PassThroughSecurity{},
	}
	_, err := exec.Run(context.Background(), orchestration.AgentTask{
		Domain:     orchestration.DomainTelemetry,
		Connectors: []string{"github"},
	}, spec)
	if err == nil {
		t.Fatal("expected connector denial")
	}
	var denied pipeline.ErrConnectorDenied
	if !errors.As(err, &denied) {
		t.Fatalf("err type=%T", err)
	}
}

func TestExecutor_CancelledContextMCPError(t *testing.T) {
	spec, ok := orchregistry.BuiltinCatalog(orchestration.DomainCode)
	if !ok {
		t.Fatal("code catalog missing")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	exec := pipeline.Executor{
		MCP: pipeline.FixtureMCP{
			LoadYAML: func(string) (map[string]any, error) {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				return map[string]any{"source_mocks": map[string]any{}}, nil
			},
			ExtractMock: func(map[string]any, string) any { return nil },
		},
		RAG:      pipeline.StubRAG{},
		Security: pipeline.PassThroughSecurity{},
	}

	_, err := exec.Run(ctx, orchestration.AgentTask{
		Domain:     orchestration.DomainCode,
		FixtureKey: "inc2847_api_gateway",
	}, spec)
	if err == nil {
		t.Fatal("expected MCP error on cancelled context")
	}
}
