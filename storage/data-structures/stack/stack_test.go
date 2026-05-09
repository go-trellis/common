/*
Copyright © 2016 Henry Huang <hhh@rutcode.com>

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

package stack

import (
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestNew(t *testing.T) {
	stack := New()
	testutils.Assert(t, stack != nil, "stack should not be nil")
	testutils.Assert(t, stack.IsEmpty(), "new stack should be empty")
	testutils.Equals(t, int64(0), stack.Length(), "new stack length should be 0")
}

func TestPush(t *testing.T) {
	stack := New()
	stack.Push(1)
	testutils.Assert(t, !stack.IsEmpty(), "stack should not be empty after push")
	testutils.Equals(t, int64(1), stack.Length(), "stack length should be 1")
}

func TestPushMany(t *testing.T) {
	stack := New()
	stack.PushMany(1, 2, 3)
	testutils.Assert(t, !stack.IsEmpty(), "stack should not be empty")
	testutils.Equals(t, int64(3), stack.Length(), "stack length should be 3")
}

func TestPop(t *testing.T) {
	stack := New()
	stack.Push(1)
	stack.Push(2)

	// PushMany adds new elements to the front, so stack is [2, 1]
	// Pop returns stack[0], which is 2 (last pushed)
	val, ok := stack.Pop()
	testutils.Assert(t, ok, "pop should succeed")
	testutils.Equals(t, 2, val, "first pop should return last pushed value")
	testutils.Equals(t, int64(1), stack.Length(), "stack length should be 1")

	val, ok = stack.Pop()
	testutils.Assert(t, ok, "pop should succeed")
	testutils.Equals(t, 1, val, "second pop should return first pushed value")
	testutils.Assert(t, stack.IsEmpty(), "stack should be empty after popping all")
}

func TestPop_Empty(t *testing.T) {
	stack := New()
	val, ok := stack.Pop()
	testutils.Assert(t, !ok, "pop should fail on empty stack")
	testutils.Assert(t, val == nil, "pop should return nil on empty stack")
}

func TestPopMany(t *testing.T) {
	stack := New()
	stack.PushMany(1, 2, 3, 4, 5)

	// PushMany adds to front, so stack is [5, 4, 3, 2, 1]
	// PopMany(3) returns stack[:3] = [5, 4, 3]
	vals, ok := stack.PopMany(3)
	testutils.Assert(t, ok, "PopMany should succeed")
	testutils.Assert(t, len(vals) == 3, "PopMany should return correct number of values")
	testutils.Equals(t, int64(2), stack.Length(), "stack length should be 2")
}

func TestPopMany_MoreThanLength(t *testing.T) {
	stack := New()
	stack.PushMany(1, 2)

	// PushMany adds to front, so stack is [2, 1]
	// PopMany(5) will cap to length=2, returns stack[:2] = [2, 1]
	vals, ok := stack.PopMany(5)
	testutils.Assert(t, ok, "PopMany should succeed")
	testutils.Assert(t, len(vals) == 2, "PopMany should return all available values")
	testutils.Assert(t, stack.IsEmpty(), "stack should be empty after popping all")
}

func TestPopMany_Empty(t *testing.T) {
	stack := New()
	vals, ok := stack.PopMany(3)
	testutils.Assert(t, !ok, "PopMany should fail on empty stack")
	testutils.Assert(t, vals == nil, "PopMany should return nil on empty stack")
}

func TestPopAll(t *testing.T) {
	stack := New()
	stack.PushMany(1, 2, 3)

	vals, ok := stack.PopAll()
	testutils.Assert(t, ok, "PopAll should succeed")
	testutils.Assert(t, len(vals) == 3, "PopAll should return all values")
	testutils.Assert(t, stack.IsEmpty(), "stack should be empty after PopAll")
	testutils.Equals(t, int64(0), stack.Length(), "stack length should be 0")
}

func TestPopAll_Empty(t *testing.T) {
	stack := New()
	vals, ok := stack.PopAll()
	testutils.Assert(t, !ok, "PopAll should fail on empty stack")
	testutils.Assert(t, vals == nil, "PopAll should return nil on empty stack")
}

func TestPeek(t *testing.T) {
	stack := New()
	stack.Push(1)
	stack.Push(2)

	// PushMany adds new elements to the front, so stack is [2, 1]
	// Peek returns stack[0], which is 2 (last pushed)
	val, ok := stack.Peek()
	testutils.Assert(t, ok, "Peek should succeed")
	testutils.Equals(t, 2, val, "Peek should return last pushed value")
	testutils.Equals(t, int64(2), stack.Length(), "Peek should not change stack length")
}

func TestPeek_Empty(t *testing.T) {
	stack := New()
	val, ok := stack.Peek()
	testutils.Assert(t, !ok, "Peek should fail on empty stack")
	testutils.Assert(t, val == nil, "Peek should return nil on empty stack")
}

func TestLength(t *testing.T) {
	stack := New()
	testutils.Equals(t, int64(0), stack.Length(), "initial length should be 0")

	stack.Push(1)
	testutils.Equals(t, int64(1), stack.Length(), "length should be 1")

	stack.PushMany(2, 3)
	testutils.Equals(t, int64(3), stack.Length(), "length should be 3")

	stack.Pop()
	testutils.Equals(t, int64(2), stack.Length(), "length should be 2")
}

func TestIsEmpty(t *testing.T) {
	stack := New()
	testutils.Assert(t, stack.IsEmpty(), "new stack should be empty")

	stack.Push(1)
	testutils.Assert(t, !stack.IsEmpty(), "stack should not be empty after push")

	stack.Pop()
	testutils.Assert(t, stack.IsEmpty(), "stack should be empty after pop")
}

func TestConcurrentOperations(t *testing.T) {
	stack := New()
	done := make(chan bool)

	// Concurrent pushes
	go func() {
		for i := 0; i < 100; i++ {
			stack.Push(i)
		}
		done <- true
	}()

	go func() {
		for i := 100; i < 200; i++ {
			stack.Push(i)
		}
		done <- true
	}()

	<-done
	<-done

	testutils.Assert(t, !stack.IsEmpty(), "stack should not be empty")
	testutils.Assert(t, stack.Length() >= 100, "stack should have at least 100 items")
}

func TestPushPopSequence(t *testing.T) {
	stack := New()

	// Push sequence: 1, 2, 3
	// PushMany adds to front, so stack becomes [3, 2, 1]
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)

	// Pop returns stack[0], which is the last pushed value
	val, _ := stack.Pop()
	testutils.Equals(t, 3, val, "first pop should return last pushed")

	val, _ = stack.Pop()
	testutils.Equals(t, 2, val, "second pop should return second to last pushed")

	val, _ = stack.Pop()
	testutils.Equals(t, 1, val, "third pop should return first pushed")
}

func TestPushManyAndPopMany(t *testing.T) {
	stack := New()
	stack.PushMany(1, 2, 3, 4, 5)

	// PushMany adds to front, so stack is [5, 4, 3, 2, 1]
	// PopMany(3) returns stack[:count] = stack[:3] = [5, 4, 3]
	// Then stack becomes stack[count:] = stack[3:] = [2, 1] (2 elements remaining)
	vals, ok := stack.PopMany(3)
	testutils.Assert(t, ok, "PopMany should succeed")
	testutils.Assert(t, len(vals) == 3, "should return 3 values")

	remaining, ok := stack.PopAll()
	testutils.Assert(t, ok, "PopAll should succeed")
	testutils.Assert(t, len(remaining) == 2, "should return remaining 2 values")
}

func TestPushPopDifferentTypes(t *testing.T) {
	stack := New()
	// PushMany adds to front, so stack is [true, 3.14, 42, "string"]
	stack.Push("string")
	stack.Push(42)
	stack.Push(3.14)
	stack.Push(true)

	testutils.Equals(t, int64(4), stack.Length(), "stack should have 4 items")

	val, _ := stack.Pop()
	testutils.Assert(t, val == true, "should pop bool first (last pushed)")

	val, _ = stack.Pop()
	testutils.Assert(t, val == 3.14, "should pop float second")

	val, _ = stack.Pop()
	testutils.Assert(t, val == 42, "should pop int third")

	val, _ = stack.Pop()
	testutils.Assert(t, val == "string", "should pop string last (first pushed)")
}
