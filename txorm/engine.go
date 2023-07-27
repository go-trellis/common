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

package txorm

import (
	"database/sql"
	"fmt"
	"sync"

	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"
	"trellis.tech/trellis/common.v1/logger"
	"trellis.tech/trellis/common.v1/transaction"

	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var _ transaction.Engine = (*XEngine)(nil)

type XEngine struct {
	*xorm.Engine
}

var locker = &sync.Mutex{}

type Option func(*Options)
type Options struct {
	logger logger.Logger
}

func OptLogger(l logger.Logger) Option {
	return func(o *Options) {
		o.logger = l
	}
}

// NewEnginesFromFile initial engines from file
func NewEnginesFromFile(file string) (map[string]transaction.Engine, error) {
	conf, err := config.NewConfigOptions(config.OptionFile(file))
	if err != nil {
		return nil, err
	}
	return NewEnginesFromConfig(conf)
}

// NewEnginesFromConfig initial engines from config
func NewEnginesFromConfig(cfg config.Config, opts ...Option) (engines map[string]transaction.Engine, err error) {
	if cfg == nil {
		return nil, errcode.New("nil config")
	}

	locker.Lock()
	defer locker.Unlock()
	es := make(map[string]transaction.Engine)

	defer func() {
		if err == nil {
			return
		}
		for _, engine := range es {
			engine.Close()
		}
	}()

	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	for _, key := range cfg.GetKeys() {
		engine, err := genXormEngine(cfg, key, options.logger)
		if err != nil {
			return nil, err
		}
		xEngine, err := newXEngine(engine)
		if err != nil {
			return nil, err
		}
		if _isD := cfg.GetBoolean(key + ".is_default"); _isD {
			es[transaction.DefaultDatabase] = xEngine
		}

		es[key] = xEngine
	}

	return es, nil
}

// NewEngine New XEngine
// Deprecated: Use NewXEngine
func NewEngine(driver string, dsn string) (*XEngine, error) {
	engine, err := NewXORMEngine(driver, dsn)
	if err != nil {
		return nil, err
	}
	x := &XEngine{
		Engine: engine,
	}
	return x, nil
}

func NewXORMEngine(driver string, dsn string) (*xorm.Engine, error) {
	return xorm.NewEngine(driver, dsn)
}

// NewXORMEnginesFromFile initial xorm engine from file
func NewXORMEnginesFromFile(file string) (map[string]*xorm.Engine, error) {
	conf, err := config.NewConfigOptions(config.OptionFile(file))
	if err != nil {
		return nil, err
	}
	return NewXORMEngines(conf, nil)
}

func NewXORMEngines(cfg config.Config, opts ...Option) (engines map[string]*xorm.Engine, err error) {
	engines = make(map[string]*xorm.Engine, 0)

	locker.Lock()
	defer locker.Unlock()
	defer func() {
		if err == nil {
			return
		}
		for _, engine := range engines {
			engine.Close()
		}
	}()

	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	for _, key := range cfg.GetKeys() {
		engine, err := genXormEngine(cfg, key, options.logger)
		if err != nil {
			return nil, err
		}
		if _isD := cfg.GetBoolean(key + ".is_default"); _isD {
			engines[transaction.DefaultDatabase] = engine
		}
		engines[key] = engine
	}
	return engines, nil
}

func genXormEngine(cfg config.Config, key string, logger logger.Logger) (*xorm.Engine, error) {
	dConfig := cfg.GetValuesConfig(key)
	if dConfig == nil {
		return nil, errcode.Newf("not found config with key: %s", key)
	}
	driver := dConfig.GetString("driver", "mysql")

	f, err := transaction.GetDSNFactory(driver)
	if err != nil {
		return nil, err
	}

	dsn, err := f(dConfig)
	if err != nil {
		return nil, err
	}

	engine, err := NewXORMEngine(driver, dsn)
	if err != nil {
		return nil, err
	}

	engine.SetMaxIdleConns(cfg.GetInt(key+".max_idle_conns", 10))
	engine.SetMaxOpenConns(cfg.GetInt(key+".max_open_conns", 100))

	if logger != nil {
		engine.SetLogger(logger)
	}

	engine.ShowSQL(cfg.GetBoolean(key + ".show_sql"))
	engine.Logger().SetLevel(log.LogLevel(cfg.GetInt(key+".log_level", 0)))
	return engine, nil
}

func NewXEngine(driver string, dsn string) (*XEngine, error) {
	engine, err := NewXORMEngine(driver, dsn)
	if err != nil {
		return nil, err
	}

	return newXEngine(engine)
}

func newXEngine(engine *xorm.Engine) (*XEngine, error) {
	if engine == nil {
		return nil, fmt.Errorf("nil engine")
	}
	x := &XEngine{
		Engine: engine,
	}
	return x, nil
}

func (p *XEngine) TransactionDo(fn func(*xorm.Session) error) error {
	return TransactionDoWithSession(p.Engine.NewSession(), fn)
}

func (p *XEngine) NewSession() (interface{}, error) {
	return p.NewXORMSession()
}

func (p *XEngine) NewXORMSession() (*xorm.Session, error) {
	return p.Engine.NewSession(), nil
}

func (p *XEngine) Exec(sql string, args ...interface{}) (sql.Result, error) {
	sqlOrArgs := append([]interface{}{sql}, args...)
	return p.Engine.Exec(sqlOrArgs...)
}

func (p *XEngine) BeginTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: true, engine: p.Engine, session: p.Engine.NewSession()}, nil
}

func (p *XEngine) BeginNonTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: false, engine: p.Engine}, nil
}
