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
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/go-trellis/common/errcode"
)

var (
	TestInitialCap = 5
	MaximumCap     = 30
	network        = "tcp"
	address        = "127.0.0.1:7777"
	factory = func() (any, error) { return net.Dial(network, address) }
	// Named closeConn (not close) so it does not shadow the builtin close used by channelPool.
	closeConn = func(c any) error {
		cc, ok := c.(net.Conn)
		if !ok {
			return errcode.New("not net connection")
		}
		if tc, ok := cc.(*net.TCPConn); ok {
			_ = tc.SetLinger(0)
		}
		return cc.Close()
	}

	r *rand.Rand
)

func init() {
	// used for factory function
	go simpleTCPServer()
	time.Sleep(time.Millisecond * 300) // wait until tcp server has been settled

	r = rand.New(rand.NewSource(time.Now().Unix()))
}

func TestNew(t *testing.T) {
	_, err := newChannelPool()
	if err != nil {
		t.Errorf("New error: %s", err)
	}
}
func TestPool_Get_Impl(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	conn, err := p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}

	_, ok := conn.(net.Conn)
	if !ok {
		t.Errorf("Conn is not of type poolConn")
	}
}

func TestPool_Get(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	_, err := p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}

	// after one get, current capacity should be lowered by one.
	if p.Len() != TestInitialCap-1 {
		t.Errorf("Get error. Expecting %d, got %d",
			TestInitialCap-1, p.Len())
	}

	// get them all
	var wg sync.WaitGroup
	for i := 0; i < TestInitialCap-1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Get()
			if err != nil {
				t.Errorf("Get error: %s", err)
			}
		}()
	}
	wg.Wait()

	if p.Len() != 0 {
		t.Errorf("Get error. Expecting %d, got %d",
			TestInitialCap-1, p.Len())
	}

	_, err = p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}
}

func TestPool_Put(t *testing.T) {
	p, err := NewPool(InitialCap(0), MaxCap(MaximumCap), OptionFactory(factory), OptionClose(closeConn))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Release()

	// get/create from the pool
	conns := make([]net.Conn, MaximumCap)
	for i := 0; i < MaximumCap; i++ {
		conn, _ := p.Get()
		if c, ok := conn.(net.Conn); ok {
			conns[i] = c
		}
	}

	// now put them all back
	for _, conn := range conns {
		p.Put(conn)
	}

	if p.Len() != MaximumCap {
		t.Errorf("Put error len. Expecting %d, got %d",
			1, p.Len())
	}

	conn, _ := p.Get()
	p.Release() // close pool

	c := conn.(net.Conn)
	c.Close() // try to put into a full pool

	if p.Len() != 0 {
		t.Errorf("Put error. Closed pool shouldn't allow to put connections.")
	}
}

func TestPool_UsedCapacity(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	if p.Len() != TestInitialCap {
		t.Errorf("InitialCap error. Expecting %d, got %d",
			TestInitialCap, p.Len())
	}
}

func TestPool_Close(t *testing.T) {
	p, _ := newChannelPool()

	// now close it and test all cases we are expecting.
	p.Release()

	c := p.(*channelPool)

	if c.conns != nil {
		t.Errorf("Close error, conns channel should be nil")
	}

	if c.options.factory != nil {
		t.Errorf("Close error, factory should be nil")
	}

	_, err := p.Get()
	if err == nil {
		t.Errorf("Close error, get conn should return an error")
	}

	if p.Len() != 0 {
		t.Errorf("Close error used capacity. Expecting 0, got %d", p.Len())
	}
}

func TestPoolConcurrent(t *testing.T) {
	p, _ := newChannelPool()
	pipe := make(chan any)

	go func() {
		p.Release()
	}()

	for i := 0; i < MaximumCap; i++ {
		go func() {
			conn, _ := p.Get()

			pipe <- conn
		}()

		go func() {
			conn := <-pipe
			cc, ok := conn.(net.Conn)
			if !ok {
				return
			}
			cc.Close()
		}()
	}
}

func TestPoolWriteRead(t *testing.T) {
	p, _ := NewPool(MaxCap(30), OptionFactory(factory), OptionClose(closeConn))

	conn, _ := p.Get()

	msg := "hello"

	cc, ok := conn.(net.Conn)
	if !ok {
		return
	}
	_, err := cc.Write([]byte(msg))
	if err != nil {
		t.Error(err)
	}
}

func TestPoolConcurrent2(t *testing.T) {
	p, _ := NewPool(MaxCap(30), OptionFactory(factory), OptionClose(closeConn))

	var wg sync.WaitGroup

	go func() {
		for i := range 10 {
			wg.Add(1)
			go func(i int) {
				conn, _ := p.Get()
				time.Sleep(time.Millisecond * time.Duration(r.Intn(100)))
				cc, ok := conn.(net.Conn)
				if ok {
					cc.Close()
				}
				wg.Done()
			}(i)
		}
	}()

	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			conn, _ := p.Get()
			time.Sleep(time.Millisecond * time.Duration(r.Intn(100)))
			cc, ok := conn.(net.Conn)
			if ok {
				cc.Close()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestPoolConcurrent3(t *testing.T) {
	p, _ := NewPool(MaxCap(1), OptionFactory(factory), OptionClose(closeConn))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		p.Release()
		wg.Done()
	}()

	if conn, err := p.Get(); err == nil {
		cc, ok := conn.(net.Conn)
		if ok {
			cc.Close()
		}
	}

	wg.Wait()
}

func TestReleaseWakesWaiters(t *testing.T) {
	var created int
	p, err := NewPool(
		InitialCap(0),
		MaxIdle(1),
		MaxCap(1),
		OptionFactory(func() (any, error) {
			created++
			return created, nil
		}),
		OptionClose(func(c any) error { return nil }),
	)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := p.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if conn == nil {
		t.Fatal("expected connection")
	}

	done := make(chan error, 1)
	go func() {
		_, err := p.Get()
		done <- err
	}()

	// Wait until the second Get is parked on the wait queue.
	cp := p.(*channelPool)
	deadline := time.Now().Add(2 * time.Second)
	for {
		cp.mu.Lock()
		nwait := len(cp.waitings)
		cp.mu.Unlock()
		if nwait >= 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for Get to enqueue")
		}
		time.Sleep(10 * time.Millisecond)
	}

	p.Release()

	select {
	case err := <-done:
		if err != ErrPoolClosed {
			t.Fatalf("waiter error = %v, want %v", err, ErrPoolClosed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Release did not wake waiting Get")
	}
}

func TestReleaseClosesIdleConnValue(t *testing.T) {
	closed := make(chan any, 1)
	p, err := NewPool(
		InitialCap(1),
		MaxIdle(1),
		MaxCap(1),
		OptionFactory(func() (any, error) { return "conn", nil }),
		OptionClose(func(c any) error {
			closed <- c
			return nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	p.Release()

	select {
	case c := <-closed:
		if c != "conn" {
			t.Fatalf("close got %v (%T), want underlying conn value", c, c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Release did not close idle connection")
	}
}

func newChannelPool() (Pool, error) {
	return NewPool(InitialCap(TestInitialCap), MaxIdle(TestInitialCap), MaxCap(MaximumCap), OptionFactory(factory), OptionClose(closeConn))
}

func simpleTCPServer() {
	l, err := net.Listen(network, address)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			buffer := make([]byte, 256)
			conn.Read(buffer)
		}()
	}
}
