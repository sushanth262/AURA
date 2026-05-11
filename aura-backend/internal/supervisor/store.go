package supervisor

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu     sync.RWMutex
	byTask map[string]*Investigation
	// incident_id -> task_id quick lookup
	byIncident map[string]string
	history    []IncidentSummary
}

func NewStore() *Store {
	return &Store{
		byTask:     make(map[string]*Investigation),
		byIncident: make(map[string]string),
	}
}

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) Create(inv *Investigation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byTask[inv.TaskID] = inv
	s.byIncident[inv.IncidentID] = inv.TaskID
	sum := IncidentSummary{
		IncidentID: inv.IncidentID,
		Title:      inv.Title,
		Severity:   inv.Severity,
		Status:     inv.Status,
		Scope:      inv.Scope,
		CreatedAt:  inv.CreatedAt.UTC().Format(time.RFC3339),
		ResolvedAt: nil,
	}
	s.history = append([]IncidentSummary{sum}, s.history...)
}

func (s *Store) GetByTask(taskID string) (*Investigation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.byTask[taskID]
	return v, ok
}

func (s *Store) GetByIncident(incidentID string) (*Investigation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tid, ok := s.byIncident[incidentID]
	if !ok {
		return nil, false
	}
	v, ok := s.byTask[tid]
	return v, ok
}

func (s *Store) UpdateStatus(taskID, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.byTask[taskID]
	if !ok {
		return false
	}
	v.Status = status
	v.UpdatedAt = time.Now()
	return true
}

func (s *Store) History(page, perPage int) HistoryPage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	start := (page - 1) * perPage
	end := start + perPage
	total := len(s.history)
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	items := append([]IncidentSummary(nil), s.history[start:end]...)
	return HistoryPage{
		Items:   items,
		Page:    page,
		PerPage: perPage,
		Total:   total,
		Stats: map[string]any{
			"total_count": total,
		},
	}
}

func ScopeFromFixture(sc map[string]any) map[string]any {
	if sc == nil {
		return map[string]any{}
	}
	out := make(map[string]any)
	for k, v := range sc {
		out[strings.ToLower(k)] = v
	}
	if _, ok := out["service"]; !ok && out["cluster"] == nil {
		out["service"] = "unknown"
	}
	return out
}

func NewInvestigationFromSubmission(body IncidentSubmission, fx *FixtureScenario, fixtureKey string) (*Investigation, error) {
	title := strings.TrimSpace(body.Title)
	if title == "" {
		title = fx.DisplayTitle
	}
	sev := strings.TrimSpace(body.Severity)
	if sev == "" {
		sev = fx.Severity
	}
	sc := ScopeFromFixture(fx.Scope)
	if len(body.Scope) > 0 {
		sc = ScopeFromFixture(body.Scope)
	}
	now := time.Now()
	return &Investigation{
		Fixture:    fx,
		FixtureKey: fixtureKey,
		IncidentID: "INC-" + strings.ToUpper(randomHex(4)),
		TaskID:     "TSK-" + strings.ToUpper(randomHex(6)),
		Title:      title,
		Severity:   sev,
		Symptoms:   strings.TrimSpace(body.Symptoms),
		Scope:      sc,
		Status:     "QUEUED",
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func ValidateSubmission(body IncidentSubmission) error {
	hasScenario := strings.TrimSpace(body.ScenarioKey) != ""
	if strings.TrimSpace(body.Title) == "" && strings.TrimSpace(body.Symptoms) == "" && !hasScenario {
		return fmt.Errorf("provide title, symptoms, or scenario_key")
	}
	return nil
}
