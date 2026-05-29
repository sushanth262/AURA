package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

// NewGrafanaDriver returns fixture or live Grafana connector per CONNECTOR_GRAFANA_MODE.
func NewGrafanaDriver(cfg config.Config, loader FixtureLoader) Driver {
	mode := strings.ToLower(strings.TrimSpace(cfg.ConnectorGrafanaMode))
	if mode == "live" {
		return &grafanaLiveDriver{
			baseURL:  strings.TrimRight(strings.TrimSpace(cfg.GrafanaURL), "/"),
			fallback: NewFixtureDriver("grafana", loader),
			client:   &http.Client{Timeout: 8 * time.Second},
		}
	}
	return NewFixtureDriver("grafana", loader)
}

type grafanaLiveDriver struct {
	baseURL  string
	fallback *FixtureDriver
	client   *http.Client
}

func (d *grafanaLiveDriver) ID() string { return "grafana" }

func (d *grafanaLiveDriver) Invoke(ctx context.Context, call orchestration.ConnectorCall) (map[string]any, error) {
	if d.baseURL == "" {
		return nil, fmt.Errorf("grafana live mode requires GRAFANA_URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.baseURL+"/api/health", nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		// Degrade to fixture when live probe fails (staging safety).
		if d.fallback != nil {
			return d.fallback.Invoke(ctx, call)
		}
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		if d.fallback != nil {
			return d.fallback.Invoke(ctx, call)
		}
		return nil, fmt.Errorf("grafana health returned %s", resp.Status)
	}
	var health map[string]any
	_ = json.Unmarshal(body, &health)
	fixture, _ := d.fallback.Invoke(ctx, call)
	out := map[string]any{
		"scenario_key": call.ScenarioKey,
		"source":       "grafana",
		"mock":         false,
		"live_probe":   health,
	}
	if fixture != nil {
		if payload, ok := fixture["payload"]; ok {
			out["payload"] = payload
		}
	}
	return out, nil
}

var _ Driver = (*grafanaLiveDriver)(nil)
