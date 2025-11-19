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
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryRateLimitInterceptor(t *testing.T) {
	t.Run("nil limiter", func(t *testing.T) {
		interceptor := UnaryRateLimitInterceptor(nil, nil)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "success", nil
		}
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}
		result, err := interceptor(context.Background(), "request", info, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "success" {
			t.Error("expected success result")
		}
	})

	t.Run("rate limit allowed", func(t *testing.T) {
		config := NewConfig(100, time.Second)
		limiter := NewTokenBucketLimiter(config)
		keyExtractor := NewKeyExtractor(KeyTypeIP, nil)

		interceptor := UnaryRateLimitInterceptor(limiter, keyExtractor)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "success", nil
		}
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"x-forwarded-for": "192.168.1.1",
		}))

		result, err := interceptor(ctx, "request", info, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "success" {
			t.Error("expected success result")
		}
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		config := NewConfig(1, time.Second)
		limiter := NewTokenBucketLimiter(config)
		keyExtractor := NewKeyExtractor(KeyTypeIP, nil)

		interceptor := UnaryRateLimitInterceptor(limiter, keyExtractor)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "success", nil
		}
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"x-forwarded-for": "192.168.1.1",
		}))

		// First request should succeed
		_, err := interceptor(ctx, "request", info, handler)
		if err != nil {
			t.Errorf("first request should succeed: %v", err)
		}

		// Second request should be rate limited
		_, err = interceptor(ctx, "request", info, handler)
		if err == nil {
			t.Error("expected rate limit error")
		}
		statusErr, ok := status.FromError(err)
		if !ok {
			t.Error("expected gRPC status error")
		}
		if statusErr.Code() != codes.ResourceExhausted {
			t.Errorf("expected ResourceExhausted, got %v", statusErr.Code())
		}
	})

	t.Run("limiter error", func(t *testing.T) {
		mockLimiter := &mockErrorLimiter{}
		keyExtractor := NewKeyExtractor(KeyTypeIP, nil)

		interceptor := UnaryRateLimitInterceptor(mockLimiter, keyExtractor)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "success", nil
		}
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

		_, err := interceptor(context.Background(), "request", info, handler)
		if err == nil {
			t.Error("expected error")
		}
		statusErr, ok := status.FromError(err)
		if !ok {
			t.Error("expected gRPC status error")
		}
		if statusErr.Code() != codes.Internal {
			t.Errorf("expected Internal, got %v", statusErr.Code())
		}
	})
}

func TestStreamRateLimitInterceptor(t *testing.T) {
	t.Run("nil limiter", func(t *testing.T) {
		interceptor := StreamRateLimitInterceptor(nil, nil)
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return nil
		}
		info := &grpc.StreamServerInfo{FullMethod: "/test.Method"}
		mockStream := &mockServerStream{ctx: context.Background()}

		err := interceptor(nil, mockStream, info, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("rate limit allowed", func(t *testing.T) {
		config := NewConfig(100, time.Second)
		limiter := NewTokenBucketLimiter(config)
		keyExtractor := NewKeyExtractor(KeyTypeIP, nil)

		interceptor := StreamRateLimitInterceptor(limiter, keyExtractor)
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return nil
		}
		info := &grpc.StreamServerInfo{FullMethod: "/test.Method"}

		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"x-forwarded-for": "192.168.1.1",
		}))
		mockStream := &mockServerStream{ctx: ctx}

		err := interceptor(nil, mockStream, info, handler)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		config := NewConfig(1, time.Second)
		limiter := NewTokenBucketLimiter(config)
		keyExtractor := NewKeyExtractor(KeyTypeIP, nil)

		interceptor := StreamRateLimitInterceptor(limiter, keyExtractor)
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return nil
		}
		info := &grpc.StreamServerInfo{FullMethod: "/test.Method"}

		ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
			"x-forwarded-for": "192.168.1.1",
		}))

		// First request should succeed
		mockStream1 := &mockServerStream{ctx: ctx}
		err := interceptor(nil, mockStream1, info, handler)
		if err != nil {
			t.Errorf("first request should succeed: %v", err)
		}

		// Second request should be rate limited
		mockStream2 := &mockServerStream{ctx: ctx}
		err = interceptor(nil, mockStream2, info, handler)
		if err == nil {
			t.Error("expected rate limit error")
		}
		statusErr, ok := status.FromError(err)
		if !ok {
			t.Error("expected gRPC status error")
		}
		if statusErr.Code() != codes.ResourceExhausted {
			t.Errorf("expected ResourceExhausted, got %v", statusErr.Code())
		}
	})
}

func TestEnrichContext(t *testing.T) {
	t.Run("with x-forwarded-for", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"x-forwarded-for": "192.168.1.1, 10.0.0.1",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

		enriched := enrichContext(ctx, info)
		if enriched.Value("client_ip") != "192.168.1.1" {
			t.Errorf("expected client_ip 192.168.1.1, got %v", enriched.Value("client_ip"))
		}
		if enriched.Value("request_path") != "/test.Method" {
			t.Errorf("expected request_path /test.Method, got %v", enriched.Value("request_path"))
		}
	})

	t.Run("with x-real-ip", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"x-real-ip": "192.168.1.2",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

		enriched := enrichContext(ctx, info)
		if enriched.Value("client_ip") != "192.168.1.2" {
			t.Errorf("expected client_ip 192.168.1.2, got %v", enriched.Value("client_ip"))
		}
	})

	t.Run("with user-id", func(t *testing.T) {
		md := metadata.New(map[string]string{
			"user-id": "user123",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

		enriched := enrichContext(ctx, info)
		if enriched.Value("user_id") != "user123" {
			t.Errorf("expected user_id user123, got %v", enriched.Value("user_id"))
		}
	})

	t.Run("with stream info", func(t *testing.T) {
		ctx := context.Background()
		info := &grpc.StreamServerInfo{FullMethod: "/test.StreamMethod"}

		enriched := enrichContext(ctx, info)
		if enriched.Value("request_path") != "/test.StreamMethod" {
			t.Errorf("expected request_path /test.StreamMethod, got %v", enriched.Value("request_path"))
		}
	})
}

type mockErrorLimiter struct{}

func (m *mockErrorLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return false, status.Errorf(codes.Internal, "limiter error")
}

func (m *mockErrorLimiter) Wait(ctx context.Context, key string) error {
	return status.Errorf(codes.Internal, "limiter error")
}

type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStream) RecvMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetTrailer(md metadata.MD) {
}
