# Circuit Breaker (Core Package)

This package provides the **core circuit breaker implementation** that can be used in any context (HTTP, gRPC, or custom applications).

## Package Overview

**Location**: `github.com/go-trellis/common.v3/middleware/circuitbreaker`

**Purpose**: Generic circuit breaker implementation

## Features

- **Three States**: Closed (normal), Open (failing fast), Half-Open (testing)
- **Configurable Thresholds**: Customizable failure thresholds and timeouts
- **Thread-Safe**: Safe for concurrent use
- **gRPC Interceptors**: Support for both unary and stream RPCs
- **Generic Interface**: Can be used in any context, not just HTTP

## Architecture

This is the **core** circuit breaker package. It provides:

- `CircuitBreaker` - Main circuit breaker implementation
- `Config` - Circuit breaker configuration
- `State` - Circuit breaker states (Closed, Open, Half-Open)
- `Counts` - Statistics tracking
- gRPC interceptors - For gRPC server integration

## Usage

### Basic Usage

```go
import (
    "context"
    "time"
    "github.com/go-trellis/common.v3/middleware/circuitbreaker"
)

// Create circuit breaker
cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
    Name:        "my-service",
    MaxRequests: 1,
    Interval:    time.Minute,
    Timeout:     time.Minute,
    ReadyToTrip: func(counts circuitbreaker.Counts) bool {
        return counts.ConsecutiveFailures >= 5
    },
})

// Execute a function with circuit breaker protection
err := cb.Execute(ctx, func() error {
    // Your function that might fail
    return doSomething()
})

if err != nil {
    if err == circuitbreaker.ErrCircuitBreakerOpen {
        // Circuit breaker is open, request was rejected
    }
    // Handle other errors
}
```

### gRPC Server

```go
import (
    "google.golang.org/grpc"
    "github.com/go-trellis/common.v3/middleware/circuitbreaker"
)

// Create circuit breaker
cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
    Name:        "grpc-service",
    MaxRequests: 1,
    Interval:    time.Minute,
    Timeout:     time.Minute,
    ReadyToTrip: func(counts circuitbreaker.Counts) bool {
        return counts.ConsecutiveFailures >= 5
    },
})

// Add interceptors
server := grpc.NewServer(
    grpc.UnaryInterceptor(circuitbreaker.UnaryCircuitBreakerInterceptor(cb)),
    grpc.StreamInterceptor(circuitbreaker.StreamCircuitBreakerInterceptor(cb)),
)
```

## Circuit Breaker States

### Closed (Normal Operation)
- Requests are allowed through
- Failures are counted
- When failure threshold is reached, transitions to Open

### Open (Failing Fast)
- All requests are immediately rejected
- Returns error without calling the function
- After timeout period, transitions to Half-Open

### Half-Open (Testing)
- Limited number of requests are allowed (MaxRequests)
- If requests succeed, transitions to Closed
- If requests fail, transitions back to Open

## Configuration

### Config Fields

- **Name**: Name of the circuit breaker (for logging/monitoring)
- **MaxRequests**: Maximum requests allowed in half-open state (default: 1)
- **Interval**: Time period for counting errors (default: 60s)
- **Timeout**: Timeout before attempting to close from open state (default: 60s)
- **ReadyToTrip**: Function that determines when to open the circuit (default: 5 consecutive failures)
- **OnStateChange**: Callback when state changes (optional)

### Configuration

Configure circuit breaker programmatically:

```go
import (
    "time"
    "github.com/go-trellis/common.v3/middleware/circuitbreaker"
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
```

## Error Handling

The circuit breaker returns specific errors:

- `ErrCircuitBreakerOpen`: Circuit breaker is in Open state, request rejected
- `ErrCircuitBreakerHalfOpen`: Circuit breaker is in Half-Open state and max requests reached

## Monitoring

You can monitor the circuit breaker state and statistics:

```go
// Get current state
state := cb.State() // StateClosed, StateOpen, or StateHalfOpen

// Get statistics
counts := cb.Counts()
// counts.Requests, counts.TotalSuccesses, counts.TotalFailures, etc.
```

