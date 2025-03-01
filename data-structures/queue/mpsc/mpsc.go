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

// This implementation is based on http://www.1024cores.net/home/lock-free-algorithms/queues/non-intrusive-mpsc-node-based-queue

import (
	"sync/atomic"
	"unsafe"
)

type node struct {
	next *node
	pos  uint64
	val  any
}

type Queue struct {
	head, tail *node
}

func New() *Queue {
	q := &Queue{}
	stub := &node{}
	q.head = stub
	q.tail = stub
	return q
}

// Push adds an item to the back of the queue. It must be called from a single, producer goroutine.
func (q *Queue) Push(x any) {
	n := &node{val: x}
	// current producer acquires head node
	prev := (*node)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&q.head)), unsafe.Pointer(n)))

	n.pos = prev.pos + 1
	// release node to consumer
	prev.next = n
}

// Pop removes an item from the front of the queue. It must be called from a single, consumer goroutine. If the queue is empty it returns nil.
func (q *Queue) Pop() any {
	tail := q.tail
	next := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)))) // acquire
	if next != nil {
		q.tail = next
		v := next.val
		next.val = nil
		return v
	}
	return nil
}

// Empty returns true if the queue is empty. It must be called from a single, consumer goroutine.
func (q *Queue) Empty() bool {
	tail := q.tail
	next := (*node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next))))
	return next == nil
}

// Length returns the number of items in the queue. It must be called from a single, consumer goroutine.
func (q *Queue) Length() uint64 {
	return atomic.LoadUint64(&q.head.pos) - atomic.LoadUint64(&q.tail.pos)
}
