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
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"trellis.tech/trellis/common.v3/utils/flagext"
	"trellis.tech/trellis/common.v3/utils/types"
)

var _ flagext.Parser = (*Config)(nil)

// Config for a new redis.Client.
type Config struct {
	// Addr is the Redis server address in format "host:port"
	// Default: "localhost:6379"
	Addr string `yaml:"addr" json:"addr"`

	// Username is the Redis username (Redis 6.0+)
	Username string `yaml:"username" json:"username"`

	// Password is the Redis password
	Password types.Secret `yaml:"password" json:"password"`

	// DB is the Redis database number (0-15)
	DB int `yaml:"db" json:"db"`

	// DialTimeout is the timeout for connecting to Redis
	DialTimeout types.Duration `yaml:"dial_timeout" json:"dial_timeout"`

	// ReadTimeout is the timeout for read operations
	ReadTimeout types.Duration `yaml:"read_timeout" json:"read_timeout"`

	// WriteTimeout is the timeout for write operations
	WriteTimeout types.Duration `yaml:"write_timeout" json:"write_timeout"`

	// PoolSize is the maximum number of socket connections
	PoolSize int `yaml:"pool_size" json:"pool_size"`

	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int `yaml:"min_idle_conns" json:"min_idle_conns"`

	// MaxRetries is the maximum number of retries for failed commands
	MaxRetries int `yaml:"max_retries" json:"max_retries"`
}

// ParseFlags adds the flags required to config this to the given FlagSet.
func (cfg *Config) ParseFlags(f *flag.FlagSet) {
	cfg.ParseFlagsWithPrefix("", f)
}

// ParseFlagsWithPrefix adds the flags required to config this to the given FlagSet.
func (cfg *Config) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	cfg.Addr = ""
	f.StringVar(&cfg.Addr, prefix+"redis.addr", "localhost:6379", "The Redis server address (host:port).")
	f.StringVar(&cfg.Username, prefix+"redis.username", "", "The Redis username (Redis 6.0+).")
	f.Var(&cfg.Password, prefix+"redis.password", "The Redis password.")
	f.IntVar(&cfg.DB, prefix+"redis.db", 0, "The Redis database number (0-15).")
	f.Var(&cfg.DialTimeout, prefix+"redis.dial-timeout", "The dial timeout for the Redis connection.")
	f.Var(&cfg.ReadTimeout, prefix+"redis.read-timeout", "The read timeout for Redis operations.")
	f.Var(&cfg.WriteTimeout, prefix+"redis.write-timeout", "The write timeout for Redis operations.")
	f.IntVar(&cfg.PoolSize, prefix+"redis.pool-size", 10, "The maximum number of socket connections.")
	f.IntVar(&cfg.MinIdleConns, prefix+"redis.min-idle-conns", 5, "The minimum number of idle connections.")
	f.IntVar(&cfg.MaxRetries, prefix+"redis.max-retries", 3, "The maximum number of retries for failed commands.")
}

// NewClient creates a new Redis client based on the configuration.
func NewClient(cfg Config) (*redis.Client, error) {
	addr := "localhost:6379"
	if cfg.Addr != "" {
		host, port, err := net.SplitHostPort(cfg.Addr)
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "6379"
			host = cfg.Addr
			addr = net.JoinHostPort(host, port)
		} else if err == nil {
			addr = net.JoinHostPort(host, port)
		} else {
			// If SplitHostPort fails, use as-is (might be a unix socket path)
			addr = cfg.Addr
		}
	}

	options := &redis.Options{
		Addr:         addr,
		Username:     cfg.Username,
		Password:     string(cfg.Password),
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	if cfg.DialTimeout > 0 {
		options.DialTimeout = time.Duration(cfg.DialTimeout)
	}
	if cfg.ReadTimeout > 0 {
		options.ReadTimeout = time.Duration(cfg.ReadTimeout)
	}
	if cfg.WriteTimeout > 0 {
		options.WriteTimeout = time.Duration(cfg.WriteTimeout)
	}

	client := redis.NewClient(options)

	return client, nil
}

// NewClientFromOptions creates a new Redis client from redis.Options.
// This is useful when you need more control over the Redis client configuration.
func NewClientFromOptions(options *redis.Options) *redis.Client {
	return redis.NewClient(options)
}
