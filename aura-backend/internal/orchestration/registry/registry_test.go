package registry

import (
	"testing"

	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

func TestDefaultRegistry_EnabledAgents(t *testing.T) {
	reg := DefaultRegistry()
	enabled := reg.EnabledAgents()
	if len(enabled) != 3 {
		t.Fatalf("want 3 enabled agents, got %d", len(enabled))
	}
	if enabled[0].Domain != orchestration.DomainTelemetry {
		t.Fatalf("first domain: got %s", enabled[0].Domain)
	}
}

func TestNew_SubsetEnabled(t *testing.T) {
	reg := New([]string{"telemetry"})
	enabled := reg.EnabledAgents()
	if len(enabled) != 1 {
		t.Fatalf("want 1, got %d", len(enabled))
	}
}

func TestValidateEnabledDomains(t *testing.T) {
	if err := ValidateEnabledDomains([]string{"telemetry", "unknown"}); err == nil {
		t.Fatal("expected error for unknown domain")
	}
	if err := ValidateEnabledDomains([]string{"telemetry", "supervisor"}); err != nil {
		t.Fatalf("supervisor should be ignored: %v", err)
	}
	if err := ValidateEnabledDomains([]string{"communications"}); err != nil {
		t.Fatalf("communications should be valid: %v", err)
	}
}

func TestNew_WithCommunications(t *testing.T) {
	reg := New([]string{"telemetry", "communications"})
	enabled := reg.EnabledAgents()
	if len(enabled) != 2 {
		t.Fatalf("want 2 enabled, got %d", len(enabled))
	}
	if enabled[1].Domain != orchestration.DomainCommunications {
		t.Fatalf("second domain: got %s", enabled[1].Domain)
	}
}
