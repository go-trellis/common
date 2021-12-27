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
	"sync"

	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"

	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var locker = &sync.Mutex{}

// NewEnginesFromFile initial xorm engine from file
func NewEnginesFromFile(file string) (map[string]*xorm.Engine, error) {
	conf, err := config.NewConfigOptions(config.OptionFile(file))
	if err != nil {
		return nil, err
	}
	return NewEnginesFromConfig(conf)
}

// NewEnginesFromConfig initial xorm engine from config
func NewEnginesFromConfig(conf config.Config) (engines map[string]*xorm.Engine, err error) {

	es := make(map[string]*xorm.Engine)
	if conf == nil {
		return nil, fmt.Errorf("config is nil")
	}

	locker.Lock()
	defer locker.Unlock()

	for _, key := range conf.GetKeys() {
		var _engine *xorm.Engine
		databaseConf := conf.GetValuesConfig(key)
		switch driver := databaseConf.GetString("driver", "mysql"); driver {
		case driver:
			databaseName := databaseConf.GetString("database", key)
			if databaseName == "" {
				return nil, errcode.Newf("not found database: %s", key)
			}

			_engine, err = newEngine(driver, GetMysqlDSNFromConfig(databaseName, databaseConf))
			if err != nil {
				return nil, err
			}

			_engine.SetMaxIdleConns(conf.GetInt(key+".max_idle_conns", 10))
			_engine.SetMaxOpenConns(conf.GetInt(key+".max_open_conns", 100))
			_engine.ShowSQL(conf.GetBoolean(key + ".show_sql"))

		default:
			return nil, fmt.Errorf("unsupported driver: %s", driver)
		}

		_engine.Logger().SetLevel(log.LogLevel(conf.GetInt(key+".log_level", 0)))

		if _isD := conf.GetBoolean(key + ".is_default"); _isD {
			es[DefaultDatabase] = _engine
		}

		es[key] = _engine
	}

	return es, nil
}

// NewEnginesFromDSN initial xorm engine by dsn
func NewEnginesFromDSN(driver, dsn string) (*xorm.Engine, error) {
	return newEngine(driver, dsn)
}

func newEngine(driver, dsn string) (*xorm.Engine, error) {
	return xorm.NewEngine(driver, dsn)
}
