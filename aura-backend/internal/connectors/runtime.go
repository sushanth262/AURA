package connectors

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/fixturesdata"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
	"gopkg.in/yaml.v3"
)

// Driver invokes a single connector backend.
type Driver interface {
	ID() string
	Invoke(ctx context.Context, call orchestration.ConnectorCall) (map[string]any, error)
}

// Runtime routes connector calls through per-connector circuit breakers.
type Runtime struct {
	mu       sync.RWMutex
	drivers  map[string]Driver
	breakers map[string]*CircuitBreaker
}

// NewRuntime builds an empty runtime; register drivers before Invoke.
func NewRuntime() *Runtime {
	return &Runtime{
		drivers:  make(map[string]Driver),
		breakers: make(map[string]*CircuitBreaker),
	}
}

// RegisterConnector adds a driver and dedicated circuit breaker.
func (r *Runtime) RegisterConnector(d Driver) {
	if d == nil {
		return
	}
	id := strings.ToLower(strings.TrimSpace(d.ID()))
	if id == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[id] = d
	if _, ok := r.breakers[id]; !ok {
		r.breakers[id] = DefaultBreaker()
	}
}

// Invoke executes a connector call with breaker protection and one idempotent retry.
func (r *Runtime) Invoke(ctx context.Context, call orchestration.ConnectorCall) (map[string]any, error) {
	id := strings.ToLower(strings.TrimSpace(call.ConnectorID))
	if id == "" {
		return nil, fmt.Errorf("connector id required")
	}
	r.mu.RLock()
	d, ok := r.drivers[id]
	br := r.breakers[id]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown connector %q", id)
	}
	if br == nil {
		br = DefaultBreaker()
	}

	var out map[string]any
	err := br.Execute(func() error {
		var invokeErr error
		out, invokeErr = invokeIdempotent(ctx, d, call)
		return invokeErr
	})
	return out, err
}

func invokeIdempotent(ctx context.Context, d Driver, call orchestration.ConnectorCall) (map[string]any, error) {
	out, err := d.Invoke(ctx, call)
	if err == nil {
		return out, nil
	}
	// One retry for idempotent read-style connector calls.
	out2, err2 := d.Invoke(ctx, call)
	if err2 == nil {
		return out2, nil
	}
	return nil, err
}

// FixtureLoader loads scenario YAML and extracts per-source mocks.
type FixtureLoader struct {
	LoadYAML    func(baseName string) (map[string]any, error)
	ExtractMock func(root map[string]any, source string) any
}

// DefaultFixtureLoader reads embedded fixturesdata YAML.
func DefaultFixtureLoader() FixtureLoader {
	return FixtureLoader{
		LoadYAML:    loadScenarioYAML,
		ExtractMock: extractSourceMock,
	}
}

func loadScenarioYAML(baseName string) (map[string]any, error) {
	b, err := fixturesdata.FS.ReadFile(baseName + ".yaml")
	if err != nil {
		return nil, fmt.Errorf("fixture %q: %w", baseName, err)
	}
	body := bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	var root map[string]any
	if err := yaml.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	return root, nil
}

func extractSourceMock(root map[string]any, source string) any {
	sm, _ := root["source_mocks"].(map[string]any)
	if sm == nil {
		return nil
	}
	return sm[source]
}

// FixtureDriver serves YAML fixture payloads for a connector id.
type FixtureDriver struct {
	connectorID string
	loader      FixtureLoader
}

func NewFixtureDriver(connectorID string, loader FixtureLoader) *FixtureDriver {
	return &FixtureDriver{connectorID: strings.ToLower(connectorID), loader: loader}
}

func (d *FixtureDriver) ID() string { return d.connectorID }

func (d *FixtureDriver) Invoke(_ context.Context, call orchestration.ConnectorCall) (map[string]any, error) {
	key := strings.TrimSpace(call.ScenarioKey)
	if key == "" {
		return nil, nil
	}
	root, err := d.loader.LoadYAML(key)
	if err != nil {
		return nil, err
	}
	payload := d.loader.ExtractMock(root, d.connectorID)
	if payload == nil {
		return nil, nil
	}
	return map[string]any{
		"scenario_key": key,
		"source":       d.connectorID,
		"mock":         true,
		"payload":      payload,
	}, nil
}

// NewBuiltinRuntime registers all built-in connector drivers for aura-worker.
func NewBuiltinRuntime(cfg config.Config) *Runtime {
	rt := NewRuntime()
	loader := DefaultFixtureLoader()
	enabled := cfg.EnabledSources

	register := func(d Driver) {
		id := d.ID()
		if len(enabled) > 0 && !config.SourceEnabled(enabled, id) {
			return
		}
		rt.RegisterConnector(d)
	}

	register(NewGrafanaDriver(cfg, loader))
	for _, factory := range []func(FixtureLoader) Driver{
		NewGitHubDriver, NewJiraDriver, NewSlackDriver, NewTeamsDriver, NewEmailDriver,
	} {
		register(factory(loader))
	}
	return rt
}

// Ensure Runtime satisfies orchestration.ConnectorRuntime.
var _ orchestration.ConnectorRuntime = (*Runtime)(nil)
