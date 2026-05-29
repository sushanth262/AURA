package connectors_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sushanth262/AURA/aura-backend/internal/connectors"
)

func TestCircuitBreaker_OpensAfterFiveFailuresInWindow(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	br := connectors.DefaultBreaker()
	br.Now = func() time.Time { return now }

	fail := errors.New("fail")
	for i := 0; i < 5; i++ {
		now = now.Add(2 * time.Second)
		if err := br.Execute(func() error { return fail }); err != fail {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
	err := br.Execute(func() error { return nil })
	if !errors.Is(err, connectors.ErrCircuitOpen) {
		t.Fatalf("expected circuit open, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenThenClosesOnTwoSuccesses(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	br := connectors.DefaultBreaker()
	br.Now = func() time.Time { return now }

	fail := errors.New("fail")
	for i := 0; i < 5; i++ {
		now = now.Add(2 * time.Second)
		_ = br.Execute(func() error { return fail })
	}
	now = now.Add(61 * time.Second)
	if err := br.Execute(func() error { return nil }); err != nil {
		t.Fatalf("half-open probe: %v", err)
	}
	if err := br.Execute(func() error { return nil }); err != nil {
		t.Fatalf("second success: %v", err)
	}
	if br.State() != 0 { // stateClosed
		t.Fatalf("expected closed state, got %d", br.State())
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	br := connectors.DefaultBreaker()
	br.Now = func() time.Time { return now }
	fail := errors.New("fail")
	for i := 0; i < 5; i++ {
		now = now.Add(2 * time.Second)
		_ = br.Execute(func() error { return fail })
	}
	now = now.Add(61 * time.Second)
	_ = br.Execute(func() error { return fail })
	err := br.Execute(func() error { return nil })
	if !errors.Is(err, connectors.ErrCircuitOpen) {
		t.Fatalf("expected open after half-open failure, got %v", err)
	}
}
