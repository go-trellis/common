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
	"sync"
	"time"

	"trellis.tech/trellis/common.v2/errcode"
)

type channelPool struct {
	mu      sync.RWMutex
	options *Options

	conns    chan *idleConn
	waitings []chan waitConn

	maxActive int
	openings  int
}

type idleConn struct {
	conn any
	t    time.Time
}

type waitConn struct {
	idleConn *idleConn
}

func (p *channelPool) init() error {

	if err := p.options.check(); err != nil {
		return err
	}

	p.conns = make(chan *idleConn, p.options.maxIdle)
	p.maxActive = p.options.maxCap
	p.openings = p.options.initialCap

	for i := 0; i < p.options.initialCap; i++ {
		c, err := p.options.factory()
		if err != nil {
			p.Release()
			return errcode.Newf("construct instance failed: %s", err.Error())
		}
		p.conns <- &idleConn{conn: c, t: time.Now()}
	}
	return nil
}

func (p *channelPool) getConns() chan *idleConn {
	p.mu.Lock()
	conns := p.conns
	p.mu.Unlock()
	return conns
}

func (p *channelPool) Get() (any, error) {
	conns := p.getConns()
	if conns == nil {
		return nil, ErrPoolClosed
	}

	for {
		select {
		case c := <-conns:
			if c == nil {
				return nil, ErrPoolClosed
			}
			if p.options.idleTimeout > 0 && c.t.Add(p.options.idleTimeout).Before(time.Now()) {
				p.Close(c.conn)
				continue
			}
			if p.options.ping != nil {
				if err := p.options.ping(c.conn); err != nil {
					p.Close(c.conn)
					continue
				}
			}
			return c.conn, nil
		default:
			p.mu.Lock()
			if p.openings >= p.maxActive {
				wait := make(chan waitConn, 1)
				p.waitings = append(p.waitings, wait)
				p.mu.Unlock()
				c, ok := <-wait
				if !ok {
					return nil, ErrOpenedMaxConns
				}
				if p.options.idleTimeout > 0 && c.idleConn.t.Add(p.options.idleTimeout).Before(time.Now()) {
					p.Close(c.idleConn.conn)
					continue
				}
				return c.idleConn.conn, nil
			}

			if p.options.factory == nil {
				p.mu.Unlock()
				return nil, ErrNilFactory
			}
			conn, err := p.options.factory()
			if err != nil {
				p.mu.Unlock()
				return nil, err
			}
			p.openings++
			p.mu.Unlock()
			return conn, nil
		}
	}
}

func (p *channelPool) Put(c any) error {
	if c == nil {
		return ErrNilConnection
	}
	p.mu.Lock()
	if p.conns == nil {
		p.mu.Unlock()
		return p.Close(c)
	}
	if l := len(p.waitings); l > 0 {
		wc := p.waitings[0]
		copy(p.waitings, p.waitings[1:])
		p.waitings = p.waitings[:l-1]
		wc <- waitConn{
			idleConn: &idleConn{
				conn: c,
				t:    time.Now(),
			},
		}

		p.mu.Unlock()
		return nil
	}

	select {
	case p.conns <- &idleConn{conn: c, t: time.Now()}:
		p.mu.Unlock()
		return nil
	default:
		p.mu.Unlock()
		return p.Close(c)
	}
}

func (p *channelPool) Close(c any) error {
	if c == nil {
		return ErrNilConnection
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.options.close == nil {
		return nil
	}
	p.openings--
	return p.options.close(c)
}

func (p *channelPool) Ping(c any) error {
	if c == nil {
		return ErrNilConnection
	}
	if p.options.ping == nil {
		return nil
	}
	return p.options.ping(c)
}

func (p *channelPool) Release() {
	p.mu.Lock()
	conns := p.conns
	p.conns = nil
	p.options.factory = nil
	p.options.ping = nil
	closeFun := p.options.close
	p.options.close = nil
	p.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)

	count := len(conns)
	for count > 0 {
		c := <-conns
		closeFun(c)
		count--
	}
}

func (p *channelPool) Len() int {
	return len(p.getConns())
}
