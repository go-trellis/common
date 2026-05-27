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

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-trellis/common/utils/testutils"
)

func TestDo_Success(t *testing.T) {
	cfg := Config{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
	}

	count := 0
	err := Do(nil, cfg, func() error {
		count++
		return nil
	})

	testutils.Ok(t, err)
	testutils.Equals(t, 1, count, "should succeed on first try")
}

func TestDo_RetrySuccess(t *testing.T) {
	cfg := Config{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
	}

	count := 0
	err := Do(nil, cfg, func() error {
		count++
		if count < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	testutils.Ok(t, err)
	testutils.Equals(t, 3, count, "should retry 3 times")
}

func TestDo_MaxRetriesReached(t *testing.T) {
	cfg := Config{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
	}

	count := 0
	err := Do(nil, cfg, func() error {
		count++
		return errors.New("persistent error")
	})

	testutils.NotOk(t, err)
	testutils.Equals(t, 3, count, "should retry max retries + 1 times")
}

func TestDo_RetryableErrors(t *testing.T) {
	retriableErr := errors.New("retriable")
	nonRetriableErr := errors.New("non-retriable")

	cfg := Config{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		RetryableErrors: func(err error) bool {
			return err == retriableErr
		},
	}

	count := 0
	err := Do(nil, cfg, func() error {
		count++
		if count == 1 {
			return retriableErr
		}
		return nonRetriableErr
	})

	testutils.NotOk(t, err)
	testutils.Equals(t, nonRetriableErr, err, "should return non-retriable error")
	testutils.Equals(t, 2, count, "should stop on non-retriable error")
}

func TestDo_WithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := Config{
		MaxRetries:   10,
		InitialDelay: 100 * time.Millisecond,
	}

	count := 0
	err := Do(ctx, cfg, func() error {
		count++
		if count == 1 {
			cancel()
			return errors.New("error")
		}
		return nil
	})

	testutils.NotOk(t, err)
	testutils.Equals(t, context.Canceled, err, "should return context error")
}

func TestDoWithResult_Success(t *testing.T) {
	cfg := Config{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
	}

	count := 0
	result, err := DoWithResult(nil, cfg, func() (int, error) {
		count++
		return 42, nil
	})

	testutils.Ok(t, err)
	testutils.Equals(t, 42, result, "should return result")
	testutils.Equals(t, 1, count, "should succeed on first try")
}

func TestExponentialBackoff(t *testing.T) {
	backoff := &ExponentialBackoff{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  1 * time.Second,
		Factor:    2.0,
	}

	testutils.Equals(t, 100*time.Millisecond, backoff.Next(0), "first retry")
	testutils.Equals(t, 200*time.Millisecond, backoff.Next(1), "second retry")
	testutils.Equals(t, 400*time.Millisecond, backoff.Next(2), "third retry")
	testutils.Equals(t, 800*time.Millisecond, backoff.Next(3), "fourth retry")

	// Test max delay
	backoff.MaxDelay = 500 * time.Millisecond
	testutils.Equals(t, 500*time.Millisecond, backoff.Next(3), "should cap at max delay")
}

func TestLinearBackoff(t *testing.T) {
	backoff := &LinearBackoff{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  500 * time.Millisecond,
		Increment: 50 * time.Millisecond,
	}

	testutils.Equals(t, 100*time.Millisecond, backoff.Next(0), "first retry")
	testutils.Equals(t, 150*time.Millisecond, backoff.Next(1), "second retry")
	testutils.Equals(t, 200*time.Millisecond, backoff.Next(2), "third retry")
	testutils.Equals(t, 250*time.Millisecond, backoff.Next(3), "fourth retry")

	// Test max delay
	testutils.Equals(t, 500*time.Millisecond, backoff.Next(10), "should cap at max delay")
}

func TestFixedBackoff(t *testing.T) {
	backoff := &FixedBackoff{
		Delay: 100 * time.Millisecond,
	}

	for i := 0; i < 5; i++ {
		testutils.Equals(t, 100*time.Millisecond, backoff.Next(i), "should return fixed delay")
	}
}
