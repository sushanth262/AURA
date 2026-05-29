package workersvc

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/connectors"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

var allowedSources = []string{"grafana", "github", "jira", "slack", "teams", "email"}

type Server struct {
	Cfg     config.Config
	runtime *connectors.Runtime
}

func (s *Server) connectorRuntime() *connectors.Runtime {
	if s.runtime == nil {
		s.runtime = connectors.NewBuiltinRuntime(s.Cfg)
	}
	return s.runtime
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"aura-worker"}`))
	})

	r.Get("/v1/sources/{source}", s.handleSourceMock)
	r.Post("/v1/agents/{domain}/execute", s.handleAgentExecute)
	return r
}

func (s *Server) handleSourceMock(w http.ResponseWriter, r *http.Request) {
	source := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "source")))
	if !isKnownSource(source) {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "unknown source")
		return
	}
	if len(s.Cfg.EnabledSources) > 0 && !config.SourceEnabled(s.Cfg.EnabledSources, source) {
		writeErr(w, http.StatusNotFound, "NOT_AVAILABLE", "source disabled on this worker instance")
		return
	}
	key := strings.TrimSpace(r.URL.Query().Get("scenario_key"))
	if key == "" {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "scenario_key query parameter required")
		return
	}

	out, err := s.connectorRuntime().Invoke(r.Context(), orchestration.ConnectorCall{
		ConnectorID: source,
		ScenarioKey: key,
	})
	if err != nil {
		if errors.Is(err, connectors.ErrCircuitOpen) {
			writeErr(w, http.StatusServiceUnavailable, "CIRCUIT_OPEN", err.Error())
			return
		}
		writeErr(w, http.StatusNotFound, "UNKNOWN_SCENARIO", err.Error())
		return
	}
	if out == nil {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "no mock payload for source")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func isKnownSource(s string) bool {
	for _, x := range allowedSources {
		if x == s {
			return true
		}
	}
	return false
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error_code": code,
		"message":    msg,
	})
}
