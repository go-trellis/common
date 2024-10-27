/*
Copyright © 2022 Henry Huang <hhh@rutcode.com>

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

package transaction

import (
	"crypto/tls"
	"fmt"
	"time"

	"trellis.tech/trellis/common.v2/config"
	"trellis.tech/trellis/common.v2/errcode"

	"github.com/go-sql-driver/mysql"
)

type DSNFactory func(config.Config) (string, error)

var dsnDrivers = map[string]DSNFactory{
	"mysql":   MysqlDSNFactory,
	"sqlite3": Sqlite3DSNFactory,
}

func SetDSNFactory(name string, factory DSNFactory) error {
	if name == "" {
		return errcode.New("name must not be empty")
	}
	if factory == nil {
		return errcode.New("nil factory")
	}
	dsnDrivers[name] = factory
	return nil
}

func GetDSNFactory(name string) (DSNFactory, error) {
	if name == "" {
		return nil, errcode.New("name must not be empty")
	}
	dsnF, ok := dsnDrivers[name]
	if !ok {
		return nil, errcode.Newf("not found driver: %s", name)
	}
	return dsnF, nil
}

// RegisterTLSConfig registers a custom tls.Config to be used with sql.Open.
// Use the key as a value in the DSN where tls=value.
func RegisterTLSConfig(key string, config *tls.Config) error {
	return mysql.RegisterTLSConfig(key, config)
}

// MysqlDSNFactory get mysql dsn from config
func MysqlDSNFactory(conf config.Config) (string, error) {
	if dsn := conf.GetString("dsn"); dsn != "" {
		return dsn, nil
	}

	name := conf.GetString("database")
	if name == "" {
		return "", errcode.New("database's name not exist, set param database")
	}

	dsnConf := mysql.Config{
		DBName:       name,
		Net:          conf.GetString("net", "tcp"),
		Timeout:      conf.GetTimeDuration("timeout", time.Second*5),
		ReadTimeout:  conf.GetTimeDuration("write_timeout", time.Minute),
		WriteTimeout: conf.GetTimeDuration("write_timeout", time.Second*30),

		User:   conf.GetString("user", "root"),
		Passwd: conf.GetString("password", ""),
		Addr:   fmt.Sprintf("%s:%d", conf.GetString("host", "localhost"), conf.GetInt("port", 3306)),
		Params: map[string]string{
			"loc":     conf.GetString("location", "Local"),
			"charset": conf.GetString("charset", "utf8"),
		},
		// Max packet size allowed
		MaxAllowedPacket: conf.GetInt("max_allowed_packet", 0),
		// Server public key name
		ServerPubKey: conf.GetString("server_pub_key"),
		// TLS configuration name
		TLSConfig: conf.GetString("tls_config"),
		// Connection collation
		Collation: conf.GetString("collation"),
		// Allow all files to be used with LOAD DATA LOCAL INFILE
		AllowAllFiles: conf.GetBoolean("allowAllFiles", false),
		// Allows the cleartext client side plugin
		AllowCleartextPasswords: conf.GetBoolean("allowCleartextPasswords", false),
		// Allows fallback to unencrypted connection if server does not support TLS
		AllowFallbackToPlaintext: conf.GetBoolean("allowFallbackToPlaintext", false),
		// Allows the native password authentication method
		AllowNativePasswords: conf.GetBoolean("allowNativePasswords", false),
		// Allows the old insecure password method
		AllowOldPasswords: conf.GetBoolean("allowOldPasswords", false),
		CheckConnLiveness: conf.GetBoolean("checkConnLiveness", false),
		// Return number of matching rows instead of rows changed
		ClientFoundRows: conf.GetBoolean("clientFoundRows", false),
		// Prepend table alias to column names
		ColumnsWithAlias: conf.GetBoolean("columnsWithAlias", false),
		// Interpolate placeholders into query string
		InterpolateParams: conf.GetBoolean("interpolateParams", false),
		// Allow multiple statements in one query
		MultiStatements: conf.GetBoolean("multiStatements", false),
		// Parse time values to time.Time
		ParseTime: conf.GetBoolean("parseTime", false),
		// RejectReadOnly           bool // Reject read-only connections
		RejectReadOnly: conf.GetBoolean("rejectReadOnly", false),
	}
	return dsnConf.FormatDSN(), nil
}

// Sqlite3DSNFactory get sqlite3 db from config
func Sqlite3DSNFactory(conf config.Config) (string, error) {
	dsn := conf.GetString("dsn")
	if dsn != "" {
		return dsn, nil
	}
	return "", errcode.New("database's name not exist, set param database")
}
