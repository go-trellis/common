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
	"trellis.tech/common.v2/errcode"
)

var (
	ErrPoolClosed     = errcode.New("pool is closed")
	ErrNilFactory     = errcode.New("nil factory")
	ErrOpenedMaxConns = errcode.New("opened the maximum conns")
	ErrNilCloseFunc   = errcode.New("nil close function")
	ErrNilConnection  = errcode.New("nil connection")
	ErrGetConnTimeout = errcode.New("timeout to get connection")
)

// Pool interface describes a pool implementation. A pool should have maximum
// capacity. An ideal pool is threadsafe and easy to use.
type Pool interface {
	Get() (interface{}, error)

	Put(interface{}) error

	Close(interface{}) error

	Release()

	Len() int
}

func NewPool(opts ...Option) (Pool, error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	p := &channelPool{
		options: options,
	}
	if err := p.init(); err != nil {
		return nil, err
	}

	return p, nil
}
