# common

A comprehensive Go utility library containing various tools and helpers for building robust applications.

## Installation

```bash
go get github.com/go-trellis/common.v3
```

## Features

### Core Modules

- **Configuration** (`config`): JSON/YAML config with `#include` support and variable substitution
- **Logging** (`logger`): Logrus integration with file rotation (time/size-based)
- **Cryptography** (`crypto`): Hash, encryption, JWT, TLS helpers
- **Database** (`orm/txorm`, `orm/transaction`): XORM wrapper and transaction management
- **Cache** (`storage/cache`): LRU cache with expiration and table management
- **Connection Pool** (`storage/pool`): Generic connection pool with health checks
- **Data Structures** (`storage/data-structures`): Stack, queue, MPSC lock-free queue

### Middleware

- **Rate Limiting** (`middleware/ratelimit`): Token bucket with Redis support
- **Circuit Breaker** (`middleware/circuitbreaker`): Three-state circuit breaker
- **Tracing** (`middleware/tracing`): Trace ID generation and propagation

### Utilities

- **retry**: Retry mechanism with exponential/linear/fixed backoff
- **slice**: Comprehensive slice operations (filter, map, reduce, unique, intersect, etc.)
- **maputil**: Map utilities (keys, values, filter, merge, group, etc.)
- **uuid** (`id/uuid`): UUID generation and validation
- **random**: Random string/number generation with crypto/rand
- **validation**: Data validation (email, URL, length, regex, numeric, etc.)
- **http**: HTTP client utilities with JSON support
- **ctxutil**: Context utilities with type-safe value access
- **path**: Path manipulation and file system utilities
- **types**: Type conversions, time formatting, string utilities
- **files**: File reading with compression support
- **flagext**: Extended flag parsing
- **json**: JSON encoding/decoding utilities
- **shell**: Shell command execution utilities
- **assets**: Embedded assets management
- **builder**: Build information utilities
- **testutils**: Testing utilities and assertions

### Other

- **Error Handling** (`errors/errcode`): Structured error codes and aggregation
- **State Machine** (`state-machine/fsm`): Finite state machine with YAML config
- **ID Generation** (`id`): Snowflake ID generator, UUID generation and validation
- **Event System** (`event-plugin`): Event bus, dependency injection, plugin system

## Quick Start

### Configuration

```go
import "github.com/go-trellis/common.v3/config"

cfg, err := config.NewConfig("config.yaml")
appName := cfg.GetString("app.name")
```

**Include support**: Use `#include` in YAML config files (must be quoted: `"#include": "other.yaml"`)

### Logging

```go
import "github.com/go-trellis/common.v3/logger"

config := logger.DefaultRotateLogsConfig("/var/log/app.log")
config.RotateMode = logger.RotateModeDay
config.MaxSize = 100 * 1024 * 1024
logrusLogger, _ := logger.NewLogrusLoggerWithRotate(config)
```

### Common Utilities

```go
import (
    "github.com/go-trellis/common.v3/utils/slice"
    "github.com/go-trellis/common.v3/id/uuid"
    "github.com/go-trellis/common.v3/utils/retry"
)

// Slice operations
evens := slice.Filter([]int{1,2,3,4,5}, func(x int) bool { return x%2 == 0 })
sum := slice.Reduce(numbers, 0, func(acc, x int) int { return acc + x })

// UUID generation
id := uuid.New()

// Retry mechanism
err := retry.Do(ctx, retry.Config{MaxRetries: 3}, func() error {
    return someOperation()
})
```

## Testing

```bash
make unittest    # Run all tests
make gofmt       # Format code
make build       # Build all packages
```

## Documentation

- [config](config/README.md) - Configuration management
- [cache](storage/cache/README.md) - LRU cache implementation
- [snowflake](id/snowflake/README.md) - Snowflake ID generator
- [fsm](state-machine/fsm/README.md) - Finite state machine
- [ratelimit](middleware/ratelimit/README.md) - Rate limiting
- [circuitbreaker](middleware/circuitbreaker/README.md) - Circuit breaker
- [tracing](middleware/tracing/README.md) - Request tracing
