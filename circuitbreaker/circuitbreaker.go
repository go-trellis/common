/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrCircuitBreakerOpen is returned when the circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	// ErrCircuitBreakerHalfOpen is returned when the circuit breaker is half-open
	ErrCircuitBreakerHalfOpen = errors.New("circuit breaker is half-open")
)

// State represents the state of a circuit breaker
type State int32

const (
	// StateClosed means the circuit breaker is closed (normal operation)
	StateClosed State = iota
	// StateOpen means the circuit breaker is open (failing fast)
	StateOpen
	// StateHalfOpen means the circuit breaker is half-open (testing)
	StateHalfOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Config contains configuration for a circuit breaker
type Config struct {
	// Name is the name of the circuit breaker
	Name string

	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests uint32

	// Interval is the time period for counting errors
	Interval time.Duration

	// Timeout is the timeout duration before attempting to close the circuit breaker
	Timeout time.Duration

	// ReadyToTrip is a function that determines if the circuit breaker should trip
	// It receives the number of requests and errors, and returns true if the circuit should open
	ReadyToTrip func(counts Counts) bool

	// OnStateChange is called when the circuit breaker state changes
	OnStateChange func(name string, from State, to State)
}

// Counts contains statistics for the circuit breaker
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// DefaultReadyToTrip returns a default function that trips when failures exceed 50% and requests >= 10
func DefaultReadyToTrip() func(Counts) bool {
	return func(counts Counts) bool {
		return counts.ConsecutiveFailures >= 5
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(Counts) bool
	onStateChange func(string, State, State)

	mu                sync.Mutex
	state             State
	generation        uint64
	expiry            time.Time
	counts            Counts
	lastFailureTime   time.Time
	lastFailureReason error
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	if cfg.Interval <= 0 {
		cfg.Interval = time.Second * 60
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = time.Second * 60
	}
	if cfg.MaxRequests <= 0 {
		cfg.MaxRequests = 1
	}
	if cfg.ReadyToTrip == nil {
		cfg.ReadyToTrip = DefaultReadyToTrip()
	}

	cb := &CircuitBreaker{
		name:          cfg.Name,
		maxRequests:   cfg.MaxRequests,
		interval:      cfg.Interval,
		timeout:       cfg.Timeout,
		readyToTrip:   cfg.ReadyToTrip,
		onStateChange: cfg.OnStateChange,
		state:         StateClosed,
		generation:    0,
	}

	cb.toNewGeneration(time.Now())

	return cb
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, true, errors.New("panic occurred"))
			panic(e)
		}
	}()

	err = fn()
	cb.afterRequest(generation, err != nil, err)
	return err
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns the current counts
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return Counts{
		Requests:             atomic.LoadUint32(&cb.counts.Requests),
		TotalSuccesses:       atomic.LoadUint32(&cb.counts.TotalSuccesses),
		TotalFailures:        atomic.LoadUint32(&cb.counts.TotalFailures),
		ConsecutiveSuccesses: atomic.LoadUint32(&cb.counts.ConsecutiveSuccesses),
		ConsecutiveFailures:  atomic.LoadUint32(&cb.counts.ConsecutiveFailures),
	}
}

// Name returns the name of the circuit breaker
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitBreakerOpen
	} else if state == StateHalfOpen {
		if cb.counts.Requests >= cb.maxRequests {
			return generation, ErrCircuitBreakerHalfOpen
		}
		cb.counts.Requests++
	}

	return generation, nil
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(beforeGeneration uint64, failed bool, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if generation != beforeGeneration {
		return
	}

	if failed {
		cb.onFailure(state, now, err)
	} else {
		cb.onSuccess(state, now)
	}
}

// currentState returns the current state and generation
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && now.After(cb.expiry) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if now.After(cb.expiry) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.onSuccess()
	case StateHalfOpen:
		cb.counts.onSuccess()
		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure(state State, now time.Time, err error) {
	switch state {
	case StateClosed:
		cb.counts.onFailure()
		// Read counts atomically for readyToTrip check
		counts := Counts{
			Requests:             atomic.LoadUint32(&cb.counts.Requests),
			TotalSuccesses:       atomic.LoadUint32(&cb.counts.TotalSuccesses),
			TotalFailures:        atomic.LoadUint32(&cb.counts.TotalFailures),
			ConsecutiveSuccesses: atomic.LoadUint32(&cb.counts.ConsecutiveSuccesses),
			ConsecutiveFailures:  atomic.LoadUint32(&cb.counts.ConsecutiveFailures),
		}
		if cb.readyToTrip(counts) {
			cb.setState(StateOpen, now)
		}
		// Don't reset counts on failure - let interval handle it
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
	cb.lastFailureTime = now
	cb.lastFailureReason = err
}

// setState changes the state of the circuit breaker
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prevState := cb.state
	cb.state = state
	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prevState, state)
	}
}

// toNewGeneration resets the counts and sets a new generation
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default:
		cb.expiry = zero
	}
}

// onSuccess updates counts for a successful request
func (c *Counts) onSuccess() {
	atomic.AddUint32(&c.Requests, 1)
	atomic.AddUint32(&c.TotalSuccesses, 1)
	atomic.AddUint32(&c.ConsecutiveSuccesses, 1)
	atomic.StoreUint32(&c.ConsecutiveFailures, 0)
}

// onFailure updates counts for a failed request
func (c *Counts) onFailure() {
	atomic.AddUint32(&c.Requests, 1)
	atomic.AddUint32(&c.TotalFailures, 1)
	atomic.AddUint32(&c.ConsecutiveFailures, 1)
	atomic.StoreUint32(&c.ConsecutiveSuccesses, 0)
}
