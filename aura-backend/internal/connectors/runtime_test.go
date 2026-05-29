package connectors_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/connectors"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

type failingDriver struct{ id string; fail bool }

func (d failingDriver) ID() string { return d.id }
func (d failingDriver) Invoke(context.Context, orchestration.ConnectorCall) (map[string]any, error) {
	if d.fail {
		return nil, errors.New("driver failed")
	}
	return map[string]any{"ok": true}, nil
}

func TestRuntime_InvokeFixtureGrafana(t *testing.T) {
	rt := connectors.NewBuiltinRuntime(config.Config{EnabledSources: []string{"grafana"}})
	out, err := rt.Invoke(context.Background(), orchestration.ConnectorCall{
		ConnectorID: "grafana",
		ScenarioKey: "inc2847_api_gateway",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out["mock"] != true {
		t.Fatalf("expected fixture mock, got %v", out["mock"])
	}
}

func TestRuntime_CircuitOpenBlocksInvoke(t *testing.T) {
	rt := connectors.NewRuntime()
	rt.RegisterConnector(failingDriver{id: "grafana", fail: true})

	for i := 0; i < 5; i++ {
		_, _ = rt.Invoke(context.Background(), orchestration.ConnectorCall{ConnectorID: "grafana", ScenarioKey: "x"})
	}
	_, err := rt.Invoke(context.Background(), orchestration.ConnectorCall{ConnectorID: "grafana", ScenarioKey: "x"})
	if !errors.Is(err, connectors.ErrCircuitOpen) {
		t.Fatalf("expected circuit open, got %v", err)
	}
}

func TestRuntime_IdempotentRetrySucceeds(t *testing.T) {
	rt := connectors.NewRuntime()
	attempts := 0
	rt.RegisterConnector(driverFunc{
		id: "github",
		fn: func(context.Context, orchestration.ConnectorCall) (map[string]any, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("transient")
			}
			return map[string]any{"ok": true}, nil
		},
	})

	out, err := rt.Invoke(context.Background(), orchestration.ConnectorCall{ConnectorID: "github", ScenarioKey: "k"})
	if err != nil {
		t.Fatal(err)
	}
	if out["ok"] != true || attempts != 2 {
		t.Fatalf("attempts=%d out=%v", attempts, out)
	}
}

type driverFunc struct {
	id string
	fn func(context.Context, orchestration.ConnectorCall) (map[string]any, error)
}

func (d driverFunc) ID() string { return d.id }
func (d driverFunc) Invoke(ctx context.Context, call orchestration.ConnectorCall) (map[string]any, error) {
	return d.fn(ctx, call)
}
