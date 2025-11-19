# common

A comprehensive Go utility library containing various tools and helpers for building robust applications.

## Installation

```bash
go get trellis.tech/trellis/common.v3
```

## Features

### Configuration Management (`config`)
- Read JSON, YAML configuration files
- Support for `#include` directive to load other config files
- Variable substitution with `${key.path}` syntax
- Deep merge for included configurations
- Type-safe getters with default values
- Dot notation for nested keys

### Logging (`logger`)
- Logrus integration with structured logging
- File rotation with time-based (hour/day) and size-based rotation
- Multiple log levels support
- Custom log output handlers

### Cryptography (`crypto`)
- Hash functions: MD5, SHA series (SHA1, SHA224, SHA256, SHA384, SHA512, etc.)
- CRC32 checksum support
- Base64 encoding/decoding
- JWT token handling
- RC4 encryption (ArcFour)
- AES encryption with ECB mode
- TLS configuration helpers

### Data Structures (`storage/data-structures`)
- Stack implementation with thread-safe operations
- Queue implementation (FIFO)
- Multi-Producer Single-Consumer (MPSC) lock-free queue

### Cache (`storage/cache`)
- LRU cache implementation
- Multiple value modes: Unique, Bag, DuplicateBag
- Expiration support
- Table-based cache management

### Connection Pool (`storage/pool`)
- Generic connection pool
- Configurable capacity and idle timeout
- Connection health checks (ping)
- Concurrent-safe operations

### Error Handling (`errors/errcode`)
- Structured error codes
- Error context support
- Multiple error aggregation
- Simple error wrapper

### State Machine (`state-machine/fsm`)
- Finite state machine implementation
- Namespace-based state management
- Transition validation
- Configurable via YAML

### ID Generation (`id/snowflake`)
- Twitter Snowflake ID generator
- Configurable worker and datacenter bits
- High-performance ID generation

### Database (`orm/txorm`)
- XORM transaction wrapper
- Prometheus metrics integration
- SQL builder helpers
- Transaction management

### Transaction (`orm/transaction`)
- Generic transaction management interface
- Transaction engine abstraction
- Commit and rollback support

### Rate Limiting (`middleware/ratelimit`)
- Token bucket algorithm implementation
- Flexible key extraction (IP, User ID, Path, custom)
- gRPC interceptors support
- Redis-based distributed rate limiting
- Multi-limiter support

### Circuit Breaker (`middleware/circuitbreaker`)
- Three-state circuit breaker (Closed, Open, Half-Open)
- Configurable failure thresholds and timeouts
- Thread-safe implementation
- gRPC interceptors support
- State change callbacks

### Tracing (`middleware/tracing`)
- Automatic trace ID generation
- Trace ID propagation through HTTP and gRPC
- Gin middleware integration
- Context-based trace ID storage
- HTTP header support (`X-Trace-Id`)

### Other Utilities
- **utils/types**: Type conversions, time formatting, string utilities
- **utils/files**: File reading utilities with compression support
- **event-plugin/event**: Event bus for pub/sub patterns
- **event-plugin/injector**: Dependency injection helpers
- **utils/flagext**: Extended flag parsing

## Quick Start

### Configuration with Include

```yaml
# config.yaml
"#include": "common.yaml"
app:
  name: "my-app"
  version: "1.0.0"
```

```go
import "trellis.tech/trellis/common.v3/config"

cfg, err := config.NewConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

appName := cfg.GetString("app.name")
version := cfg.GetString("app.version")
```

### Logging with File Rotation

```go
import "trellis.tech/trellis/common.v3/logger"

config := logger.DefaultRotateLogsConfig("/var/log/app.log")
config.RotateMode = logger.RotateModeDay
config.MaxAge = 7 * 24 * time.Hour
config.MaxSize = 100 * 1024 * 1024 // 100MB

logrusLogger, err := logger.NewLogrusLoggerWithRotate(config)
```

### Snowflake ID Generation

```go
import "trellis.tech/trellis/common.v3/id/snowflake"

worker, _ := snowflake.NewWorker()
id := worker.Next()
```

### Cache Usage

```go
import "trellis.tech/trellis/common.v3/storage/cache"

c := cache.New("table1", cache.OptValueMode(cache.ValueModeUnique))
c.Insert("key1", "value1")
values, ok := c.Lookup("key1")
```

### Connection Pool

```go
import "trellis.tech/trellis/common.v3/storage/pool"

factory := func() (any, error) {
    return net.Dial("tcp", "localhost:8080")
}

p, _ := pool.NewPool(
    pool.OptionFactory(factory),
    pool.OptionClose(func(conn any) error {
        return conn.(net.Conn).Close()
    }),
    pool.InitialCap(5),
    pool.MaxCap(10),
)

conn, _ := p.Get()
defer p.Put(conn)
```

### Rate Limiting

```go
import (
    "time"
    "trellis.tech/trellis/common.v3/middleware/ratelimit"
)

config := ratelimit.NewConfig(100, time.Second)
limiter := ratelimit.NewTokenBucketLimiter(config)

allowed, err := limiter.Allow(ctx, "user-key")
if !allowed {
    // Handle rate limit exceeded
}
```

### Circuit Breaker

```go
import (
    "context"
    "time"
    "trellis.tech/trellis/common.v3/middleware/circuitbreaker"
)

cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
    Name:        "my-service",
    MaxRequests: 1,
    Interval:    time.Minute,
    Timeout:     time.Minute,
    ReadyToTrip: func(counts circuitbreaker.Counts) bool {
        return counts.ConsecutiveFailures >= 5
    },
})

err := cb.Execute(ctx, func() error {
    return doSomething()
})
```

### Tracing

```go
import (
    "github.com/gin-gonic/gin"
    "trellis.tech/trellis/common.v3/middleware/tracing"
)

router := gin.Default()
router.Use(tracing.GinTracingMiddleware())

// In your handler
func MyHandler(c *gin.Context) {
    traceID := tracing.TraceIDFromContext(c.Request.Context())
    // Use traceID in your logic
}
```

## Configuration Include Feature

The config package supports including other configuration files using the `#include` directive:

```yaml
# main.yaml
"#include": "database.yaml"
app:
  name: "my-app"
```

```yaml
# database.yaml
database:
  host: "localhost"
  port: 5432
```

**Note**: In YAML files, the `#include` key must be quoted (`"#include"`) because `#` starts a comment in YAML. In JSON files, quotes are not required.

### Features
- Supports single file or array of files
- Recursive includes
- Circular reference detection
- Deep merge (included configs override main config)
- Relative and absolute paths supported

## Testing

Run all tests:
```bash
make unittest
```

Format code:
```bash
make gofmt
```

Build:
```bash
make build
```

## Sub-packages Documentation

- [config](config/README.md) - Configuration management
- [cache](storage/cache/README.md) - LRU cache implementation
- [snowflake](id/snowflake/README.md) - Snowflake ID generator
- [fsm](state-machine/fsm/README.md) - Finite state machine
- [ratelimit](middleware/ratelimit/README.md) - Rate limiting implementation
- [circuitbreaker](middleware/circuitbreaker/README.md) - Circuit breaker implementation
- [tracing](middleware/tracing/README.md) - Request tracing and trace ID propagation

## Other Useful Utilities

* backoff: https://github.com/grafana/dskit/blob/main/backoff/backoff.go
* grpc limiter: https://github.com/grafana/dskit/tree/main/limiter
* go pool: https://github.com/Jeffail/tunny