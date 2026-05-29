package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// FixtureMCP loads connector mocks from embedded YAML fixtures.
type FixtureMCP struct {
	EnabledSources []string
	LoadYAML       func(baseName string) (map[string]any, error)
	ExtractMock    func(root map[string]any, source string) any
}

func (m FixtureMCP) Fetch(ctx context.Context, connectorID, fixtureKey string) (map[string]any, error) {
	_ = ctx
	connectorID = strings.ToLower(strings.TrimSpace(connectorID))
	fixtureKey = strings.TrimSpace(fixtureKey)
	if connectorID == "" || fixtureKey == "" {
		return nil, nil
	}
	if len(m.EnabledSources) > 0 && !config.SourceEnabled(m.EnabledSources, connectorID) {
		return nil, fmt.Errorf("connector %q disabled", connectorID)
	}
	root, err := m.LoadYAML(fixtureKey)
	if err != nil {
		return nil, err
	}
	payload := m.ExtractMock(root, connectorID)
	if payload == nil {
		return nil, nil
	}
	return map[string]any{
		"scenario_key": fixtureKey,
		"source":       connectorID,
		"mock":         true,
		"payload":      payload,
	}, nil
}

// StubRAG returns empty retrieval results.
type StubRAG struct{}

func (StubRAG) Retrieve(context.Context, orchestration.RAGQuery) ([]map[string]any, error) {
	return nil, nil
}

// PassThroughSecurity applies no redaction (stub).
type PassThroughSecurity struct{}

func (PassThroughSecurity) Redact(_ context.Context, raw map[string]any, _ string) (map[string]any, error) {
	return raw, nil
}
