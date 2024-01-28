/*
Copyright © 2022 Henry Huang <hhh@rutcode.com>

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

package pool

import (
	"time"

	"trellis.tech/common.v2/errcode"
)

type Factory func() (interface{}, error)
type Executor func(interface{}) error

type Option func(*Options)

type Options struct {
	initialCap  int
	maxCap      int
	maxIdle     int
	idleTimeout time.Duration

	factory Factory
	close   Executor
	ping    Executor
}

func (p *Options) check() error {

	if p.maxIdle <= 0 {
		p.maxIdle = p.maxCap
	}

	if !(p.initialCap <= p.maxIdle && p.maxCap >= p.maxIdle && p.initialCap >= 0) {
		return errcode.New("invalid settings: capacity")
	}
	if p.factory == nil {
		return errcode.New("invalid settings: factory function")
	}
	if p.close == nil {
		return errcode.New("invalid settings: close function")
	}

	return nil
}

func InitialCap(cap int) Option {
	return func(o *Options) {
		o.initialCap = cap
	}
}

func MaxCap(cap int) Option {
	return func(o *Options) {
		o.maxCap = cap
	}
}

func MaxIdle(idle int) Option {
	return func(o *Options) {
		o.maxIdle = idle
	}
}

func IdleTimeout(idle time.Duration) Option {
	return func(o *Options) {
		o.idleTimeout = idle
	}
}

func OptionFactory(f Factory) Option {
	return func(o *Options) {
		o.factory = f
	}
}

func OptionClose(f Executor) Option {
	return func(o *Options) {
		o.close = f
	}
}

func OptionPing(f Executor) Option {
	return func(o *Options) {
		o.ping = f
	}
}
