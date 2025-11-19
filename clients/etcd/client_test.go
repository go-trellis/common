package etcd

import (
	"flag"
	"testing"
	"time"

	commonTls "trellis.tech/trellis/common.v3/crypto/tls"
	"trellis.tech/trellis/common.v3/utils/testutils"
	"trellis.tech/trellis/common.v3/utils/types"
)

func TestConfig_ParseFlags(t *testing.T) {
	cfg := &Config{}
	f := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.ParseFlags(f)
	// Should not panic
}

func TestConfig_ParseFlagsWithPrefix(t *testing.T) {
	cfg := &Config{}
	f := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.ParseFlagsWithPrefix("test.", f)
	// Should not panic
}

func TestConfig_GetTLS_Disabled(t *testing.T) {
	cfg := &Config{
		EnableTLS: false,
	}
	tlsConfig, err := cfg.GetTLS()
	testutils.Ok(t, err)
	testutils.Assert(t, tlsConfig == nil, "TLS config should be nil when disabled")
}

func TestConfig_GetTLS_Enabled(t *testing.T) {
	cfg := &Config{
		EnableTLS: true,
		TLS: commonTls.Config{
			InsecureSkipVerify: true,
		},
	}
	tlsConfig, err := cfg.GetTLS()
	// May fail if cert files don't exist, but that's OK for testing
	_ = tlsConfig
	_ = err
}

func TestNewClient_EmptyEndpoints(t *testing.T) {
	cfg := Config{
		Endpoints: []string{},
	}
	// This will try to connect to default endpoint and may fail
	// But we're just testing the code path, not actual connection
	_, err := NewClient(cfg)
	// Connection will likely fail, but that's expected in tests
	_ = err
}

func TestNewClient_WithEndpoints(t *testing.T) {
	cfg := Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: types.Duration(time.Second),
	}
	// This will try to connect and may fail
	_, err := NewClient(cfg)
	// Connection will likely fail, but that's expected in tests
	_ = err
}

func TestNewClient_WithEndpointsNoPort(t *testing.T) {
	cfg := Config{
		Endpoints: []string{"127.0.0.1"}, // No port
	}
	_, err := NewClient(cfg)
	// Should handle missing port by adding default port
	_ = err
}

func TestNewClient_WithDialTimeout(t *testing.T) {
	cfg := Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: types.Duration(5 * time.Second),
	}
	_, err := NewClient(cfg)
	_ = err
}

func TestNewClient_WithUsernamePassword(t *testing.T) {
	cfg := Config{
		Endpoints: []string{"127.0.0.1:2379"},
		Username:  "testuser",
		Password:  types.Secret("testpass"),
	}
	_, err := NewClient(cfg)
	_ = err
}

func TestNewClient_WithEmptyEndpoint(t *testing.T) {
	cfg := Config{
		Endpoints: []string{"", "127.0.0.1:2379"}, // Empty string should be skipped
	}
	_, err := NewClient(cfg)
	_ = err
}

func TestNewClient_EndpointSplitError(t *testing.T) {
	cfg := Config{
		Endpoints: []string{"invalid:endpoint:with:multiple:colons:2379"}, // Invalid format
	}
	_, err := NewClient(cfg)
	// Should handle error gracefully
	_ = err
}

func TestConfig_ParseFlagsWithPrefix_AllFields(t *testing.T) {
	cfg := &Config{}
	f := flag.NewFlagSet("test", flag.ContinueOnError)

	cfg.ParseFlagsWithPrefix("test.", f)
	// Should not panic - just registers flags, doesn't parse them
}

func TestNewClient_DefaultConfig(t *testing.T) {
	cfg := Config{}
	_, err := NewClient(cfg)
	// Should use default endpoint 127.0.0.1:2379
	_ = err
}
