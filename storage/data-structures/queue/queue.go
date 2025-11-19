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

package queue

import (
	"sync"
	"sync/atomic"
)

// Queue functions for manager datas in queue
type Queue interface {
	// Push a data into queue
	Push(v any)
	// PushMany many data into queue
	PushMany(vs ...any)
	// Pop first data
	Pop() (any, bool)
	// PopMany pop many of data
	PopMany(count int64) ([]any, bool)
	// PopAll pop all data
	PopAll() ([]any, bool)
	// Front peek first data
	Front() (any, bool)
	// End peek end data
	End() (any, bool)
	// Length get length of queue
	Length() int64
	// IsEmpty judge queue's lenght if 0
	IsEmpty() bool
}

type defaultQueue struct {
	sync.Mutex
	length int64
	queue  []any
}

// New get queue functions manager
func New() Queue {
	return &defaultQueue{}
}

func (p *defaultQueue) Push(v any) {
	p.PushMany(v)
}

func (p *defaultQueue) PushMany(vs ...any) {
	p.Lock()
	defer p.Unlock()

	p.queue = append(p.queue, vs...)
	p.length += int64(len(vs))
}

func (p *defaultQueue) Pop() (v any, exist bool) {
	if p.IsEmpty() {
		return
	}

	p.Lock()
	defer p.Unlock()

	v, p.queue, exist = p.queue[0], p.queue[1:], true
	p.length--

	return
}

func (p *defaultQueue) PopMany(count int64) (vs []any, exist bool) {
	if count < 1 {
		return nil, false
	}
	if p.IsEmpty() {
		return
	}

	p.Lock()
	defer p.Unlock()

	if count >= p.length {
		count = p.length
	}
	p.length -= count

	vs, p.queue, exist = p.queue[:count], p.queue[count:], true
	return
}

func (p *defaultQueue) PopAll() (all []any, exist bool) {
	if p.IsEmpty() {
		return
	}
	p.Lock()
	defer p.Unlock()

	all, p.queue, exist = p.queue[:], nil, true
	p.length = 0
	return
}

func (p *defaultQueue) Front() (any, bool) {
	if p.IsEmpty() {
		return nil, false
	}
	p.Lock()
	defer p.Unlock()
	return p.queue[0], true
}

func (p *defaultQueue) End() (any, bool) {
	if p.IsEmpty() {
		return nil, false
	}
	p.Lock()
	defer p.Unlock()
	return p.queue[p.length-1], true
}

func (p *defaultQueue) Length() int64 {
	return atomic.LoadInt64(&p.length)
}

func (p *defaultQueue) IsEmpty() bool {
	return p.Length() == 0
}
