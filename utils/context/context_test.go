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

package context

import (
	"testing"
	"time"

	"trellis.tech/trellis/common.v3/utils/testutils"
)

func TestWithTimeout(t *testing.T) {
	ctx, cancel := WithTimeout(nil, 100*time.Millisecond)
	defer cancel()

	testutils.Assert(t, ctx != nil, "context should not be nil")

	select {
	case <-ctx.Done():
		t.Fatal("context should not be done immediately")
	case <-time.After(50 * time.Millisecond):
		// OK
	}

	select {
	case <-ctx.Done():
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should be done after timeout")
	}
}

func TestWithValue(t *testing.T) {
	type keyType string
	key := keyType("test")
	value := "test-value"

	ctx := WithValue(nil, key, value)
	retrieved, ok := Value[keyType, string](ctx, key)
	testutils.Assert(t, ok, "should retrieve value")
	testutils.Equals(t, value, retrieved, "should match value")
}

func TestValueOrDefault(t *testing.T) {
	type keyType string
	key := keyType("test")

	ctx := Background()
	defaultVal := "default"

	result := ValueOrDefault(ctx, key, defaultVal)
	testutils.Equals(t, defaultVal, result, "should return default")

	ctx = WithValue(ctx, key, "actual")
	result = ValueOrDefault(ctx, key, defaultVal)
	testutils.Equals(t, "actual", result, "should return actual value")
}

func TestIsDone(t *testing.T) {
	ctx := Background()
	testutils.Assert(t, !IsDone(ctx), "should not be done")

	ctx, cancel := WithCancel(ctx)
	cancel()
	testutils.Assert(t, IsDone(ctx), "should be done after cancel")
}

func TestErr(t *testing.T) {
	ctx := Background()
	testutils.Ok(t, Err(ctx))

	ctx, cancel := WithCancel(ctx)
	cancel()
	err := Err(ctx)
	testutils.NotOk(t, err, "should return error after cancel")
}

func TestMerge(t *testing.T) {
	ctx1, cancel1 := WithCancel(Background())
	ctx2, cancel2 := WithCancel(Background())

	merged := Merge(ctx1, ctx2)
	testutils.Assert(t, merged != nil, "merged context should not be nil")

	cancel1()
	select {
	case <-merged.Done():
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("merged context should be done when one is cancelled")
	}

	cancel2()
}

func TestWithValues(t *testing.T) {
	type key1Type string
	type key2Type string

	values := map[interface{}]interface{}{
		key1Type("key1"): "value1",
		key2Type("key2"): "value2",
	}

	ctx := WithValues(nil, values)

	val1, ok1 := Value[key1Type, string](ctx, key1Type("key1"))
	testutils.Assert(t, ok1, "should retrieve value1")
	testutils.Equals(t, "value1", val1, "should match")

	val2, ok2 := Value[key2Type, string](ctx, key2Type("key2"))
	testutils.Assert(t, ok2, "should retrieve value2")
	testutils.Equals(t, "value2", val2, "should match")
}
