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
	"testing"
	"time"
)

func TestCircuitBreaker_State(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	})

	if cb.State() != StateClosed {
		t.Errorf("expected state Closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	})

	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("expected state Closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	})

	testErr := errors.New("test error")
	err := cb.Execute(context.Background(), func() error {
		return testErr
	})

	if err != testErr {
		t.Errorf("expected test error, got %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("expected state Closed after 1 failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_Open(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	})

	// Fail once - circuit should still be closed
	cb.Execute(context.Background(), func() error {
		return errors.New("error 1")
	})

	if cb.State() != StateClosed {
		t.Errorf("expected state Closed after 1 failure, got %v", cb.State())
	}

	// Fail again - should open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("error 2")
	})

	if cb.State() != StateOpen {
		t.Errorf("expected state Open, got %v", cb.State())
	}

	// Try to execute - should fail immediately
	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	if err != ErrCircuitBreakerOpen {
		t.Errorf("expected ErrCircuitBreakerOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 1
		},
	})

	// Open the circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("error")
	})

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	if cb.State() != StateHalfOpen {
		t.Errorf("expected state HalfOpen, got %v", cb.State())
	}
}

func TestCircuitBreaker_Counts(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return false
		},
	})

	// Execute successful request
	cb.Execute(context.Background(), func() error {
		return nil
	})

	counts := cb.Counts()
	if counts.Requests != 1 {
		t.Errorf("expected 1 request, got %d", counts.Requests)
	}
	if counts.TotalSuccesses != 1 {
		t.Errorf("expected 1 success, got %d", counts.TotalSuccesses)
	}
}
