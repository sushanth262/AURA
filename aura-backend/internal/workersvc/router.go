package workersvc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/fixturesdata"
	"gopkg.in/yaml.v3"
)

var allowedSources = []string{"grafana", "github", "jira"}

type Server struct {
	Cfg config.Config
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

	root, err := loadScenarioYAML(key)
	if err != nil {
		writeErr(w, http.StatusNotFound, "UNKNOWN_SCENARIO", err.Error())
		return
	}

	payload := extractSourceMock(root, source)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"scenario_key": key,
		"source":       source,
		"mock":         true,
		"payload":      payload,
	})
}

func isKnownSource(s string) bool {
	for _, x := range allowedSources {
		if x == s {
			return true
		}
	}
	return false
}

func loadScenarioYAML(baseName string) (map[string]any, error) {
	b, err := fixturesdata.FS.ReadFile(baseName + ".yaml")
	if err != nil {
		return nil, fmt.Errorf("fixture %q: %w", baseName, err)
	}
	body := bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	var root map[string]any
	if err := yaml.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	return root, nil
}

func extractSourceMock(root map[string]any, source string) any {
	sm, _ := root["source_mocks"].(map[string]any)
	if sm == nil {
		return nil
	}
	return sm[source]
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error_code": code,
		"message":    msg,
	})
}
