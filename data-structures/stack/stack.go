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
	"sync"
	"sync/atomic"
)

// Stack functions for manager datas in stack
type Stack interface {
	// Push a data into stack
	Push(v any)
	// PushMany many data into stack
	PushMany(vs ...any)
	// Pop last data
	Pop() (any, bool)
	// PopMany pop many of data
	PopMany(count int64) ([]any, bool)
	// PopAll pop all data
	PopAll() ([]any, bool)
	// Peek last data
	Peek() (any, bool)
	// Length get length of stack
	Length() int64
	// IsEmpty judge stack's length if 0
	IsEmpty() bool
}

type defaultStack struct {
	sync.Mutex
	length int64
	stack  []any
}

// New get stack functions manager
func New() Stack {
	return &defaultStack{}
}

func (p *defaultStack) Push(v any) {
	p.PushMany(v)
}

func (p *defaultStack) PushMany(vs ...any) {
	p.Lock()
	defer p.Unlock()

	lenVS := len(vs)

	prepend := make([]any, lenVS)
	copy(prepend, vs)

	p.stack = append(prepend, p.stack...)
	p.length += int64(lenVS)
}

func (p *defaultStack) Pop() (v any, exist bool) {
	if p.IsEmpty() {
		return
	}

	p.Lock()
	defer p.Unlock()

	v, p.stack, exist = p.stack[0], p.stack[1:], true
	p.length--

	return
}

func (p *defaultStack) PopMany(count int64) (vs []any, exist bool) {

	if p.IsEmpty() {
		return
	}

	p.Lock()
	defer p.Unlock()

	if count >= p.length {
		count = p.length
	}
	p.length -= count

	vs, p.stack, exist = p.stack[:count-1], p.stack[count:], true
	return
}

func (p *defaultStack) PopAll() (all []any, exist bool) {
	if p.IsEmpty() {
		return
	}
	p.Lock()
	defer p.Unlock()

	all, p.stack, exist = p.stack[:], nil, true
	p.length = 0
	return
}

func (p *defaultStack) Peek() (v any, exist bool) {
	if p.IsEmpty() {
		return
	}

	p.Lock()
	defer p.Unlock()

	return p.stack[0], true
}

func (p *defaultStack) Length() int64 {
	return atomic.LoadInt64(&p.length)
}

func (p *defaultStack) IsEmpty() bool {
	return p.Length() == 0
}
