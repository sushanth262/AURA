package workersvc_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/workersvc"
)

func TestHandleAgentExecute_Contract(t *testing.T) {
	srv := &workersvc.Server{Cfg: config.Config{EnabledSources: []string{"grafana"}}}
	router := srv.Router()

	task := orchestration.AgentTask{
		IncidentID: "inc-2847",
		TaskID:     "task-abc",
		Domain:     orchestration.DomainTelemetry,
		FixtureKey: "inc2847_api_gateway",
		Connectors: []string{"grafana"},
	}
	body, _ := json.Marshal(task)
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/telemetry/execute", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var result orchestration.AgentResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Status != orchestration.AgentSuccess {
		t.Fatalf("status=%q", result.Status)
	}
	if result.Domain != orchestration.DomainTelemetry {
		t.Fatalf("domain=%q", result.Domain)
	}
	if result.FindingCount < 1 && len(result.Findings) < 1 {
		t.Fatal("expected findings")
	}
}

func TestHandleAgentExecute_UnknownDomain(t *testing.T) {
	srv := &workersvc.Server{Cfg: config.Config{}}
	router := srv.Router()
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/unknown/execute", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestHandleAgentExecute_ForbiddenConnector(t *testing.T) {
	srv := &workersvc.Server{Cfg: config.Config{}}
	router := srv.Router()
	task := orchestration.AgentTask{
		Domain:     orchestration.DomainTelemetry,
		Connectors: []string{"github"},
	}
	body, _ := json.Marshal(task)
	req := httptest.NewRequest(http.MethodPost, "/v1/agents/telemetry/execute", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
