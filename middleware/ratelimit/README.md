# Rate Limiting (Core Package)

This package provides the **core rate limiting implementation** that can be used in any context (HTTP, gRPC, or custom applications).

## Package Overview

**Location**: `github.com/go-trellis/common/middleware/ratelimit`

**Purpose**: Generic rate limiting implementation

## Features

- **Token Bucket Algorithm**: Efficient rate limiting implementation
- **Flexible Key Extraction**: Support for IP, User ID, Path, and custom extractors
- **gRPC Interceptors**: Support for both unary and stream RPCs
- **Generic Interface**: Can be used in any context, not just HTTP

## Architecture

This is the **core** rate limiting package. It provides:

- `Limiter` interface - Defines rate limiting behavior
- `TokenBucketLimiter` - Token bucket algorithm implementation
- `KeyExtractor` - Key extraction functions
- `Config` - Rate limiter configuration
- gRPC interceptors - For gRPC server integration
- Redis limiter - Distributed rate limiting using Redis

## Usage

### gRPC Server (Recommended)

```go
import (
    "time"
    "google.golang.org/grpc"
    "github.com/go-trellis/common/middleware/ratelimit"
)
```

```go
// Create rate limiter
config := ratelimit.NewConfig(100, time.Second)
limiter := ratelimit.NewTokenBucketLimiter(config)

// Create key extractor
keyExtractor := ratelimit.NewKeyExtractor(ratelimit.KeyTypeIP, nil)

// Add interceptors
server := grpc.NewServer(
    grpc.UnaryInterceptor(ratelimit.UnaryRateLimitInterceptor(limiter, keyExtractor)),
    grpc.StreamInterceptor(ratelimit.StreamRateLimitInterceptor(limiter, keyExtractor)),
)
```

### Redis Distributed Limiter

For distributed rate limiting across multiple instances:

```go
import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/go-trellis/common/middleware/ratelimit"
)

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

config := ratelimit.NewConfig(100, time.Second)
limiter := ratelimit.NewRedisLimiter(rdb, config)

allowed, err := limiter.Allow(ctx, "user-key")
if !allowed {
    // Handle rate limit exceeded
}
```

### Custom Implementation

You can also use this package directly for custom rate limiting needs:

```go
import (
    "context"
    "time"
    "github.com/go-trellis/common/middleware/ratelimit"
)

// Create rate limiter
config := ratelimit.NewConfig(100, time.Second)
limiter := ratelimit.NewTokenBucketLimiter(config)

// Check rate limit
allowed, err := limiter.Allow(ctx, "user-key")
if !allowed {
    // Handle rate limit exceeded
}
```

## Key Types

- `KeyTypeIP`: Rate limit by client IP address
- `KeyTypeUserID`: Rate limit by user ID (from context)
- `KeyTypePath`: Rate limit by request path
- `KeyTypeIPPath`: Rate limit by IP + path combination
- `KeyTypeUserPath`: Rate limit by user ID + path combination
- `KeyTypeCustom`: Use custom extractor function

## Custom Key Extractor

```go
import (
    "context"
    "fmt"
    "github.com/go-trellis/common/middleware/ratelimit"
)

customExtractor := func(ctx context.Context) string {
    // Extract custom key from context
    if value := ctx.Value("custom_key"); value != nil {
        return fmt.Sprintf("%v", value)
    }
    return "default"
}

keyExtractor := ratelimit.NewKeyExtractor(ratelimit.KeyTypeCustom, customExtractor)
```

## Configuration

Configure rate limiting programmatically:

```go
import (
    "time"
    "github.com/go-trellis/common/middleware/ratelimit"
)

config := ratelimit.NewConfig(100, time.Second)
config.Burst = 150
limiter := ratelimit.NewTokenBucketLimiter(config)
```

