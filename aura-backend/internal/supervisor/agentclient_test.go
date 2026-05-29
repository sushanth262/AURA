package supervisor_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/supervisor"
)

func TestHTTPAgentWorkerClient_Execute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/agents/telemetry/execute" {
			http.NotFound(w, r)
			return
		}
		var task orchestration.AgentTask
		_ = json.NewDecoder(r.Body).Decode(&task)
		if task.IncidentID != "inc-1" {
			t.Errorf("incident_id=%q", task.IncidentID)
		}
		_ = json.NewEncoder(w).Encode(orchestration.AgentResult{
			Domain:       orchestration.DomainTelemetry,
			TaskID:       task.TaskID,
			Status:       orchestration.AgentSuccess,
			FindingCount: 2,
			CompletedAt:  time.Now().UTC(),
		})
	}))
	defer ts.Close()

	client := supervisor.NewAgentWorkerClient(config.Config{WorkerURL: ts.URL})
	result, err := client.Execute(context.Background(), orchestration.DomainTelemetry, orchestration.AgentTask{
		IncidentID: "inc-1",
		TaskID:     "task-1",
		FixtureKey: "inc2847_api_gateway",
		Connectors: []string{"grafana"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != orchestration.AgentSuccess || result.FindingCount != 2 {
		t.Fatalf("result=%+v", result)
	}
}
