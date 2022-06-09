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
	"fmt"
	"time"

	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"

	"github.com/go-sql-driver/mysql"
)

type DSNFactory func(config.Config) (string, error)

var dsnDrivers = map[string]DSNFactory{
	"mysql":  MysqlDSNFactory,
	"sqlite": SqliteDSNFactory,
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

// MysqlDSNFactory get mysql dsn from config
func MysqlDSNFactory(conf config.Config) (string, error) {
	dsn := conf.GetString("dsn")
	if dsn != "" {
		return dsn, nil
	}

	name := conf.GetString("database")
	if name == "" {
		return "", errcode.New("database's name not exist, set param database")
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
	return dsnConf.FormatDSN(), nil
}

// SqliteDSNFactory get sqlite db from config
func SqliteDSNFactory(conf config.Config) (string, error) {
	dsn := conf.GetString("dsn")
	if dsn != "" {
		return dsn, nil
	}

	return "", errcode.New("database's name not exist, set param database")
}
