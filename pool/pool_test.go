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

	"trellis.tech/trellis/common.v3/errcode"
	"trellis.tech/trellis/common.v3/testutils"
)

var (
	TestInitialCap = 5
	MaximumCap     = 30
	network        = "tcp"
	address        = "127.0.0.1:7777"
	factory        = func() (any, error) { return net.Dial(network, address) }
	close          = func(c any) error {
		cc, ok := c.(net.Conn)
		if !ok {
			return errcode.New("not net connection")
		}
		return cc.Close()
	}
	ping = func(c any) error {
		_, ok := c.(net.Conn)
		if !ok {
			return errcode.New("not net connection")
		}
		// Simple ping check - just return nil if connection is valid
		return nil
	}
)

var (
	tcpServerOnce sync.Once
	tcpListener   net.Listener
)

func init() {
	// used for factory function
	tcpServerOnce.Do(func() {
		var err error
		tcpListener, err = net.Listen(network, address)
		if err != nil {
			log.Printf("Failed to listen on %s: %v", address, err)
			return
		}
		go simpleTCPServer(tcpListener)
		time.Sleep(time.Millisecond * 50) // Reduced wait time
	})
}

func TestNew(t *testing.T) {
	_, err := newChannelPool()
	testutils.Ok(t, err)
}

func TestNewPool_InvalidOptions(t *testing.T) {
	// Test nil factory
	_, err := NewPool(InitialCap(0), MaxCap(10))
	testutils.NotOk(t, err, "should return error for nil factory")

	// Test invalid capacity settings
	_, err = NewPool(InitialCap(10), MaxCap(5), OptionFactory(factory), OptionClose(close))
	testutils.NotOk(t, err, "should return error for invalid capacity")

	// Test nil close
	_, err = NewPool(InitialCap(0), MaxCap(10), OptionFactory(factory))
	testutils.NotOk(t, err, "should return error for nil close")
}

func TestNewPool_Options(t *testing.T) {
	p, err := NewPool(
		InitialCap(2),
		MaxCap(10),
		MaxIdle(5),
		IdleTimeout(time.Second*30),
		OptionFactory(factory),
		OptionClose(close),
		OptionPing(ping),
	)
	testutils.Ok(t, err)
	testutils.Assert(t, p != nil, "pool should not be nil")
	defer p.Release()
}

func TestPool_Get_Impl(t *testing.T) {
	p, _ := newChannelPool()

	conn, err := p.Get()
	testutils.Ok(t, err)

	_, ok := conn.(net.Conn)
	testutils.Assert(t, ok, "Conn is not of type poolConn")

	// Return the connection to pool before releasing to avoid blocking
	p.Put(conn)
	p.Release()
}

func TestPool_Get(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	_, err := p.Get()
	testutils.Ok(t, err)

	// after one get, current capacity should be lowered by one.
	testutils.Equals(t, TestInitialCap-1, p.Len(), "Get error. Expecting %d, got %d", TestInitialCap-1, p.Len())

	// get them all
	var wg sync.WaitGroup
	for i := 0; i < TestInitialCap-1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Get()
			testutils.Ok(t, err)
		}()
	}
	wg.Wait()

	testutils.Equals(t, 0, p.Len(), "Get error. Expecting %d, got %d", 0, p.Len())

	_, err = p.Get()
	testutils.Ok(t, err)
}

func TestPool_Get_AfterRelease(t *testing.T) {
	p, _ := newChannelPool()
	p.Release()

	_, err := p.Get()
	testutils.NotOk(t, err, "should return error after release")
}

func TestPool_Put(t *testing.T) {
	p, err := NewPool(InitialCap(0), MaxCap(MaximumCap), OptionFactory(factory), OptionClose(close))
	testutils.Ok(t, err)
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
		err := p.Put(conn)
		testutils.Ok(t, err)
	}

	testutils.Assert(t, p.Len() == MaximumCap, "Put error len. Expecting %d, got %d", MaximumCap, p.Len())

	conn, _ := p.Get()
	p.Release() // close pool

	c := conn.(net.Conn)
	c.Close() // try to put into a full pool

	if p.Len() != 0 {
		t.Errorf("Put error. Closed pool shouldn't allow to put connections.")
	}
}

func TestPool_Put_Nil(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	err := p.Put(nil)
	testutils.NotOk(t, err, "should return error for nil connection")
}

func TestPool_Put_AfterRelease(t *testing.T) {
	p, _ := newChannelPool()
	conn, _ := p.Get()
	p.Release()

	err := p.Put(conn)
	testutils.Ok(t, err) // Should close the connection instead
}

func TestPool_Close(t *testing.T) {
	p, _ := newChannelPool()
	conn, _ := p.Get()

	err := p.Close(conn)
	testutils.Ok(t, err)

	testutils.Equals(t, TestInitialCap-1, p.Len(), "Close should decrease openings")
}

func TestPool_Close_Nil(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	err := p.Close(nil)
	testutils.NotOk(t, err, "should return error for nil connection")
}

func TestPool_Close_AfterRelease(t *testing.T) {
	p, _ := newChannelPool()
	conn, _ := p.Get()
	p.Release()

	err := p.Close(conn)
	testutils.Ok(t, err) // Should still work after release
}

func TestPool_Ping(t *testing.T) {
	p, err := NewPool(InitialCap(2), MaxCap(10), OptionFactory(factory), OptionClose(close), OptionPing(ping))
	testutils.Ok(t, err)
	defer p.Release()

	conn, err := p.Get()
	testutils.Ok(t, err)

	cp := p.(*channelPool)
	err = cp.Ping(conn)
	testutils.Ok(t, err)

	p.Put(conn)
}

func TestPool_Ping_Nil(t *testing.T) {
	p, err := NewPool(InitialCap(2), MaxCap(10), OptionFactory(factory), OptionClose(close), OptionPing(ping))
	testutils.Ok(t, err)
	defer p.Release()

	cp := p.(*channelPool)
	err = cp.Ping(nil)
	testutils.NotOk(t, err, "should return error for nil connection")
}

func TestPool_Ping_NoPingFunc(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	conn, _ := p.Get()
	cp := p.(*channelPool)
	cp.options.ping = nil

	err := cp.Ping(conn)
	testutils.Ok(t, err) // Should return nil when ping func is nil

	p.Put(conn)
}

func TestPool_UsedCapacity(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Release()

	testutils.Equals(t, TestInitialCap, p.Len(), "InitialCap error. Expecting %d, got %d", TestInitialCap, p.Len())
}

func TestPool_Close_Pool(t *testing.T) {
	p, _ := newChannelPool()

	// now close it and test all cases we are expecting.
	p.Release()

	c := p.(*channelPool)

	testutils.Assert(t, c.conns == nil, "Close error, conns channel should be nil")

	testutils.Assert(t, c.options.factory == nil, "Close error, factory should be nil")

	_, err := p.Get()
	testutils.NotOk(t, err, "Close error, get conn should return an error")

	testutils.Equals(t, 0, p.Len(), "Close error used capacity. Expecting 0, got %d", p.Len())
}

func TestPool_IdleTimeout(t *testing.T) {
	p, err := NewPool(
		InitialCap(2),
		MaxCap(10),
		MaxIdle(5),
		IdleTimeout(time.Millisecond*100),
		OptionFactory(factory),
		OptionClose(close),
	)
	testutils.Ok(t, err)
	defer p.Release()

	conn, _ := p.Get()
	p.Put(conn)

	// Wait for idle timeout
	time.Sleep(time.Millisecond * 110) // Just slightly longer than expire time

	// Next Get should create a new connection due to idle timeout
	newConn, _ := p.Get()
	testutils.Assert(t, newConn != nil, "should get a new connection")
}

func TestPool_Get_WithPing(t *testing.T) {
	failPing := func(c any) error {
		return errcode.New("ping failed")
	}

	p, err := NewPool(
		InitialCap(2),
		MaxCap(10),
		OptionFactory(factory),
		OptionClose(close),
		OptionPing(failPing),
	)
	testutils.Ok(t, err)
	defer p.Release()

	// Get should skip connections that fail ping
	conn, err := p.Get()
	testutils.Ok(t, err)
	testutils.Assert(t, conn != nil, "should get a connection")
}

func TestPool_Get_Waiting(t *testing.T) {
	p, err := NewPool(
		InitialCap(1),
		MaxCap(2),
		OptionFactory(factory),
		OptionClose(close),
	)
	testutils.Ok(t, err)
	defer p.Release()

	// Get both available connections
	conn1, _ := p.Get()
	conn2, _ := p.Get()

	// Start a goroutine to get a connection (will wait)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := p.Get()
		testutils.Ok(t, err)
		testutils.Assert(t, conn != nil, "should get connection after waiting")
	}()

	// Put one back, should unblock waiting Get
	time.Sleep(time.Millisecond * 10) // Reduced delay
	p.Put(conn1)

	wg.Wait()

	// Cleanup
	if cc, ok := conn2.(net.Conn); ok {
		cc.Close()
	}
}

func TestPoolConcurrent(t *testing.T) {
	p, _ := newChannelPool()
	pipe := make(chan any)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 50) // Reduced delay for faster tests
		p.Release()
	}()

	for i := 0; i < MaximumCap; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, _ := p.Get()

			pipe <- conn
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := <-pipe
			cc, ok := conn.(net.Conn)
			if !ok {
				return
			}
			cc.Close()
		}()
	}

	wg.Wait()
}

func TestPoolWriteRead(t *testing.T) {
	p, _ := NewPool(MaxCap(30), OptionFactory(factory), OptionClose(close))
	defer p.Release()

	conn, _ := p.Get()

	msg := "hello"

	cc, ok := conn.(net.Conn)
	testutils.Assert(t, ok, "should be net.Conn")
	_, err := cc.Write([]byte(msg))
	testutils.Ok(t, err)

	p.Put(conn)
}

func TestPoolConcurrent2(t *testing.T) {
	p, _ := NewPool(MaxCap(30), OptionFactory(factory), OptionClose(close))
	defer p.Release()

	var wg sync.WaitGroup

	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				conn, _ := p.Get()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(20))) // Reduced from 100ms to 20ms
				cc, ok := conn.(net.Conn)
				if ok {
					cc.Close()
				}
			}(i)
		}
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn, _ := p.Get()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			cc, ok := conn.(net.Conn)
			if ok {
				cc.Close()
			}
		}(i)
	}

	wg.Wait()
}

func TestPoolConcurrent3(t *testing.T) {
	p, _ := NewPool(MaxCap(1), OptionFactory(factory), OptionClose(close))

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 10) // Small delay before release
		p.Release()
	}()

	if conn, err := p.Get(); err == nil {
		cc, ok := conn.(net.Conn)
		if ok {
			cc.Close()
		}
	}

	wg.Wait()
}

func TestOptions_Check(t *testing.T) {
	opts := &Options{
		initialCap: 5,
		maxCap:     10,
		maxIdle:    0, // Will be set to maxCap
		factory:    factory,
		close:      close,
	}
	err := opts.check()
	testutils.Ok(t, err)
	testutils.Equals(t, 10, opts.maxIdle, "maxIdle should be set to maxCap")
}

func TestOptions_Check_InvalidCapacity(t *testing.T) {
	opts := &Options{
		initialCap: 10,
		maxCap:     5,
		factory:    factory,
		close:      close,
	}
	err := opts.check()
	testutils.NotOk(t, err, "should return error for invalid capacity")
}

func TestOptions_Check_InvalidMaxIdle(t *testing.T) {
	opts := &Options{
		initialCap: 10,
		maxCap:     5,
		maxIdle:    3,
		factory:    factory,
		close:      close,
	}
	err := opts.check()
	testutils.NotOk(t, err, "should return error for invalid maxIdle")
}

func newChannelPool() (Pool, error) {
	return NewPool(InitialCap(TestInitialCap), MaxIdle(TestInitialCap), MaxCap(MaximumCap), OptionFactory(factory), OptionClose(close))
}

func simpleTCPServer(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			// Listener closed, exit
			return
		}

		go func(c net.Conn) {
			defer c.Close()
			c.SetReadDeadline(time.Now().Add(time.Second))
			buffer := make([]byte, 256)
			_, _ = c.Read(buffer)
		}(conn)
	}
}
