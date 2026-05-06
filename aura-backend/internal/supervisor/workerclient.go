package supervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
)

// SnapshotFetcher pulls mocked connector payloads from aura-worker (or returns nil when disabled).
type SnapshotFetcher interface {
	Fetch(ctx context.Context, source string, scenarioKey string) (map[string]any, error)
}

type noopFetcher struct{}

func (noopFetcher) Fetch(context.Context, string, string) (map[string]any, error) {
	return nil, nil
}

// NewSnapshotFetcher builds an HTTP client for WORKER_URL when set; otherwise returns a no-op fetcher.
func NewSnapshotFetcher(cfg config.Config) SnapshotFetcher {
	base := strings.TrimSpace(cfg.WorkerURL)
	if base == "" {
		return noopFetcher{}
	}
	return &httpWorkerFetcher{
		base:    strings.TrimRight(base, "/"),
		sources: cfg.WorkerSources,
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

type httpWorkerFetcher struct {
	base    string
	sources []string
	client  *http.Client
}

func (c *httpWorkerFetcher) Fetch(ctx context.Context, source string, scenarioKey string) (map[string]any, error) {
	source = strings.ToLower(strings.TrimSpace(source))
	scenarioKey = strings.TrimSpace(scenarioKey)
	if source == "" || scenarioKey == "" {
		return nil, nil
	}
	if len(c.sources) > 0 && !config.SourceEnabled(c.sources, source) {
		return nil, nil
	}
	joined, err := url.JoinPath(c.base, "v1", "sources", source)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(joined)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("scenario_key", scenarioKey)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("worker %s returned %s", source, resp.Status)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
