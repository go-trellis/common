/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

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

package redis

import (
	"flag"
	"testing"
	"time"

	"github.com/go-trellis/common.v3/utils/testutils"
	"github.com/go-trellis/common.v3/utils/types"
	"github.com/redis/go-redis/v9"
)

func TestConfig_ParseFlags(t *testing.T) {
	cfg := &Config{}
	f := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.ParseFlags(f)
	// Should not panic
}

func TestConfig_ParseFlagsWithPrefix(t *testing.T) {
	cfg := &Config{}
	f := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.ParseFlagsWithPrefix("test.", f)
	// Should not panic
}

func TestNewClient_DefaultConfig(t *testing.T) {
	cfg := Config{}
	client, err := NewClient(cfg)
	testutils.Ok(t, err)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, "localhost:6379", client.Options().Addr, "should use default address")
}

func TestNewClient_WithAddr(t *testing.T) {
	cfg := Config{
		Addr: "127.0.0.1:6379",
	}
	client, err := NewClient(cfg)
	testutils.Ok(t, err)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, "127.0.0.1:6379", client.Options().Addr, "should use configured address")
}

func TestNewClient_WithAddrNoPort(t *testing.T) {
	cfg := Config{
		Addr: "127.0.0.1",
	}
	client, err := NewClient(cfg)
	testutils.Ok(t, err)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, "127.0.0.1:6379", client.Options().Addr, "should add default port")
}

func TestNewClient_WithUsernamePassword(t *testing.T) {
	cfg := Config{
		Addr:     "127.0.0.1:6379",
		Username: "testuser",
		Password: types.Secret("testpass"),
	}
	client, err := NewClient(cfg)
	testutils.Ok(t, err)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, "testuser", client.Options().Username, "should use configured username")
	testutils.Equals(t, "testpass", client.Options().Password, "should use configured password")
}

func TestNewClient_WithDB(t *testing.T) {
	cfg := Config{
		Addr: "127.0.0.1:6379",
		DB:   5,
	}
	client, err := NewClient(cfg)
	testutils.Ok(t, err)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, 5, client.Options().DB, "should use configured DB")
}

func TestNewClient_WithTimeouts(t *testing.T) {
	cfg := Config{
		Addr:         "127.0.0.1:6379",
		DialTimeout:  types.Duration(10 * time.Second),
		ReadTimeout:  types.Duration(5 * time.Second),
		WriteTimeout: types.Duration(5 * time.Second),
	}
	client, err := NewClient(cfg)
	testutils.Ok(t, err)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, 10*time.Second, client.Options().DialTimeout, "should use configured dial timeout")
	testutils.Equals(t, 5*time.Second, client.Options().ReadTimeout, "should use configured read timeout")
	testutils.Equals(t, 5*time.Second, client.Options().WriteTimeout, "should use configured write timeout")
}

func TestNewClient_WithPoolConfig(t *testing.T) {
	cfg := Config{
		Addr:         "127.0.0.1:6379",
		PoolSize:     20,
		MinIdleConns: 10,
		MaxRetries:   5,
	}
	client, err := NewClient(cfg)
	testutils.Ok(t, err)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, 20, client.Options().PoolSize, "should use configured pool size")
	testutils.Equals(t, 10, client.Options().MinIdleConns, "should use configured min idle conns")
	testutils.Equals(t, 5, client.Options().MaxRetries, "should use configured max retries")
}

func TestNewClientFromOptions(t *testing.T) {
	options := &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "testpass",
		DB:       0,
	}
	client := NewClientFromOptions(options)
	testutils.Assert(t, client != nil, "client should not be nil")
	testutils.Equals(t, "127.0.0.1:6379", client.Options().Addr, "should use provided address")
}

func TestConfig_ParseFlagsWithPrefix_AllFields(t *testing.T) {
	cfg := &Config{}
	f := flag.NewFlagSet("test", flag.ContinueOnError)

	cfg.ParseFlagsWithPrefix("test.", f)
	// Should not panic - just registers flags, doesn't parse them
}
