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
	"net"
	"testing"
)

func TestExtractByIP(t *testing.T) {
	t.Run("with IP in context", func(t *testing.T) {
		ctx := WithValue(context.Background(), "client_ip", "192.168.1.1")
		ip := ExtractByIP(ctx)
		if ip != "192.168.1.1" {
			t.Errorf("expected 192.168.1.1, got %s", ip)
		}
	})

	t.Run("without IP in context", func(t *testing.T) {
		ctx := context.Background()
		ip := ExtractByIP(ctx)
		if ip != "unknown" {
			t.Errorf("expected 'unknown', got %s", ip)
		}
	})

	t.Run("invalid IP type", func(t *testing.T) {
		ctx := WithValue(context.Background(), "client_ip", 123)
		ip := ExtractByIP(ctx)
		if ip != "unknown" {
			t.Errorf("expected 'unknown', got %s", ip)
		}
	})
}

func TestExtractByUserID(t *testing.T) {
	t.Run("with user ID in context", func(t *testing.T) {
		ctx := WithValue(context.Background(), "user_id", "user123")
		userID := ExtractByUserID(ctx)
		if userID != "user123" {
			t.Errorf("expected 'user123', got %s", userID)
		}
	})

	t.Run("without user ID in context", func(t *testing.T) {
		ctx := context.Background()
		userID := ExtractByUserID(ctx)
		if userID != "anonymous" {
			t.Errorf("expected 'anonymous', got %s", userID)
		}
	})

	t.Run("numeric user ID", func(t *testing.T) {
		ctx := WithValue(context.Background(), "user_id", 12345)
		userID := ExtractByUserID(ctx)
		if userID != "12345" {
			t.Errorf("expected '12345', got %s", userID)
		}
	})
}

func TestExtractByPath(t *testing.T) {
	t.Run("with path in context", func(t *testing.T) {
		ctx := WithValue(context.Background(), "request_path", "/api/test")
		path := ExtractByPath(ctx)
		if path != "/api/test" {
			t.Errorf("expected '/api/test', got %s", path)
		}
	})

	t.Run("without path in context", func(t *testing.T) {
		ctx := context.Background()
		path := ExtractByPath(ctx)
		if path != "unknown" {
			t.Errorf("expected 'unknown', got %s", path)
		}
	})

	t.Run("invalid path type", func(t *testing.T) {
		ctx := WithValue(context.Background(), "request_path", 123)
		path := ExtractByPath(ctx)
		if path != "unknown" {
			t.Errorf("expected 'unknown', got %s", path)
		}
	})
}

func TestExtractByIPPath(t *testing.T) {
	t.Run("with both IP and path", func(t *testing.T) {
		ctx := WithValue(context.Background(), "client_ip", "192.168.1.1")
		ctx = WithValue(ctx, "request_path", "/api/test")
		key := ExtractByIPPath(ctx)
		if key != "192.168.1.1:/api/test" {
			t.Errorf("expected '192.168.1.1:/api/test', got %s", key)
		}
	})

	t.Run("without IP", func(t *testing.T) {
		ctx := WithValue(context.Background(), "request_path", "/api/test")
		key := ExtractByIPPath(ctx)
		if key != "unknown:/api/test" {
			t.Errorf("expected 'unknown:/api/test', got %s", key)
		}
	})
}

func TestExtractByUserPath(t *testing.T) {
	t.Run("with both user ID and path", func(t *testing.T) {
		ctx := WithValue(context.Background(), "user_id", "user123")
		ctx = WithValue(ctx, "request_path", "/api/test")
		key := ExtractByUserPath(ctx)
		if key != "user123:/api/test" {
			t.Errorf("expected 'user123:/api/test', got %s", key)
		}
	})

	t.Run("without user ID", func(t *testing.T) {
		ctx := WithValue(context.Background(), "request_path", "/api/test")
		key := ExtractByUserPath(ctx)
		if key != "anonymous:/api/test" {
			t.Errorf("expected 'anonymous:/api/test', got %s", key)
		}
	})
}

func TestExtractIPFromAddr(t *testing.T) {
	t.Run("IPNet address", func(t *testing.T) {
		_, ipNet, _ := net.ParseCIDR("192.168.1.1/24")
		ip := ExtractIPFromAddr(ipNet)
		if ip == "" || ip == "unknown" {
			t.Errorf("expected valid IP, got %s", ip)
		}
	})

	t.Run("IPAddr address", func(t *testing.T) {
		ipAddr := &net.IPAddr{IP: net.ParseIP("192.168.1.2")}
		ip := ExtractIPFromAddr(ipAddr)
		if ip != "192.168.1.2" {
			t.Errorf("expected 192.168.1.2, got %s", ip)
		}
	})

	t.Run("TCPAddr address", func(t *testing.T) {
		tcpAddr := &net.TCPAddr{IP: net.ParseIP("192.168.1.3"), Port: 8080}
		ip := ExtractIPFromAddr(tcpAddr)
		if ip != "192.168.1.3" {
			t.Errorf("expected 192.168.1.3, got %s", ip)
		}
	})

	t.Run("UDPAddr address", func(t *testing.T) {
		udpAddr := &net.UDPAddr{IP: net.ParseIP("192.168.1.4"), Port: 8080}
		ip := ExtractIPFromAddr(udpAddr)
		if ip != "192.168.1.4" {
			t.Errorf("expected 192.168.1.4, got %s", ip)
		}
	})

	t.Run("loopback address", func(t *testing.T) {
		ipAddr := &net.IPAddr{IP: net.ParseIP("127.0.0.1")}
		ip := ExtractIPFromAddr(ipAddr)
		if ip != "127.0.0.1" {
			t.Errorf("expected 127.0.0.1, got %s", ip)
		}
	})

	t.Run("IPv6 address", func(t *testing.T) {
		ipAddr := &net.IPAddr{IP: net.ParseIP("2001:db8::1")}
		ip := ExtractIPFromAddr(ipAddr)
		if ip != "unknown" {
			t.Errorf("expected 'unknown', got %s", ip)
		}
	})

	t.Run("unsupported address type", func(t *testing.T) {
		ip := ExtractIPFromAddr(nil)
		if ip != "127.0.0.1" {
			t.Errorf("expected '127.0.0.1', got %s", ip)
		}
	})
}

func TestNewKeyExtractor(t *testing.T) {
	ctx := WithValue(context.Background(), "client_ip", "192.168.1.1")
	ctx = WithValue(ctx, "user_id", "user123")
	ctx = WithValue(ctx, "request_path", "/api/test")

	t.Run("KeyTypeIP", func(t *testing.T) {
		extractor := NewKeyExtractor(KeyTypeIP, nil)
		key := extractor(ctx)
		if key != "192.168.1.1" {
			t.Errorf("expected '192.168.1.1', got %s", key)
		}
	})

	t.Run("KeyTypeUserID", func(t *testing.T) {
		extractor := NewKeyExtractor(KeyTypeUserID, nil)
		key := extractor(ctx)
		if key != "user123" {
			t.Errorf("expected 'user123', got %s", key)
		}
	})

	t.Run("KeyTypePath", func(t *testing.T) {
		extractor := NewKeyExtractor(KeyTypePath, nil)
		key := extractor(ctx)
		if key != "/api/test" {
			t.Errorf("expected '/api/test', got %s", key)
		}
	})

	t.Run("KeyTypeIPPath", func(t *testing.T) {
		extractor := NewKeyExtractor(KeyTypeIPPath, nil)
		key := extractor(ctx)
		if key != "192.168.1.1:/api/test" {
			t.Errorf("expected '192.168.1.1:/api/test', got %s", key)
		}
	})

	t.Run("KeyTypeUserPath", func(t *testing.T) {
		extractor := NewKeyExtractor(KeyTypeUserPath, nil)
		key := extractor(ctx)
		if key != "user123:/api/test" {
			t.Errorf("expected 'user123:/api/test', got %s", key)
		}
	})

	t.Run("KeyTypeCustom with extractor", func(t *testing.T) {
		customExtractor := func(ctx context.Context) string {
			return "custom_key"
		}
		extractor := NewKeyExtractor(KeyTypeCustom, customExtractor)
		key := extractor(ctx)
		if key != "custom_key" {
			t.Errorf("expected 'custom_key', got %s", key)
		}
	})

	t.Run("KeyTypeCustom without extractor", func(t *testing.T) {
		extractor := NewKeyExtractor(KeyTypeCustom, nil)
		key := extractor(ctx)
		if key != "192.168.1.1" {
			t.Errorf("expected fallback to IP, got %s", key)
		}
	})

	t.Run("unknown key type", func(t *testing.T) {
		extractor := NewKeyExtractor(KeyType("unknown"), nil)
		key := extractor(ctx)
		if key != "192.168.1.1" {
			t.Errorf("expected fallback to IP, got %s", key)
		}
	})
}

func TestCombineExtractors(t *testing.T) {
	t.Run("combine multiple extractors", func(t *testing.T) {
		ctx := WithValue(context.Background(), "client_ip", "192.168.1.1")
		ctx = WithValue(ctx, "user_id", "user123")
		ctx = WithValue(ctx, "request_path", "/api/test")

		extractor := CombineExtractors(ExtractByIP, ExtractByUserID, ExtractByPath)
		key := extractor(ctx)
		if key != "192.168.1.1:user123:/api/test" {
			t.Errorf("expected combined key, got %s", key)
		}
	})

	t.Run("with unknown values", func(t *testing.T) {
		ctx := context.Background()
		extractor := CombineExtractors(ExtractByIP, ExtractByUserID, ExtractByPath)
		key := extractor(ctx)
		if key != "default" {
			t.Errorf("expected 'default', got %s", key)
		}
	})

	t.Run("empty extractors", func(t *testing.T) {
		ctx := context.Background()
		extractor := CombineExtractors()
		key := extractor(ctx)
		if key != "default" {
			t.Errorf("expected 'default', got %s", key)
		}
	})

	t.Run("filters out unknown and anonymous", func(t *testing.T) {
		ctx := WithValue(context.Background(), "client_ip", "192.168.1.1")
		extractor := CombineExtractors(ExtractByIP, ExtractByUserID)
		key := extractor(ctx)
		if key != "192.168.1.1" {
			t.Errorf("expected '192.168.1.1', got %s", key)
		}
	})
}

func TestWithValue(t *testing.T) {
	ctx := context.Background()
	ctx = WithValue(ctx, "key1", "value1")
	ctx = WithValue(ctx, "key2", 123)

	if ctx.Value("key1") != "value1" {
		t.Error("key1 value mismatch")
	}
	if ctx.Value("key2") != 123 {
		t.Error("key2 value mismatch")
	}
}
