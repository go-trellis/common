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
	"net/http"
)

// HTTPTracingHandler wraps an http.Handler with tracing support
func HTTPTracingHandler(handler http.Handler) http.Handler {
	return HTTPTracingHandlerWithConfig(handler, nil)
}

// HTTPTracingHandlerWithConfig wraps an http.Handler with tracing and request ID support
func HTTPTracingHandlerWithConfig(handler http.Handler, config *RequestIDConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract or generate trace ID from header
		traceID := r.Header.Get(TraceIDHeader)
		if traceID == "" {
			traceID = generateTraceID()
		}
		ctx = WithTraceID(ctx, traceID)
		w.Header().Set(TraceIDHeader, traceID)

		// Handle request ID if enabled
		if config != nil && config.Enabled {
			headerName := config.HeaderName
			if headerName == "" {
				headerName = DefaultRequestIDHeader
			}

			requestID := r.Header.Get(headerName)
			if requestID == "" && config.GenerateIfMissing {
				requestID = generateRequestID()
			}

			if requestID != "" {
				ctx = WithRequestID(ctx, requestID)
				w.Header().Set(headerName, requestID)
			}
		}

		r = r.WithContext(ctx)
		handler.ServeHTTP(w, r)
	})
}
