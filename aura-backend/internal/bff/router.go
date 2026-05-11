package bff

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/jwtauth"
)

type ctxKey int

const principalKey ctxKey = 1

func principalFrom(ctx context.Context) (jwtauth.Principal, bool) {
	p, ok := ctx.Value(principalKey).(jwtauth.Principal)
	return p, ok
}

type Server struct {
	Cfg config.Config
	Out *http.Client
}

func NewServer(cfg config.Config) *Server {
	if cfg.AuthzURL == "" {
		cfg.AuthzURL = "http://127.0.0.1:8081"
	}
	if cfg.SupervisorURL == "" {
		cfg.SupervisorURL = "http://127.0.0.1:8082"
	}
	return &Server{
		Cfg: cfg,
		Out: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(cors(s.Cfg.CORSAllowedOrigins))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"aura-bff-api"}`))
	})

	r.Post("/v1/auth/dev-token", s.handleDevToken)

	r.Get("/ws/investigations/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		s.ProxyInvestigationWebSocket(w, r, chi.URLParam(r, "taskId"))
	})

	r.Route("/v1/api", func(r chi.Router) {
		r.Use(s.requireBearer)
		r.Post("/incidents", s.authzWrap("incident:create", "POST /v1/api/incidents", s.handleCreateIncident))
		r.Get("/incidents/history", s.authzWrap("incident:history", "GET /v1/api/incidents/history", s.handleHistory))
		r.Get("/incidents/by-task/{taskID}", s.authzWrap("incident:read", "GET /v1/api/incidents/by-task/{task}", s.handleGetIncidentByTask))
		r.Get("/incidents/{incidentID}", s.authzWrap("incident:read", "GET /v1/api/incidents/{id}", s.handleGetIncident))
		r.Get("/investigations/{taskID}/evidence", s.authzWrap("incident:read", "GET /v1/api/investigations/{task}/evidence", s.handleGetEvidenceBundle))
	})

	return r
}

func cors(origins []string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range origins {
		allowed[strings.TrimSpace(o)] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) requireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := jwtauth.VerifyBearer(r.Context(), s.Cfg, r.Header.Get("Authorization"))
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
			return
		}
		ctx := context.WithValue(r.Context(), principalKey, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type handlerFn func(http.ResponseWriter, *http.Request)

func (s *Server) authzWrap(action, route string, h handlerFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := principalFrom(r.Context())
		if !ok {
			writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing principal")
			return
		}
		allowed, decID, polVer, reason := s.remoteEvaluate(r, p, action, route)
		if !allowed {
			w.Header().Set("X-Decision-Id", decID)
			w.Header().Set("X-Policy-Version", polVer)
			writeErr(w, http.StatusForbidden, "FORBIDDEN", reason)
			return
		}
		w.Header().Set("X-Decision-Id", decID)
		w.Header().Set("X-Policy-Version", polVer)
		h(w, r)
	}
}

type evaluateAPIResponse struct {
	Allowed       bool   `json:"allowed"`
	Reason        string `json:"reason"`
	PolicyVersion string `json:"policy_version"`
	DecisionID    string `json:"decision_id"`
}

func (s *Server) remoteEvaluate(r *http.Request, p jwtauth.Principal, action, route string) (allowed bool, decisionID, policyVer, reason string) {
	body := map[string]any{
		"subject":        p.Sub,
		"tenant_id":      p.TenantID,
		"roles":          p.Roles,
		"action":         action,
		"resource_type":  "http_route",
		"resource_id":    route,
		"route":          route,
		"trace_id":       middleware.GetReqID(r.Context()),
	}
	buf, _ := json.Marshal(body)
	authzURL := strings.TrimRight(s.Cfg.AuthzURL, "/") + "/v1/evaluate"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, authzURL, bytes.NewReader(buf))
	if err != nil {
		return false, "", "", err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.Out.Do(req)
	if err != nil {
		return false, "", "", "authz unreachable: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return false, "", "", "authz error: " + resp.Status + " " + string(b)
	}
	var out evaluateAPIResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return false, "", "", "authz response parse error"
	}
	return out.Allowed, out.DecisionID, out.PolicyVersion, out.Reason
}

func (s *Server) handleDevToken(w http.ResponseWriter, r *http.Request) {
	if !s.Cfg.AuthDevMock {
		writeErr(w, http.StatusNotFound, "NOT_AVAILABLE", "dev token minting disabled")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body := map[string]any{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	sub := "demo-operator"
	if v, ok := body["sub"].(string); ok && strings.TrimSpace(v) != "" {
		sub = v
	}
	roles := []string{"operator"}
	if arr, ok := body["roles"].([]any); ok && len(arr) > 0 {
		roles = nil
		for _, x := range arr {
			if st, ok := x.(string); ok {
				roles = append(roles, st)
			}
		}
	}
	tenant := "demo"
	if v, ok := body["tenant_id"].(string); ok && v != "" {
		tenant = v
	}
	token, err := jwtauth.MintDevToken(s.Cfg, sub, roles, tenant)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "TOKEN_ERROR", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   28800,
		"sub":          sub,
		"tenant_id":    tenant,
		"roles":        roles,
	})
}

func (s *Server) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	s.forwardSupervisor(w, r, http.MethodPost, "/internal/v1/incidents")
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	path := "/internal/v1/incidents/history"
	if raw := r.URL.RawQuery; raw != "" {
		path += "?" + raw
	}
	s.forwardSupervisor(w, r, http.MethodGet, path)
}

func (s *Server) handleGetIncidentByTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	path := "/internal/v1/incidents/task/" + url.PathEscape(taskID)
	s.forwardSupervisor(w, r, http.MethodGet, path)
}

func (s *Server) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "incidentID")
	path := "/internal/v1/incidents/" + url.PathEscape(id)
	s.forwardSupervisor(w, r, http.MethodGet, path)
}

func (s *Server) handleGetEvidenceBundle(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	path := "/internal/v1/investigations/" + url.PathEscape(taskID) + "/evidence"
	s.forwardSupervisor(w, r, http.MethodGet, path)
}

func (s *Server) forwardSupervisor(w http.ResponseWriter, r *http.Request, method, path string) {
	target := strings.TrimRight(s.Cfg.SupervisorURL, "/") + path
	var body io.Reader
	if r.Body != nil && method == http.MethodPost {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "cannot read body")
			return
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(r.Context(), method, target, body)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "BAD_GATEWAY", err.Error())
		return
	}
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	if sec := strings.TrimSpace(s.Cfg.InternalSharedSecret); sec != "" {
		req.Header.Set("X-Internal-Secret", sec)
	}
	resp, err := s.Out.Do(req)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "BAD_GATEWAY", "supervisor unreachable: "+err.Error())
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error_code": code,
		"message":    msg,
	})
}
