/*
Copyright © 2016 Henry Huang <hhh@rutcode.com>

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

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"

	"trellis.tech/trellis/common.v1/errcode"
	"trellis.tech/trellis/common.v1/flagext"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var _ flagext.Parser = (*Config)(nil)

// Config is the config for client TLS.
type Config struct {
	CertPath           string `yaml:"cert_path" json:"cert_path"`
	KeyPath            string `yaml:"key_path" json:"key_path"`
	CAPath             string `yaml:"ca_path" json:"ca_path"`
	ServerName         string `yaml:"server_name" json:"server_name"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

var (
	errKeyMissing  = errcode.New("certificate given but no key configured")
	errCertMissing = errcode.New("key given but no certificate configured")
)

// ParseFlags parse flags with prefix.
func (cfg *Config) ParseFlags(f *flag.FlagSet) {
	cfg.ParseFlagsWithPrefix("", f)
}

// ParseFlagsWithPrefix registers flags with prefix.
func (cfg *Config) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.StringVar(&cfg.CertPath, prefix+"tls-cert-path", "", "Path to the client certificate file, which will be used for authenticating with the server. Also requires the key path to be configured.")
	f.StringVar(&cfg.KeyPath, prefix+"tls-key-path", "", "Path to the key file for the client certificate. Also requires the client certificate to be configured.")
	f.StringVar(&cfg.CAPath, prefix+"tls-ca-path", "", "Path to the CA certificates file to validate server certificate against. If not set, the host's root CA certificates are used.")
	f.StringVar(&cfg.ServerName, prefix+"tls-server-name", "", "Override the expected name on the server certificate.")
	f.BoolVar(&cfg.InsecureSkipVerify, prefix+"tls-insecure-skip-verify", false, "Skip validating server certificate.")
}

// GetTLSConfig initialises tls.Config from config options
func (cfg *Config) GetTLSConfig() (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		ServerName:         cfg.ServerName,
	}

	// read ca certificates
	if cfg.CAPath != "" {
		var caCertPool *x509.CertPool
		caCert, err := ioutil.ReadFile(cfg.CAPath)
		if err != nil {
			return nil, errcode.NewErrors(err, errcode.Newf("error loading ca cert: %s", cfg.CAPath))
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		config.RootCAs = caCertPool
	}

	// read client certificate
	if cfg.CertPath != "" || cfg.KeyPath != "" {
		if cfg.CertPath == "" {
			return nil, errCertMissing
		}
		if cfg.KeyPath == "" {
			return nil, errKeyMissing
		}
		clientCert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, errcode.NewErrors(err,
				errcode.Newf("failed to load TLS certificate %s,%s", cfg.CertPath, cfg.KeyPath))
		}
		config.Certificates = []tls.Certificate{clientCert}
	}

	return config, nil
}

// GetGRPCDialOptions creates GRPC DialOptions for TLS
func (cfg *Config) GetGRPCDialOptions(enabled bool) ([]grpc.DialOption, error) {
	if !enabled {
		return []grpc.DialOption{grpc.WithInsecure()}, nil
	}

	tlsConfig, err := cfg.GetTLSConfig()
	if err != nil {
		return nil, errcode.NewErrors(err, errcode.Newf("error creating grpc dial options"))
	}

	return []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))}, nil
}
