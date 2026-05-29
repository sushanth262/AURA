package graph

import (
	"sync"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// CheckpointStore persists graph execution progress.
type CheckpointStore interface {
	Load(taskID string) (orchestration.GraphCheckpoint, bool)
	Save(cp orchestration.GraphCheckpoint)
	MarkNodeComplete(taskID, nodeID string, status orchestration.InvestigationStatus)
	MarkNodeFailed(taskID, nodeID string)
}

// MemoryCheckpointStore is an in-process CheckpointStore for dev and tests.
type MemoryCheckpointStore struct {
	mu   sync.RWMutex
	data map[string]orchestration.GraphCheckpoint
}

func NewMemoryCheckpointStore() *MemoryCheckpointStore {
	return &MemoryCheckpointStore{data: make(map[string]orchestration.GraphCheckpoint)}
}

func (s *MemoryCheckpointStore) Load(taskID string) (orchestration.GraphCheckpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp, ok := s.data[taskID]
	return cp, ok
}

func (s *MemoryCheckpointStore) Save(cp orchestration.GraphCheckpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp.LastUpdated = time.Now().UTC()
	s.data[cp.TaskID] = cp
}

func (s *MemoryCheckpointStore) MarkNodeComplete(taskID, nodeID string, status orchestration.InvestigationStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := s.data[taskID]
	if cp.TaskID == "" {
		cp.TaskID = taskID
	}
	cp.CurrentStatus = status
	cp.CompletedNodes = appendUnique(cp.CompletedNodes, nodeID)
	cp.LastUpdated = time.Now().UTC()
	s.data[taskID] = cp
}

func (s *MemoryCheckpointStore) MarkNodeFailed(taskID, nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := s.data[taskID]
	if cp.TaskID == "" {
		cp.TaskID = taskID
	}
	cp.FailedNodes = appendUnique(cp.FailedNodes, nodeID)
	cp.LastUpdated = time.Now().UTC()
	s.data[taskID] = cp
}

func appendUnique(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}
