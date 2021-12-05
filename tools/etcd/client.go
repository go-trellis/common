package etcd

import (
	"crypto/tls"
	"flag"
	"time"

	commonTls "trellis.tech/trellis/common.v0/crypto/tls"
	"trellis.tech/trellis/common.v0/errcode"
	"trellis.tech/trellis/common.v0/flagext"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config for a new etcd.Client.
type Config struct {
	Endpoints   []string               `yaml:"endpoints"`
	DialTimeout time.Duration          `yaml:"dial_timeout"`
	MaxRetries  int                    `yaml:"max_retries"`
	EnableTLS   bool                   `yaml:"tls_enabled"`
	TLS         commonTls.ClientConfig `yaml:",inline"`
	UserName    string                 `yaml:"username"`
	Password    string                 `yaml:"password"`
}

// Clientv3Facade is a subset of all Etcd client operations that are required
// to implement an Etcd version of kv.Client
type Clientv3Facade interface {
	clientv3.KV
	clientv3.Watcher
}

//// Client implements kv.Client for etcd.
//type Client struct {
//	cfg Config
//	//codec  codec.Codec
//	cli    Clientv3Facade
//	logger logger.Logger
//}

func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	cfg.ParseFlagsWithPrefix(f, "")
}

// ParseFlagsWithPrefix adds the flags required to config this to the given FlagSet.
func (cfg *Config) ParseFlagsWithPrefix(f *flag.FlagSet, prefix string) {
	cfg.Endpoints = []string{}
	f.Var((*flagext.StringSlice)(&cfg.Endpoints), prefix+"etcd.endpoints", "The etcd endpoints to connect to.")
	f.DurationVar(&cfg.DialTimeout, prefix+"etcd.dial-timeout", 10*time.Second, "The dial timeout for the etcd connection.")
	f.IntVar(&cfg.MaxRetries, prefix+"etcd.max-retries", 10, "The maximum number of retries to do for failed ops.")
	f.BoolVar(&cfg.EnableTLS, prefix+"etcd.tls-enabled", false, "Enable TLS.")
	f.StringVar(&cfg.UserName, prefix+"etcd.username", "", "Etcd username.")
	f.StringVar(&cfg.Password, prefix+"etcd.password", "", "Etcd password.")
	cfg.TLS.ParseFlagsWithPrefix(prefix+"etcd", f)
}

// GetTLS sets the TLS config field with certs
func (cfg *Config) GetTLS() (*tls.Config, error) {
	if !cfg.EnableTLS {
		return nil, nil
	}
	tlsInfo := &transport.TLSInfo{
		CertFile:           cfg.TLS.CertPath,
		KeyFile:            cfg.TLS.KeyPath,
		TrustedCAFile:      cfg.TLS.CAPath,
		ServerName:         cfg.TLS.ServerName,
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
	}
	return tlsInfo.ClientConfig()
}

// NewClient makes a new Client.
func NewClient(cfg Config) (Clientv3Facade, error) {
	tlsConfig, err := cfg.GetTLS()
	if err != nil {
		return nil, errcode.NewErrors(err, errcode.New("unable to initialise TLS configuration for etcd"))
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
		// Configure the keepalive to make sure that the client reconnects
		// to the etcd service endpoint(s) in case the current connection is
		// dead (ie. the node where etcd is running is dead or a network
		// partition occurs).
		//
		// The settings:
		// - DialKeepAliveTime: time before the client pings the server to
		//   see if transport is alive (10s hardcoded)
		// - DialKeepAliveTimeout: time the client waits for a response for
		//   the keep-alive probe (set to 2x dial timeout, in order to avoid
		//   exposing another config option which is likely to be a factor of
		//   the dial timeout anyway)
		// - PermitWithoutStream: whether the client should send keepalive pings
		//   to server without any active streams (enabled)
		DialKeepAliveTime:    10 * time.Second,
		DialKeepAliveTimeout: 2 * cfg.DialTimeout,
		PermitWithoutStream:  true,
		TLS:                  tlsConfig,
		Username:             cfg.UserName,
		Password:             cfg.Password,
	})
	if err != nil {
		return nil, err
	}

	return cli, nil
}
