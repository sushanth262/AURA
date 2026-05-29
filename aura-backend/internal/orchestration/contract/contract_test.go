package contract

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration/graph"
)

// Progress event shape aligned with aura-frontend/specs/openapi.yaml TaskProgressEvent.
func TestProgressEvent_JSONShape(t *testing.T) {
	ev := orchestration.ProgressEvent{
		TaskID:      "TSK-1",
		IncidentID:  "INC-1",
		EventType:   "AGENT_STARTED",
		AgentDomain: orchestration.DomainTelemetry,
		Payload:     map[string]any{"progress_pct": 52},
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
		SequenceNum: 1,
	}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"task_id", "incident_id", "event_type", "payload", "timestamp", "sequence_num"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing key %q in %s", key, string(b))
		}
	}
	if m["agent_domain"] != string(orchestration.DomainTelemetry) {
		t.Fatalf("agent_domain: %v", m["agent_domain"])
	}
}

func TestAgentResult_JSONShape(t *testing.T) {
	ar := orchestration.AgentResult{
		Domain:      orchestration.DomainCode,
		TaskID:      "TSK-1",
		Status:      orchestration.AgentSuccess,
		FindingCount: 1,
		CompletedAt: time.Now().UTC(),
	}
	b, err := json.Marshal(ar)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"domain", "task_id", "status"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing key %q", key)
		}
	}
}

func TestGraphCheckpoint_JSONShape(t *testing.T) {
	cp := orchestration.GraphCheckpoint{
		TaskID:         "TSK-1",
		IncidentID:     "INC-1",
		CurrentStatus:  orchestration.StatusRetrieving,
		CompletedNodes: []string{"telemetry_start"},
		GraphVersion:   1,
		LastUpdated:    time.Now().UTC(),
	}
	b, err := json.Marshal(cp)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"task_id", "incident_id", "current_status", "completed_nodes", "graph_version"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing key %q", key)
		}
	}
}

func TestGraphPlannedPayload_JSONShape(t *testing.T) {
	m := graph.GraphManifest{
		GraphVersion: 1,
		Lanes: []graph.GraphLane{
			{Domain: "supervisor", Label: "Supervisor", Color: "#1B2B65"},
			{Domain: "communications", Label: "Communications", Color: "#F59E0B"},
		},
	}
	ev := orchestration.ProgressEvent{
		TaskID:      "TSK-1",
		IncidentID:  "INC-1",
		EventType:   "GRAPH_PLANNED",
		AgentDomain: orchestration.DomainSupervisor,
		Payload:     map[string]any{"graph_manifest": m},
		SequenceNum: 4,
		Timestamp:   time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}
	payload, _ := root["payload"].(map[string]any)
	manifest, _ := payload["graph_manifest"].(map[string]any)
	lanes, _ := manifest["lanes"].([]any)
	if len(lanes) != 2 {
		t.Fatalf("lanes: got %d", len(lanes))
	}
}
