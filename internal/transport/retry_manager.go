package transport

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// RetryManager implements exponential backoff with jitter and a simple circuit breaker.
type RetryManager struct {
	MaxRetries       int
	BaseBackoff      time.Duration
	MaxBackoff       time.Duration
	BackoffMultiplier float64
	JitterFactor     float64

	mu      sync.Mutex
	failures map[string]int
	state    map[string]CircuitState
}

// NewRetryManager creates a new RetryManager with sane defaults.
func NewRetryManager() *RetryManager {
	return &RetryManager{
		MaxRetries:       5,
		BaseBackoff:      100 * time.Millisecond,
		MaxBackoff:       30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterFactor:     0.1,
		failures:         make(map[string]int),
		state:            make(map[string]CircuitState),
	}
}

// ShouldRetry returns whether another attempt should be made.
func (r *RetryManager) ShouldRetry(attempt int, err error) bool {
	if attempt >= r.MaxRetries {
		return false
	}
	return true
}

// NextBackoff calculates the next backoff duration given the attempt count and RTT.
func (r *RetryManager) NextBackoff(attempt int, rtt time.Duration) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	backoff := float64(r.BaseBackoff) * math.Pow(r.BackoffMultiplier, float64(attempt-1))
	if rtt > 0 {
		backoff = math.Max(backoff, float64(rtt))
	}
	if backoff > float64(r.MaxBackoff) {
		backoff = float64(r.MaxBackoff)
	}
	// jitter
	jitter := backoff * r.JitterFactor * (rand.Float64()*2 - 1) // +/- jitterFactor
	backoff += jitter
	if backoff < float64(r.BaseBackoff) {
		backoff = float64(r.BaseBackoff)
	}
	return time.Duration(backoff)
}

// RecordSuccess resets failure count and closes circuit for identifier.
func (r *RetryManager) RecordSuccess(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.failures, id)
	r.state[id] = CircuitClosed
}

// RecordFailure increments failure count and may open circuit.
func (r *RetryManager) RecordFailure(id string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failures[id]++
	if r.failures[id] > r.MaxRetries {
		r.state[id] = CircuitOpen
	}
}

// GetCircuitState returns current circuit state for identifier.
func (r *RetryManager) GetCircuitState(id string) CircuitState {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.state[id]; ok {
		return s
	}
	return CircuitClosed
}


