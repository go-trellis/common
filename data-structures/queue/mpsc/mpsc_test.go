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

package mpsc

import (
	"testing"

	"trellis.tech/trellis/common.v3/testutils"
)

func TestNew(t *testing.T) {
	q := New()
	testutils.Assert(t, q != nil, "queue should not be nil")
	testutils.Assert(t, q.Empty(), "new queue should be empty")
	testutils.Equals(t, uint64(0), q.Length(), "new queue length should be 0")
}

func TestPush(t *testing.T) {
	q := New()
	q.Push(1)
	testutils.Assert(t, !q.Empty(), "queue should not be empty after push")
	testutils.Equals(t, uint64(1), q.Length(), "queue length should be 1")
}

func TestPop(t *testing.T) {
	q := New()
	q.Push(1)
	q.Push(2)

	val := q.Pop()
	testutils.Assert(t, val != nil, "Pop should return a value")
	testutils.Equals(t, 1, val, "first pop should return first pushed value")
	testutils.Equals(t, uint64(1), q.Length(), "queue length should be 1")

	val = q.Pop()
	testutils.Assert(t, val != nil, "Pop should return a value")
	testutils.Equals(t, 2, val, "second pop should return second pushed value")
	testutils.Assert(t, q.Empty(), "queue should be empty after popping all")
}

func TestPop_Empty(t *testing.T) {
	q := New()
	val := q.Pop()
	testutils.Assert(t, val == nil, "Pop should return nil on empty queue")
}

func TestEmpty(t *testing.T) {
	q := New()
	testutils.Assert(t, q.Empty(), "new queue should be empty")

	q.Push(1)
	testutils.Assert(t, !q.Empty(), "queue should not be empty after push")

	q.Pop()
	testutils.Assert(t, q.Empty(), "queue should be empty after pop")
}

func TestLength(t *testing.T) {
	q := New()
	testutils.Equals(t, uint64(0), q.Length(), "initial length should be 0")

	q.Push(1)
	testutils.Equals(t, uint64(1), q.Length(), "length should be 1")

	q.Push(2)
	testutils.Equals(t, uint64(2), q.Length(), "length should be 2")

	q.Pop()
	testutils.Equals(t, uint64(1), q.Length(), "length should be 1 after pop")
}

func TestPushPopSequence(t *testing.T) {
	q := New()

	for i := 0; i < 10; i++ {
		q.Push(i)
	}

	testutils.Equals(t, uint64(10), q.Length(), "queue should have 10 items")

	for i := 0; i < 10; i++ {
		val := q.Pop()
		testutils.Equals(t, i, val, "pop should return values in order")
	}

	testutils.Assert(t, q.Empty(), "queue should be empty after popping all")
}

func TestPushPopDifferentTypes(t *testing.T) {
	q := New()
	q.Push("string")
	q.Push(42)
	q.Push(3.14)
	q.Push(true)

	testutils.Equals(t, uint64(4), q.Length(), "queue should have 4 items")

	val := q.Pop()
	testutils.Assert(t, val == "string", "should pop string first")

	val = q.Pop()
	testutils.Assert(t, val == 42, "should pop int second")

	val = q.Pop()
	testutils.Assert(t, val == 3.14, "should pop float third")

	val = q.Pop()
	testutils.Assert(t, val == true, "should pop bool last")
}

func TestConcurrentPush(t *testing.T) {
	q := New()
	done := make(chan bool)

	// Concurrent pushes from multiple goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				q.Push(id*10 + j)
			}
			done <- true
		}(i)
	}

	// Wait for all pushes
	for i := 0; i < 10; i++ {
		<-done
	}

	testutils.Assert(t, !q.Empty(), "queue should not be empty")
	testutils.Assert(t, q.Length() >= 10, "queue should have at least 10 items")
}

func TestManyPushPop(t *testing.T) {
	q := New()

	// Push many items
	for i := 0; i < 1000; i++ {
		q.Push(i)
	}

	testutils.Equals(t, uint64(1000), q.Length(), "queue should have 1000 items")

	// Pop all items
	count := 0
	for !q.Empty() {
		val := q.Pop()
		testutils.Assert(t, val != nil, "Pop should return a value")
		count++
	}

	testutils.Equals(t, 1000, count, "should pop all 1000 items")
	testutils.Assert(t, q.Empty(), "queue should be empty after popping all")
}

func TestLengthAfterOperations(t *testing.T) {
	q := New()

	testutils.Equals(t, uint64(0), q.Length(), "initial length should be 0")

	q.Push(1)
	testutils.Equals(t, uint64(1), q.Length(), "length after first push")

	q.Push(2)
	testutils.Equals(t, uint64(2), q.Length(), "length after second push")

	q.Pop()
	testutils.Equals(t, uint64(1), q.Length(), "length after first pop")

	q.Pop()
	testutils.Equals(t, uint64(0), q.Length(), "length after second pop")
}

func TestEmptyAfterPop(t *testing.T) {
	q := New()
	q.Push(1)

	testutils.Assert(t, !q.Empty(), "queue should not be empty")

	q.Pop()
	testutils.Assert(t, q.Empty(), "queue should be empty after pop")
}
