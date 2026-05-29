// Package mocks provides test doubles for orchestration interfaces.
package mocks

import (
	"context"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// SnapshotFetcher returns fixed payloads per source.
type SnapshotFetcher struct {
	Data map[string]map[string]any
	Err  error
}

func (m SnapshotFetcher) Fetch(_ context.Context, source, _ string) (map[string]any, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Data != nil {
		if v, ok := m.Data[source]; ok {
			return v, nil
		}
	}
	return nil, nil
}

// SecurityPassThrough implements SecurityClient without redaction (tests only).
type SecurityPassThrough struct{}

func (SecurityPassThrough) Redact(_ context.Context, raw map[string]any, _ string) (map[string]any, error) {
	return raw, nil
}

// StaticAgentExecutor runs a fixed result for one domain.
type StaticAgentExecutor struct {
	AgentDomain orchestration.AgentDomain
	Result      orchestration.AgentResult
	Err         error
}

func (e StaticAgentExecutor) Domain() orchestration.AgentDomain { return e.AgentDomain }

func (e StaticAgentExecutor) RequiredConnectors() []string { return nil }

func (e StaticAgentExecutor) Execute(context.Context, orchestration.AgentTask) (orchestration.AgentResult, error) {
	if e.Err != nil {
		return orchestration.AgentResult{}, e.Err
	}
	return e.Result, nil
}
