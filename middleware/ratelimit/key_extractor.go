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
	"fmt"
	"net"
	"strings"
)

// WithValue adds a value to context for key extraction
func WithValue(ctx context.Context, key, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// KeyType defines the type of key to extract
type KeyType string

const (
	KeyTypeIP       KeyType = "ip"        // Extract from IP address
	KeyTypeUserID   KeyType = "user_id"   // Extract from user ID in context
	KeyTypePath     KeyType = "path"      // Extract from request path
	KeyTypeIPPath   KeyType = "ip_path"   // Combine IP and path
	KeyTypeUserPath KeyType = "user_path" // Combine user ID and path
	KeyTypeCustom   KeyType = "custom"    // Custom extractor function
)

// ExtractByIP extracts key from IP address
func ExtractByIP(ctx context.Context) string {
	if ip := ctx.Value("client_ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return "unknown"
}

// ExtractByUserID extracts key from user ID in context
func ExtractByUserID(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		return fmt.Sprintf("%v", userID)
	}
	return "anonymous"
}

// ExtractByPath extracts key from request path
func ExtractByPath(ctx context.Context) string {
	if path := ctx.Value("request_path"); path != nil {
		if pathStr, ok := path.(string); ok {
			return pathStr
		}
	}
	return "unknown"
}

// ExtractByIPPath combines IP and path
func ExtractByIPPath(ctx context.Context) string {
	ip := ExtractByIP(ctx)
	path := ExtractByPath(ctx)
	return fmt.Sprintf("%s:%s", ip, path)
}

// ExtractByUserPath combines user ID and path
func ExtractByUserPath(ctx context.Context) string {
	userID := ExtractByUserID(ctx)
	path := ExtractByPath(ctx)
	return fmt.Sprintf("%s:%s", userID, path)
}

// ExtractIPFromAddr extracts IP from network address
func ExtractIPFromAddr(addr net.Addr) string {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	case *net.TCPAddr:
		ip = v.IP
	case *net.UDPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return "127.0.0.1"
	}
	ip = ip.To4()
	if ip == nil {
		return "unknown"
	}
	return ip.String()
}

// NewKeyExtractor creates a key extractor based on key type
func NewKeyExtractor(keyType KeyType, customExtractor KeyExtractor) KeyExtractor {
	switch keyType {
	case KeyTypeIP:
		return ExtractByIP
	case KeyTypeUserID:
		return ExtractByUserID
	case KeyTypePath:
		return ExtractByPath
	case KeyTypeIPPath:
		return ExtractByIPPath
	case KeyTypeUserPath:
		return ExtractByUserPath
	case KeyTypeCustom:
		if customExtractor != nil {
			return customExtractor
		}
		return ExtractByIP // fallback to IP
	default:
		return ExtractByIP
	}
}

// CombineExtractors combines multiple extractors into one
func CombineExtractors(extractors ...KeyExtractor) KeyExtractor {
	return func(ctx context.Context) string {
		var keys []string
		for _, extractor := range extractors {
			key := extractor(ctx)
			if key != "" && key != "unknown" && key != "anonymous" {
				keys = append(keys, key)
			}
		}
		if len(keys) == 0 {
			return "default"
		}
		return strings.Join(keys, ":")
	}
}
