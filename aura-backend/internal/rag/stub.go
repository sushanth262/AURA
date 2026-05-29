package rag

import (
	"context"
	"strings"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

type docEntry struct {
	TenantID  string
	Namespace string
	Title     string
	Snippet   string
}

// stubCatalog is tenant-scoped RAG fixture content (Phase 7 stub).
var stubCatalog = []docEntry{
	{TenantID: "demo", Namespace: "incident_memory", Title: "Prior gateway spike", Snippet: "Similar 5xx pattern on api-gateway in Jan 2025."},
	{TenantID: "demo", Namespace: "runbooks", Title: "Gateway rollback", Snippet: "Rollback upstream pool changes when checkout 5xx elevated."},
	{TenantID: "demo", Namespace: "source_code", Title: "checkout handlers", Snippet: "Timeout changes in handlers.go correlate with deploys."},
	{TenantID: "demo", Namespace: "incident_memory", Title: "Comms overlap", Snippet: "Pager and Slack mentions aligned with telemetry spikes."},
	{TenantID: "other", Namespace: "runbooks", Title: "Other tenant runbook", Snippet: "Must not leak across tenants."},
	{TenantID: "other", Namespace: "incident_memory", Title: "Other tenant memory", Snippet: "Isolated incident memory."},
}

// StubClient returns namespace-filtered fixture retrieval results.
type StubClient struct{}

func (StubClient) Retrieve(_ context.Context, query orchestration.RAGQuery) ([]map[string]any, error) {
	tenant := strings.TrimSpace(query.TenantID)
	if tenant == "" {
		tenant = "demo"
	}
	allowed := namespaceSet(query.Namespaces)
	out := make([]map[string]any, 0, 4)
	for _, d := range stubCatalog {
		if d.TenantID != tenant {
			continue
		}
		if len(allowed) > 0 && !allowed[d.Namespace] {
			continue
		}
		out = append(out, map[string]any{
			"tenant_id":  d.TenantID,
			"namespace":  d.Namespace,
			"title":      d.Title,
			"snippet":    d.Snippet,
			"incident_id": query.IncidentID,
		})
	}
	return out, nil
}

func namespaceSet(namespaces []string) map[string]bool {
	if len(namespaces) == 0 {
		return nil
	}
	out := make(map[string]bool, len(namespaces))
	for _, ns := range namespaces {
		ns = strings.TrimSpace(ns)
		if ns != "" {
			out[ns] = true
		}
	}
	return out
}
