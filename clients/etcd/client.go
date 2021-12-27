package etcd

import (
	"crypto/tls"
	"flag"
	"net"
	"time"

	commonTls "trellis.tech/trellis/common.v1/crypto/tls"
	"trellis.tech/trellis/common.v1/flagext"
	"trellis.tech/trellis/common.v1/types"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ flagext.Parser = (*Config)(nil)

// Config for a new etcd.Client.
type Config struct {
	Endpoints   types.Strings    `yaml:"endpoints" json:"endpoints"`
	DialTimeout types.Duration   `yaml:"dial_timeout" json:"dial_timeout"`
	MaxRetries  int              `yaml:"max_retries" json:"max_retries"`
	EnableTLS   bool             `yaml:"tls_enabled" json:"enable_tls"`
	TLS         commonTls.Config `yaml:",inline"`
	Username    string           `yaml:"username" json:"username"`
	Password    types.Secret     `yaml:"password" json:"password"`
}

// Clientv3Facade is a subset of all Etcd client operations that are required
// to implement an Etcd version of kv.Client
type Clientv3Facade interface {
	clientv3.Cluster
	clientv3.KV
	clientv3.Watcher
	clientv3.Lease
}

func (cfg *Config) ParseFlags(f *flag.FlagSet) {
	cfg.ParseFlagsWithPrefix("", f)
}

// ParseFlagsWithPrefix adds the flags required to config this to the given FlagSet.
func (cfg *Config) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	cfg.Endpoints = []string{}
	f.Var(&cfg.Endpoints, prefix+"etcd.endpoints", "The etcd endpoints to connect to.")
	f.Var(&cfg.DialTimeout, prefix+"etcd.dial-timeout", "The dial timeout for the etcd connection.")
	f.IntVar(&cfg.MaxRetries, prefix+"etcd.max-retries", 10, "The maximum number of retries to do for failed ops.")
	f.BoolVar(&cfg.EnableTLS, prefix+"etcd.tls-enabled", false, "Enable TLS.")
	f.StringVar(&cfg.Username, prefix+"etcd.username", "", "Etcd username.")
	f.Var(&cfg.Password, prefix+"etcd.password", "Etcd password.")
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
		config.DialTimeout = time.Duration(cfg.DialTimeout)
	}

	config.Username = cfg.Username
	config.Password = string(cfg.Password)

	return clientv3.New(config)
}
