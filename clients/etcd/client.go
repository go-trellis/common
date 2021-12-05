package etcd

import (
	"crypto/tls"
	"flag"
	"net"
	"time"

	commonTls "trellis.tech/trellis/common.v0/crypto/tls"
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
	Username    string                 `yaml:"username"`
	Password    string                 `yaml:"password"`
}

// Clientv3Facade is a subset of all Etcd client operations that are required
// to implement an Etcd version of kv.Client
type Clientv3Facade interface {
	clientv3.KV
	clientv3.Watcher
}

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
	f.StringVar(&cfg.Username, prefix+"etcd.username", "", "Etcd username.")
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
	config := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	tlsConfig, err := cfg.GetTLS()
	if err != nil {
		return nil, err
	}
	config.TLS = tlsConfig

	var endpoints []string
	for _, endpoint := range cfg.Endpoints {
		if len(endpoint) == 0 {
			continue
		}
		host, port, err := net.SplitHostPort(endpoint)
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "2379"
			host = endpoint
			endpoints = append(endpoints, net.JoinHostPort(host, port))
		} else if err == nil {
			endpoints = append(endpoints, net.JoinHostPort(host, port))
		}
	}

	// if we got endpoints then we'll update
	if len(endpoints) > 0 {
		config.Endpoints = endpoints
	}

	if cfg.DialTimeout != 0 {
		config.DialTimeout = cfg.DialTimeout
	}

	config.Username = cfg.Username
	config.Password = cfg.Password

	return clientv3.New(config)
}
