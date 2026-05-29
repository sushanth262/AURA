package graph

import (
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

func TestMemoryCheckpointStore(t *testing.T) {
	s := NewMemoryCheckpointStore()
	s.Save(orchestration.GraphCheckpoint{TaskID: "TSK-1", IncidentID: "INC-1", CurrentStatus: orchestration.StatusIntake})
	s.MarkNodeComplete("TSK-1", "n1", orchestration.StatusRetrieving)
	s.MarkNodeFailed("TSK-1", "n2")

	cp, ok := s.Load("TSK-1")
	if !ok {
		t.Fatal("expected checkpoint")
	}
	if len(cp.CompletedNodes) != 1 || cp.CompletedNodes[0] != "n1" {
		t.Fatalf("completed: %v", cp.CompletedNodes)
	}
	if len(cp.FailedNodes) != 1 || cp.FailedNodes[0] != "n2" {
		t.Fatalf("failed: %v", cp.FailedNodes)
	}
}
