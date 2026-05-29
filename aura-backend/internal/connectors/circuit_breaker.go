package connectors

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the breaker is open and not yet probing.
var ErrCircuitOpen = errors.New("connector circuit open")

// CircuitBreaker implements PRODUCTION_SPEC §2.3 thresholds.
type CircuitBreaker struct {
	mu sync.Mutex

	state              int
	failures           []time.Time
	consecutiveSuccess int
	openedAt           time.Time

	FailureThreshold int
	FailureWindow    time.Duration
	HalfOpenAfter    time.Duration
	SuccessToClose   int

	Now func() time.Time
}

const (
	stateClosed = iota
	stateOpen
	stateHalfOpen
)

// DefaultBreaker returns a breaker with production default thresholds.
func DefaultBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		FailureThreshold: 5,
		FailureWindow:    30 * time.Second,
		HalfOpenAfter:    60 * time.Second,
		SuccessToClose:   2,
		Now:              func() time.Time { return time.Now().UTC() },
	}
}

// Execute runs fn through the breaker.
func (b *CircuitBreaker) Execute(fn func() error) error {
	if err := b.beforeCall(); err != nil {
		return err
	}
	err := fn()
	b.afterCall(err)
	return err
}

func (b *CircuitBreaker) beforeCall() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()

	switch b.state {
	case stateOpen:
		if now.Sub(b.openedAt) >= b.halfOpenAfter() {
			b.state = stateHalfOpen
			b.consecutiveSuccess = 0
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

func (b *CircuitBreaker) afterCall(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.now()

	if err == nil {
		switch b.state {
		case stateHalfOpen:
			b.consecutiveSuccess++
			if b.consecutiveSuccess >= b.successToClose() {
				b.state = stateClosed
				b.failures = nil
				b.consecutiveSuccess = 0
			}
		case stateClosed:
			b.failures = nil
		}
		return
	}

	b.consecutiveSuccess = 0
	switch b.state {
	case stateHalfOpen:
		b.state = stateOpen
		b.openedAt = now
	case stateClosed:
		b.failures = append(b.failures, now)
		b.pruneFailures(now)
		if len(b.failures) >= b.failureThreshold() {
			b.state = stateOpen
			b.openedAt = now
		}
	}
}

func (b *CircuitBreaker) pruneFailures(now time.Time) {
	cutoff := now.Add(-b.failureWindow())
	kept := b.failures[:0]
	for _, t := range b.failures {
		if !t.Before(cutoff) {
			kept = append(kept, t)
		}
	}
	b.failures = kept
}

func (b *CircuitBreaker) State() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *CircuitBreaker) now() time.Time {
	if b.Now != nil {
		return b.Now()
	}
	return time.Now().UTC()
}

func (b *CircuitBreaker) failureThreshold() int {
	if b.FailureThreshold > 0 {
		return b.FailureThreshold
	}
	return 5
}

func (b *CircuitBreaker) failureWindow() time.Duration {
	if b.FailureWindow > 0 {
		return b.FailureWindow
	}
	return 30 * time.Second
}

func (b *CircuitBreaker) halfOpenAfter() time.Duration {
	if b.HalfOpenAfter > 0 {
		return b.HalfOpenAfter
	}
	return 60 * time.Second
}

func (b *CircuitBreaker) successToClose() int {
	if b.SuccessToClose > 0 {
		return b.SuccessToClose
	}
	return 2
}
