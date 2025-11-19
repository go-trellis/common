# Tracing Package

This package provides OpenTracing-like functionality for request tracing, including automatic trace ID generation and propagation through HTTP and gRPC contexts.

## Overview

The tracing package automatically generates and propagates trace IDs across service boundaries, making it easy to track requests through distributed systems.

## Features

- **Automatic Trace ID Generation**: Generates unique trace IDs for each request if not provided
- **Trace ID Propagation**: Extracts trace ID from HTTP headers or gRPC metadata, or generates new ones
- **Context Integration**: Stores trace ID in context for easy access throughout request lifecycle
- **Logging Integration**: Easily integrates with logging systems to include trace ID in log entries
- **HTTP and gRPC Support**: Works with both HTTP (Gin) and gRPC services

## Usage

### HTTP (Gin) Services

Tracing is automatically enabled for all Gin services. The middleware extracts or generates a trace ID and stores it in the request context.

**Trace ID Header**: `X-Trace-Id`

**Example Request**:
```bash
curl -H "X-Trace-Id: my-custom-trace-id" http://localhost:8080/api/endpoint
```

**In Your Handler**:
```go
func MyHandler(c *gin.Context) {
    // Get trace ID from context
    traceID := tracing.TraceIDFromContext(c.Request.Context())
    
    // Use trace ID in your logic or logging
    log.WithField("trace_id", traceID).Info("Processing request")
}
```

### gRPC Services

For gRPC services, trace ID is extracted from metadata or generated automatically.

**Trace ID Metadata Key**: `x-trace-id`

**Example gRPC Client**:
```go
md := metadata.New(map[string]string{
    "x-trace-id": "my-custom-trace-id",
})
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

**In Your gRPC Handler**:
```go
func (s *MyService) MyMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    traceID := tracing.TraceIDFromContext(ctx)
    // Use traceID as needed
    return &pb.Response{}, nil
}
```

### Getting Trace ID from Context

```go
import "trellis.tech/trellis/common.v3/tracing"

// From any context
traceID := tracing.TraceIDFromContext(ctx)
if traceID != "" {
    // Use trace ID
}
```

### Manual Trace ID Setting

```go
import "trellis.tech/trellis/common.v3/tracing"

ctx := tracing.WithTraceID(context.Background(), "my-trace-id")
```

## Trace ID Format

Trace IDs are 32-character hexadecimal strings (16 bytes encoded as hex), generated using cryptographically secure random number generation.

Example: `a1b2c3d4e5f6789012345678901234ab`

## Integration with Logging

You can easily integrate trace IDs with your logging system:

```go
func MyHandler(c *gin.Context) {
    traceID := tracing.TraceIDFromContext(c.Request.Context())
    
    // Include trace ID in logs
    log.WithFields(log.Fields{
        "trace_id": traceID,
        "path": c.Request.URL.Path,
    }).Info("Request received")
}
```

## HTTP Response Headers

The trace ID is automatically added to HTTP response headers as `X-Trace-Id`, allowing clients to track requests across service boundaries.

## Architecture

```
Request → Tracing Middleware → Extract/Generate Trace ID → Store in Context
                                                              ↓
                                                         Logger() → Include trace_id in logs
```

## Constants

- `TraceIDKey`: Context key for storing trace ID (`"trace_id"`)
- `TraceIDHeader`: HTTP header name for trace ID (`"X-Trace-Id"`)
- `TraceIDMetadataKey`: gRPC metadata key for trace ID (`"x-trace-id"`)

## Functions

- `TraceIDFromContext(ctx context.Context) string`: Extract trace ID from context
- `WithTraceID(ctx context.Context, traceID string) context.Context`: Add trace ID to context
- `GinTracingMiddleware() gin.HandlerFunc`: Gin middleware for tracing
- `HTTPTracingHandler(handler http.Handler) http.Handler`: HTTP handler wrapper for tracing

## Integration

To use tracing in your HTTP service, add the middleware:

```go
import (
    "github.com/gin-gonic/gin"
    "trellis.tech/trellis/common.v3/tracing"
)

router := gin.Default()
router.Use(tracing.GinTracingMiddleware())
```

The middleware:
1. Extracts trace ID from request headers (if present)
2. Generates a new trace ID if not present
3. Stores trace ID in context
4. Adds trace ID to response headers

## Best Practices

1. **Propagate trace IDs**: Always include `X-Trace-Id` header when making downstream HTTP requests
2. **Use in logs**: The framework automatically includes trace IDs in logs - no manual work needed
3. **Monitor trace IDs**: Use trace IDs to correlate logs across services
4. **Custom trace IDs**: You can set custom trace IDs via HTTP headers or gRPC metadata

