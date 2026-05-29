package workersvc

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	orchregistry "github.com/sushanth262/AURA/aura-backend/internal/orchestration/registry"
	"github.com/sushanth262/AURA/aura-backend/internal/workersvc/pipeline"
)

func (s *Server) handleAgentExecute(w http.ResponseWriter, r *http.Request) {
	domain := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "domain")))
	spec, ok := orchregistry.BuiltinCatalog(orchestration.AgentDomain(domain))
	if !ok {
		writeErr(w, http.StatusNotFound, "NOT_FOUND", "unknown agent domain")
		return
	}

	var task orchestration.AgentTask
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}
	if task.Domain != "" && string(task.Domain) != domain {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "task domain mismatch")
		return
	}
	task.Domain = orchestration.AgentDomain(domain)

	exec := pipeline.Executor{
		MCP: pipeline.RuntimeMCP{
			Runtime:        s.connectorRuntime(),
			EnabledSources: s.Cfg.EnabledSources,
		},
		RAG:      pipeline.StubRAG{},
		Security: pipeline.PassThroughSecurity{},
	}

	result, err := exec.Run(r.Context(), task, spec)
	if err != nil {
		if _, denied := err.(pipeline.ErrConnectorDenied); denied {
			writeErr(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, "AGENT_FAILED", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}
