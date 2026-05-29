package pipeline

import (
	"context"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"github.com/sushanth262/AURA/aura-backend/internal/rag"
	"github.com/sushanth262/AURA/aura-backend/internal/security"
)

type ragAdapter struct{ inner rag.Client }

func (a ragAdapter) Retrieve(ctx context.Context, query orchestration.RAGQuery) ([]map[string]any, error) {
	return a.inner.Retrieve(ctx, query)
}

type securityAdapter struct{ inner security.Client }

func (a securityAdapter) Redact(ctx context.Context, raw map[string]any, source string) (map[string]any, error) {
	return a.inner.Redact(ctx, raw, source)
}

// NewRAGClient builds the RAG stage client from worker config.
func NewRAGClient(cfg config.Config) RAGClient {
	return ragAdapter{inner: rag.NewClient(cfg)}
}

// NewSecurityClient builds the mandatory security stage client from worker config.
func NewSecurityClient(cfg config.Config) SecurityClient {
	return securityAdapter{inner: security.NewClient(cfg)}
}

// ResolveTenantID returns task tenant or configured default.
func ResolveTenantID(task orchestration.AgentTask, cfg config.Config) string {
	if t := task.TenantID; t != "" {
		return t
	}
	if d := cfg.DefaultTenantID; d != "" {
		return d
	}
	return "demo"
}
