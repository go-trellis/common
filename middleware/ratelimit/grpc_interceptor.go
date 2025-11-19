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

package ratelimit

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryRateLimitInterceptor creates a unary gRPC interceptor for rate limiting
func UnaryRateLimitInterceptor(limiter Limiter, keyExtractor KeyExtractor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if limiter == nil {
			return handler(ctx, req)
		}

		// Extract key from context
		ctx = enrichContext(ctx, info)
		key := keyExtractor(ctx)

		// Check rate limit
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "rate limiter error: %v", err)
		}

		if !allowed {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// StreamRateLimitInterceptor creates a stream gRPC interceptor for rate limiting
func StreamRateLimitInterceptor(limiter Limiter, keyExtractor KeyExtractor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if limiter == nil {
			return handler(srv, ss)
		}

		ctx := ss.Context()
		ctx = enrichContext(ctx, info)
		key := keyExtractor(ctx)

		// Check rate limit
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			return status.Errorf(codes.Internal, "rate limiter error: %v", err)
		}

		if !allowed {
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(srv, ss)
	}
}

// enrichContext adds metadata to context for key extraction
func enrichContext(ctx context.Context, info interface{}) context.Context {
	// Extract client IP from metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Try to get IP from various headers
		if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
			ip := strings.Split(ips[0], ",")[0]
			ctx = WithValue(ctx, "client_ip", strings.TrimSpace(ip))
		} else if ips := md.Get("x-real-ip"); len(ips) > 0 {
			ctx = WithValue(ctx, "client_ip", ips[0])
		} else if ips := md.Get("remote-addr"); len(ips) > 0 {
			ctx = WithValue(ctx, "client_ip", ips[0])
		}

		// Extract user ID if available
		if userIDs := md.Get("user-id"); len(userIDs) > 0 {
			ctx = WithValue(ctx, "user_id", userIDs[0])
		}
	}

	// Add method path
	var methodPath string
	switch v := info.(type) {
	case *grpc.UnaryServerInfo:
		methodPath = v.FullMethod
	case *grpc.StreamServerInfo:
		methodPath = v.FullMethod
	}
	if methodPath != "" {
		ctx = WithValue(ctx, "request_path", methodPath)
	}

	return ctx
}
