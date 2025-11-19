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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

func TestTraceIDFromContext(t *testing.T) {
	t.Run("with trace ID", func(t *testing.T) {
		ctx := WithTraceID(context.Background(), "test-trace-id")
		traceID := TraceIDFromContext(ctx)
		if traceID != "test-trace-id" {
			t.Errorf("expected trace ID 'test-trace-id', got '%s'", traceID)
		}
	})

	t.Run("without trace ID", func(t *testing.T) {
		ctx := context.Background()
		traceID := TraceIDFromContext(ctx)
		if traceID != "" {
			t.Errorf("expected empty trace ID, got '%s'", traceID)
		}
	})
}

func TestGinTracingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("with existing trace ID in header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(TraceIDHeader, "existing-trace-id")

		engine := gin.New()
		engine.Use(GinTracingMiddleware())
		engine.GET("/test", func(c *gin.Context) {
			traceID := TraceIDFromContext(c.Request.Context())
			if traceID != "existing-trace-id" {
				t.Errorf("expected trace ID 'existing-trace-id', got '%s'", traceID)
			}
			c.String(http.StatusOK, "OK")
		})

		engine.ServeHTTP(w, req)

		if w.Header().Get(TraceIDHeader) != "existing-trace-id" {
			t.Errorf("expected trace ID in response header")
		}
	})

	t.Run("without trace ID in header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		engine := gin.New()
		engine.Use(GinTracingMiddleware())
		engine.GET("/test", func(c *gin.Context) {
			traceID := TraceIDFromContext(c.Request.Context())
			if traceID == "" {
				t.Error("expected generated trace ID")
			}
			c.String(http.StatusOK, "OK")
		})

		engine.ServeHTTP(w, req)

		if w.Header().Get(TraceIDHeader) == "" {
			t.Error("expected trace ID in response header")
		}
	})
}

func TestHTTPTracingHandler(t *testing.T) {
	t.Run("with existing trace ID in header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := TraceIDFromContext(r.Context())
			if traceID != "existing-trace-id" {
				t.Errorf("expected trace ID 'existing-trace-id', got '%s'", traceID)
			}
			w.WriteHeader(http.StatusOK)
		})

		wrapped := HTTPTracingHandler(handler)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(TraceIDHeader, "existing-trace-id")

		wrapped.ServeHTTP(w, req)

		if w.Header().Get(TraceIDHeader) != "existing-trace-id" {
			t.Errorf("expected trace ID in response header")
		}
	})

	t.Run("without trace ID in header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := TraceIDFromContext(r.Context())
			if traceID == "" {
				t.Error("expected generated trace ID")
			}
			w.WriteHeader(http.StatusOK)
		})

		wrapped := HTTPTracingHandler(handler)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		wrapped.ServeHTTP(w, req)

		if w.Header().Get(TraceIDHeader) == "" {
			t.Error("expected trace ID in response header")
		}
	})
}

func TestExtractTraceIDFromGRPC(t *testing.T) {
	t.Run("with trace ID in metadata", func(t *testing.T) {
		md := metadata.New(map[string]string{
			TraceIDMetadataKey: "grpc-trace-id",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		traceID := extractTraceIDFromGRPC(ctx)
		if traceID != "grpc-trace-id" {
			t.Errorf("expected trace ID 'grpc-trace-id', got '%s'", traceID)
		}
	})

	t.Run("without trace ID in metadata", func(t *testing.T) {
		ctx := context.Background()
		traceID := extractTraceIDFromGRPC(ctx)
		if traceID == "" {
			t.Error("expected generated trace ID")
		}
		if len(traceID) != 32 {
			t.Errorf("expected trace ID length 32, got %d", len(traceID))
		}
	})
}

func TestRequestIDFromContext(t *testing.T) {
	t.Run("with request ID", func(t *testing.T) {
		ctx := WithRequestID(context.Background(), "test-request-id")
		requestID := RequestIDFromContext(ctx)
		if requestID != "test-request-id" {
			t.Errorf("expected request ID 'test-request-id', got '%s'", requestID)
		}
	})

	t.Run("without request ID", func(t *testing.T) {
		ctx := context.Background()
		requestID := RequestIDFromContext(ctx)
		if requestID != "" {
			t.Errorf("expected empty request ID, got '%s'", requestID)
		}
	})
}

func TestGinTracingMiddlewareWithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("with request ID enabled and existing header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(DefaultRequestIDHeader, "existing-request-id")

		config := &RequestIDConfig{
			Enabled:           true,
			HeaderName:        DefaultRequestIDHeader,
			GenerateIfMissing: true,
		}

		engine := gin.New()
		engine.Use(GinTracingMiddlewareWithConfig(config))
		engine.GET("/test", func(c *gin.Context) {
			requestID := RequestIDFromContext(c.Request.Context())
			if requestID != "existing-request-id" {
				t.Errorf("expected request ID 'existing-request-id', got '%s'", requestID)
			}
			c.String(http.StatusOK, "OK")
		})

		engine.ServeHTTP(w, req)

		if w.Header().Get(DefaultRequestIDHeader) != "existing-request-id" {
			t.Errorf("expected request ID in response header")
		}
	})

	t.Run("with request ID enabled and generate if missing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		config := &RequestIDConfig{
			Enabled:           true,
			HeaderName:        DefaultRequestIDHeader,
			GenerateIfMissing: true,
		}

		engine := gin.New()
		engine.Use(GinTracingMiddlewareWithConfig(config))
		engine.GET("/test", func(c *gin.Context) {
			requestID := RequestIDFromContext(c.Request.Context())
			if requestID == "" {
				t.Error("expected generated request ID")
			}
			c.String(http.StatusOK, "OK")
		})

		engine.ServeHTTP(w, req)

		if w.Header().Get(DefaultRequestIDHeader) == "" {
			t.Error("expected request ID in response header")
		}
	})

	t.Run("with request ID disabled", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		config := &RequestIDConfig{
			Enabled: false,
		}

		engine := gin.New()
		engine.Use(GinTracingMiddlewareWithConfig(config))
		engine.GET("/test", func(c *gin.Context) {
			requestID := RequestIDFromContext(c.Request.Context())
			if requestID != "" {
				t.Errorf("expected empty request ID when disabled, got '%s'", requestID)
			}
			c.String(http.StatusOK, "OK")
		})

		engine.ServeHTTP(w, req)

		if w.Header().Get(DefaultRequestIDHeader) != "" {
			t.Error("expected no request ID in response header when disabled")
		}
	})

	t.Run("with custom header name", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Custom-Request-Id", "custom-request-id")

		config := &RequestIDConfig{
			Enabled:           true,
			HeaderName:        "X-Custom-Request-Id",
			GenerateIfMissing: true,
		}

		engine := gin.New()
		engine.Use(GinTracingMiddlewareWithConfig(config))
		engine.GET("/test", func(c *gin.Context) {
			requestID := RequestIDFromContext(c.Request.Context())
			if requestID != "custom-request-id" {
				t.Errorf("expected request ID 'custom-request-id', got '%s'", requestID)
			}
			c.String(http.StatusOK, "OK")
		})

		engine.ServeHTTP(w, req)

		if w.Header().Get("X-Custom-Request-Id") != "custom-request-id" {
			t.Errorf("expected custom request ID in response header")
		}
	})
}

func TestHTTPTracingHandlerWithRequestID(t *testing.T) {
	t.Run("with request ID enabled", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := RequestIDFromContext(r.Context())
			if requestID == "" {
				t.Error("expected generated request ID")
			}
			w.WriteHeader(http.StatusOK)
		})

		config := &RequestIDConfig{
			Enabled:           true,
			HeaderName:        DefaultRequestIDHeader,
			GenerateIfMissing: true,
		}

		wrapped := HTTPTracingHandlerWithConfig(handler, config)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		wrapped.ServeHTTP(w, req)

		if w.Header().Get(DefaultRequestIDHeader) == "" {
			t.Error("expected request ID in response header")
		}
	})

	t.Run("with request ID disabled", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := RequestIDFromContext(r.Context())
			if requestID != "" {
				t.Errorf("expected empty request ID when disabled, got '%s'", requestID)
			}
			w.WriteHeader(http.StatusOK)
		})

		config := &RequestIDConfig{
			Enabled: false,
		}

		wrapped := HTTPTracingHandlerWithConfig(handler, config)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		wrapped.ServeHTTP(w, req)

		if w.Header().Get(DefaultRequestIDHeader) != "" {
			t.Error("expected no request ID in response header when disabled")
		}
	})
}
