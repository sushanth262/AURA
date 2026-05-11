package authzsvc

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sushanth262/AURA/aura-backend/internal/audit"
	"github.com/sushanth262/AURA/aura-backend/internal/authz"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

type Server struct {
	Cfg    config.Config
	Engine authz.Engine
}

type evaluateRequest struct {
	Subject      string   `json:"subject"`
	TenantID     string   `json:"tenant_id"`
	Roles        []string `json:"roles"`
	Action       string   `json:"action"`
	ResourceType string   `json:"resource_type"`
	ResourceID   string   `json:"resource_id"`
	Route        string   `json:"route"`
	TraceID      string   `json:"trace_id"`
}

type evaluateResponse struct {
	Allowed       bool   `json:"allowed"`
	Reason        string `json:"reason"`
	PolicyVersion string `json:"policy_version"`
	DecisionID    string `json:"decision_id"`
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"aura-authz"}`))
	})

	r.Post("/v1/evaluate", s.handleEvaluate)
	return r
}

func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req evaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return
	}
	req.Action = strings.TrimSpace(req.Action)
	if req.Subject == "" || req.Action == "" {
		writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "subject and action required")
		return
	}
	trace := req.TraceID
	if trace == "" {
		trace = middleware.GetReqID(r.Context())
	}
	resource := req.Action + ":" + strings.TrimPrefix(req.Route, "/")
	dec := s.Engine.Authorize(r.Context(), s.Cfg.PolicyVersion, authz.Request{
		Subject:      req.Subject,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		TenantID:     req.TenantID,
		Roles:        req.Roles,
	})
	rec := audit.DecisionRecord{
		PolicyVersion: dec.PolicyVersion,
		Allowed:       dec.Allowed,
		Reason:        dec.Reason,
		Subject:       req.Subject,
		Action:        req.Action,
		Resource:      resource,
		Route:         req.Route,
		TraceID:       trace,
	}
	rec = audit.LogAuthz(rec)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(evaluateResponse{
		Allowed:       dec.Allowed,
		Reason:        dec.Reason,
		PolicyVersion: dec.PolicyVersion,
		DecisionID:    rec.DecisionID,
	})
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error_code": code,
		"message":    msg,
	})
}
