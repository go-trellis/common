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
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"trellis.tech/trellis/common.v3/errcode"
	"trellis.tech/trellis/common.v3/testutils"
)

func TestConfig_ParseFlags(t *testing.T) {
	cfg := &Config{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	cfg.ParseFlags(fs)

	err := fs.Parse([]string{"-tls-cert-path", "/path/to/cert", "-tls-key-path", "/path/to/key"})
	testutils.Ok(t, err)
	testutils.Equals(t, "/path/to/cert", cfg.CertPath, "CertPath should be set")
	testutils.Equals(t, "/path/to/key", cfg.KeyPath, "KeyPath should be set")
}

func TestConfig_ParseFlagsWithPrefix(t *testing.T) {
	cfg := &Config{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	cfg.ParseFlagsWithPrefix("prefix.", fs)

	err := fs.Parse([]string{"-prefix.tls-cert-path", "/path/to/cert", "-prefix.tls-insecure-skip-verify", "true"})
	testutils.Ok(t, err)
	testutils.Equals(t, "/path/to/cert", cfg.CertPath, "CertPath should be set with prefix")
	testutils.Assert(t, cfg.InsecureSkipVerify, "InsecureSkipVerify should be set")
}

func TestConfig_GetTLSConfig_NoTLS(t *testing.T) {
	cfg := &Config{}

	config, err := cfg.GetTLSConfig()
	testutils.Ok(t, err)
	testutils.Assert(t, config != nil, "GetTLSConfig should return non-nil config")
}

func TestConfig_GetTLSConfig_InsecureSkipVerify(t *testing.T) {
	cfg := &Config{
		InsecureSkipVerify: true,
		ServerName:         "test.example.com",
	}

	config, err := cfg.GetTLSConfig()
	testutils.Ok(t, err)
	testutils.Assert(t, config != nil, "GetTLSConfig should return non-nil config")
	testutils.Assert(t, config.InsecureSkipVerify, "InsecureSkipVerify should be set")
	testutils.Equals(t, "test.example.com", config.ServerName, "ServerName should be set")
}

func TestConfig_GetTLSConfig_CertWithoutKey(t *testing.T) {
	cfg := &Config{
		CertPath: "/path/to/cert",
		KeyPath:  "",
	}

	_, err := cfg.GetTLSConfig()
	testutils.NotOk(t, err, "should return error when cert without key")
}

func TestConfig_GetTLSConfig_KeyWithoutCert(t *testing.T) {
	cfg := &Config{
		CertPath: "",
		KeyPath:  "/path/to/key",
	}

	_, err := cfg.GetTLSConfig()
	testutils.NotOk(t, err, "should return error when key without cert")
}

func TestConfig_GetTLSConfig_InvalidCertPath(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "nonexistent.crt")
	keyPath := filepath.Join(tmpDir, "nonexistent.key")

	cfg := &Config{
		CertPath: certPath,
		KeyPath:  keyPath,
	}

	_, err := cfg.GetTLSConfig()
	testutils.NotOk(t, err, "should return error for invalid cert path")
}

func TestConfig_GetTLSConfig_InvalidCAPath(t *testing.T) {
	tmpDir := t.TempDir()
	caPath := filepath.Join(tmpDir, "nonexistent.ca")

	cfg := &Config{
		CAPath: caPath,
	}

	_, err := cfg.GetTLSConfig()
	testutils.NotOk(t, err, "should return error for invalid ca path")
}

func TestConfig_GetTLSConfig_ValidCAPath(t *testing.T) {
	tmpDir := t.TempDir()
	caPath := filepath.Join(tmpDir, "ca.crt")
	err := os.WriteFile(caPath, []byte("-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----\n"), 0644)
	testutils.Ok(t, err)

	cfg := &Config{
		CAPath: caPath,
	}

	config, err := cfg.GetTLSConfig()
	testutils.Ok(t, err)
	testutils.Assert(t, config != nil, "GetTLSConfig should return non-nil config")
	testutils.Assert(t, config.RootCAs != nil, "RootCAs should be set")
}

func TestConfig_UnmarshalYAML(t *testing.T) {
	cfg := &Config{}
	unmarshal := func(v interface{}) error {
		// UnmarshalYAML passes (*plain)(cfg) to unmarshal function
		// Use reflect to access and modify the fields since we can't directly type assert
		rv := reflect.ValueOf(v).Elem()
		rv.FieldByName("CertPath").SetString("/test/cert")
		rv.FieldByName("KeyPath").SetString("/test/key")
		return nil
	}

	err := cfg.UnmarshalYAML(unmarshal)
	testutils.Ok(t, err)
	testutils.Equals(t, "/test/cert", cfg.CertPath, "CertPath should be set")
	testutils.Equals(t, "/test/key", cfg.KeyPath, "KeyPath should be set")
}

func TestConfig_UnmarshalYAML_Error(t *testing.T) {
	cfg := &Config{}
	unmarshal := func(v interface{}) error {
		return errcode.New("unmarshal error")
	}

	err := cfg.UnmarshalYAML(unmarshal)
	testutils.NotOk(t, err, "should return error when unmarshal fails")
}
