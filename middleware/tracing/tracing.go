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

package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

const (
	// TraceIDKey is the key used to store trace ID in context
	TraceIDKey = "trace_id"
	// TraceIDHeader is the HTTP header name for trace ID
	TraceIDHeader = "X-Trace-Id"
	// TraceIDMetadataKey is the gRPC metadata key for trace ID
	TraceIDMetadataKey = "x-trace-id"

	// RequestIDKey is the key used to store request ID in context
	RequestIDKey = "request_id"
	// DefaultRequestIDHeader is the default HTTP header name for request ID
	DefaultRequestIDHeader = "X-Request-Id"
	// DefaultRequestIDMetadataKey is the default gRPC metadata key for request ID
	DefaultRequestIDMetadataKey = "x-request-id"
)

// RequestIDConfig contains configuration for request ID generation and propagation
type RequestIDConfig struct {
	// Enabled enables request ID generation and propagation
	Enabled bool

	// HeaderName is the HTTP header name for request ID (default: "X-Request-Id")
	HeaderName string

	// MetadataKey is the gRPC metadata key for request ID (default: "x-request-id")
	MetadataKey string

	// GenerateIfMissing generates a new request ID if not present in request (default: true)
	GenerateIfMissing bool
}

// TraceIDFromContext extracts trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		if id, ok := traceID.(string); ok && id != "" {
			return id
		}
	}
	return ""
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// generateTraceID generates a new trace ID
func generateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%x", b)
	}
	return hex.EncodeToString(b)
}

// generateRequestID generates a new request ID
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%x", b)
	}
	return hex.EncodeToString(b)
}

// RequestIDFromContext extracts request ID from context
func RequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok && id != "" {
			return id
		}
	}
	return ""
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// extractTraceIDFromGin extracts trace ID from Gin context (HTTP headers)
func extractTraceIDFromGin(c *gin.Context) string {
	// Try to get trace ID from header
	if traceID := c.GetHeader(TraceIDHeader); traceID != "" {
		return traceID
	}
	// Generate new trace ID if not present
	return generateTraceID()
}

// extractTraceIDFromGRPC extracts trace ID from gRPC metadata
func extractTraceIDFromGRPC(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if traceIDs := md.Get(TraceIDMetadataKey); len(traceIDs) > 0 && traceIDs[0] != "" {
			return traceIDs[0]
		}
	}
	// Generate new trace ID if not present
	return generateTraceID()
}

// GinTracingMiddleware creates a Gin middleware for tracing
func GinTracingMiddleware() gin.HandlerFunc {
	return GinTracingMiddlewareWithConfig(nil)
}

// GinTracingMiddlewareWithConfig creates a Gin middleware for tracing with request ID support
func GinTracingMiddlewareWithConfig(config *RequestIDConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Extract or generate trace ID
		traceID := extractTraceIDFromGin(c)
		ctx = WithTraceID(ctx, traceID)
		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)

		// Handle request ID if enabled
		if config != nil && config.Enabled {
			headerName := config.HeaderName
			if headerName == "" {
				headerName = DefaultRequestIDHeader
			}

			requestID := c.GetHeader(headerName)
			if requestID == "" && config.GenerateIfMissing {
				requestID = generateRequestID()
			}

			if requestID != "" {
				ctx = WithRequestID(ctx, requestID)
				c.Set(RequestIDKey, requestID)
				c.Header(headerName, requestID)
			}
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// GRPCTracingUnaryInterceptor creates a unary gRPC interceptor for tracing
// This is a helper function that wraps the standard gRPC interceptor pattern
// Note: For actual gRPC usage, you should use grpc.UnaryServerInterceptor wrapper
func GRPCTracingUnaryInterceptor() func(ctx context.Context, req interface{}, info interface{}, handler interface{}) (interface{}, error) {
	return GRPCTracingUnaryInterceptorWithConfig(nil)
}

// GRPCTracingUnaryInterceptorWithConfig creates a unary gRPC interceptor for tracing with request ID support
func GRPCTracingUnaryInterceptorWithConfig(config *RequestIDConfig) func(ctx context.Context, req interface{}, info interface{}, handler interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info interface{}, handler interface{}) (interface{}, error) {
		// Extract or generate trace ID
		traceID := extractTraceIDFromGRPC(ctx)
		ctx = WithTraceID(ctx, traceID)

		// Handle request ID if enabled
		if config != nil && config.Enabled {
			metadataKey := config.MetadataKey
			if metadataKey == "" {
				metadataKey = DefaultRequestIDMetadataKey
			}

			requestID := extractRequestIDFromGRPC(ctx, metadataKey)
			if requestID == "" && config.GenerateIfMissing {
				requestID = generateRequestID()
			}

			if requestID != "" {
				ctx = WithRequestID(ctx, requestID)
			}
		}

		// Call handler
		if h, ok := handler.(func(context.Context, interface{}) (interface{}, error)); ok {
			return h(ctx, req)
		}
		return nil, fmt.Errorf("invalid handler type")
	}
}

// extractRequestIDFromGRPC extracts request ID from gRPC metadata
func extractRequestIDFromGRPC(ctx context.Context, metadataKey string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if requestIDs := md.Get(metadataKey); len(requestIDs) > 0 && requestIDs[0] != "" {
			return requestIDs[0]
		}
	}
	return ""
}

// GRPCTracingStreamInterceptor creates a stream gRPC interceptor for tracing
// This is a helper function that wraps the standard gRPC interceptor pattern
// Note: For actual gRPC usage, you should use grpc.StreamServerInterceptor wrapper
func GRPCTracingStreamInterceptor() func(srv interface{}, ss interface{}, info interface{}, handler interface{}) error {
	return GRPCTracingStreamInterceptorWithConfig(nil)
}

// GRPCTracingStreamInterceptorWithConfig creates a stream gRPC interceptor for tracing with request ID support
func GRPCTracingStreamInterceptorWithConfig(config *RequestIDConfig) func(srv interface{}, ss interface{}, info interface{}, handler interface{}) error {
	return func(srv interface{}, ss interface{}, info interface{}, handler interface{}) error {
		// Get stream context
		type streamContext interface {
			Context() context.Context
		}
		stream, ok := ss.(streamContext)
		if !ok {
			return fmt.Errorf("invalid stream type")
		}

		ctx := stream.Context()
		// Extract or generate trace ID
		traceID := extractTraceIDFromGRPC(ctx)
		ctx = WithTraceID(ctx, traceID)

		// Handle request ID if enabled
		if config != nil && config.Enabled {
			metadataKey := config.MetadataKey
			if metadataKey == "" {
				metadataKey = DefaultRequestIDMetadataKey
			}

			requestID := extractRequestIDFromGRPC(ctx, metadataKey)
			if requestID == "" && config.GenerateIfMissing {
				requestID = generateRequestID()
			}

			if requestID != "" {
				ctx = WithRequestID(ctx, requestID)
			}
		}

		// Call handler
		if h, ok := handler.(func(interface{}, interface{}) error); ok {
			return h(srv, ss)
		}
		return fmt.Errorf("invalid handler type")
	}
}
