package supervisor

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

// HTTPServer hosts supervisor-only HTTP routes (internal REST + investigation WebSocket).
type HTTPServer struct {
	Cfg     config.Config
	Store   *Store
	Hub     *WSClients
	Fetcher SnapshotFetcher
}

func InternalSecretGate(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.TrimSpace(secret) == "" {
				next.ServeHTTP(w, r)
				return
			}
			if r.Header.Get("X-Internal-Secret") != secret {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error_code":"FORBIDDEN","message":"internal secret mismatch"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *HTTPServer) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"aura-supervisor"}`))
	})

	r.Get("/ws/investigations/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		taskID := chi.URLParam(r, "taskId")
		s.Hub.ServeHTTP(w, r, taskID)
	})

	r.Route("/internal/v1", func(r chi.Router) {
		r.Use(InternalSecretGate(s.Cfg.InternalSharedSecret))
		r.Post("/incidents", s.handleCreateIncident)
		r.Get("/incidents/history", s.handleHistory)
		r.Get("/incidents/{incidentID}", s.handleGetIncident)
	})

	return r
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error_code": code,
		"message":    msg,
	})
}

func (s *HTTPServer) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	var body IncidentSubmission
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}
	if err := ValidateSubmission(body); err != nil {
		writeErr(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	fixtureName := strings.TrimSpace(body.ScenarioKey)
	if fixtureName == "" {
		fixtureName = DefaultFixtureName()
	}
	fx, err := LoadFixture(fixtureName)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "UNKNOWN_SCENARIO", err.Error())
		return
	}
	inv, err := NewInvestigationFromSubmission(body, fx, fixtureName)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	s.Store.Create(inv)
	RunMockScenario(s.Cfg, s.Store, s.Hub, inv, s.Fetcher)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(QueuedResponse{
		TaskID:     inv.TaskID,
		IncidentID: inv.IncidentID,
		Status:     "QUEUED",
	})
}

func (s *HTTPServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	page := 1
	perPage := 20
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			perPage = n
		}
	}
	out := s.Store.History(page, perPage)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *HTTPServer) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "incidentID")
	inv, ok := s.Store.GetByIncident(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "incident not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(IncidentStateResponse{
		IncidentID: inv.IncidentID,
		TaskID:     inv.TaskID,
		Status:     inv.Status,
		Severity:   inv.Severity,
		Title:      inv.Title,
		Scope:      inv.Scope,
		CreatedAt:  inv.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  inv.UpdatedAt.UTC().Format(time.RFC3339),
	})
}
