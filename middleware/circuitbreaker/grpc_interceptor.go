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

package circuitbreaker

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryCircuitBreakerInterceptor creates a unary gRPC interceptor for circuit breaking
func UnaryCircuitBreakerInterceptor(cb *CircuitBreaker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cb == nil {
			return handler(ctx, req)
		}

		var result interface{}
		var err error

		executeErr := cb.Execute(ctx, func() error {
			result, err = handler(ctx, req)
			if err != nil {
				// Check if error is a gRPC error with status code >= 500
				if st, ok := status.FromError(err); ok {
					if st.Code() >= codes.Internal {
						return err
					}
				}
				// For non-gRPC errors, treat as failure
				return err
			}
			return nil
		})

		if executeErr != nil {
			if executeErr == ErrCircuitBreakerOpen || executeErr == ErrCircuitBreakerHalfOpen {
				return nil, status.Errorf(codes.Unavailable, "circuit breaker is open")
			}
			return nil, executeErr
		}

		return result, err
	}
}

// StreamCircuitBreakerInterceptor creates a stream gRPC interceptor for circuit breaking
func StreamCircuitBreakerInterceptor(cb *CircuitBreaker) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if cb == nil {
			return handler(srv, ss)
		}

		var err error

		executeErr := cb.Execute(ss.Context(), func() error {
			err = handler(srv, ss)
			if err != nil {
				// Check if error is a gRPC error with status code >= 500
				if st, ok := status.FromError(err); ok {
					if st.Code() >= codes.Internal {
						return err
					}
				}
				// For non-gRPC errors, treat as failure
				return err
			}
			return nil
		})

		if executeErr != nil {
			if executeErr == ErrCircuitBreakerOpen || executeErr == ErrCircuitBreakerHalfOpen {
				return status.Errorf(codes.Unavailable, "circuit breaker is open")
			}
			return executeErr
		}

		return err
	}
}
