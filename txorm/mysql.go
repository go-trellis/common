/*
Copyright © 2019 Henry Huang <hhh@rutcode.com>

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

package txorm

import (
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"trellis.tech/trellis/common.v0.1"
	"trellis.tech/trellis/common.v0.1/config"
)

// DefaultDatabase default database key
var (
	DefaultDatabase = common.FormatNamespaceString("xorm_ext:database")
)

// GetMysqlDSNFromConfig get mysql dsn from gogap config
func GetMysqlDSNFromConfig(name string, conf config.Config) string {
	if name == "" {
		panic("database's name not exist")
	}

	dsn := conf.GetString("dsn")
	if dsn != "" {
		return dsn
	}

	dsnConf := mysql.Config{
		DBName:  name,
		Net:     "tcp",
		Timeout: conf.GetTimeDuration("timeout", time.Second*5),

		User:   conf.GetString("user", "root"),
		Passwd: conf.GetString("password", ""),
		Addr:   fmt.Sprintf("%s:%d", conf.GetString("host", "localhost"), conf.GetInt("port", 3306)),
		Params: map[string]string{
			"charset":              conf.GetString("charset", "utf8"),
			"parseTime":            conf.GetString("parseTime", "True"),
			"loc":                  conf.GetString("location", "Local"),
			"allowNativePasswords": conf.GetString("allowNativePasswords", "true"),
		},
	}
	return dsnConf.FormatDSN()
}
