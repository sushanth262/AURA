package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/connectors"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// RuntimeMCP adapts connectors.Runtime to the pipeline MCPClient interface.
type RuntimeMCP struct {
	Runtime        *connectors.Runtime
	EnabledSources []string
}

func (m RuntimeMCP) Fetch(ctx context.Context, connectorID, fixtureKey string) (map[string]any, error) {
	connectorID = strings.ToLower(strings.TrimSpace(connectorID))
	fixtureKey = strings.TrimSpace(fixtureKey)
	if connectorID == "" || fixtureKey == "" {
		return nil, nil
	}
	if len(m.EnabledSources) > 0 && !config.SourceEnabled(m.EnabledSources, connectorID) {
		return nil, fmt.Errorf("connector %q disabled", connectorID)
	}
	return m.Runtime.Invoke(ctx, orchestration.ConnectorCall{
		ConnectorID: connectorID,
		ScenarioKey: fixtureKey,
	})
}
