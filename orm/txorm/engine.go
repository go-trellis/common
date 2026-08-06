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

	"github.com/go-trellis/common/config"
	"github.com/go-trellis/common/errors/errcode"
	"github.com/go-trellis/common/logger"
	"github.com/go-trellis/common/orm/transaction"

	"xorm.io/xorm"
	"xorm.io/xorm/core"
	"xorm.io/xorm/log"
)

var _ transaction.Engine = (*XEngine)(nil)

var (
	defaultOptions = &Options{
		maxIdleConns: 5,            // default max idle connections
		maxOpenConns: 10,           // default max open connections
		showSQL:      false,        // default show sql
		logLevel:     log.LOG_INFO, // default log level
		driver:       "mysql",      // default driver
	}
)

type XEngine struct {
	*xorm.Engine
}

type Option func(*Options)
type Options struct {
	logger             log.Logger
	driver, coreDriver string
	maxIdleConns       int
	maxOpenConns       int
	showSQL            bool
	logLevel           log.LogLevel
	isDefault          bool
}

func OptDriver(d string) Option {
	return func(o *Options) {
		o.driver = d
	}
}

func OptLogger(l log.Logger) Option {
	return func(o *Options) {
		o.logger = l
	}
}

func OptMaxIdleConns(maxIdleConns int) Option {
	return func(o *Options) {
		o.maxIdleConns = maxIdleConns
	}
}

func OptMaxOpenConns(maxOpenConns int) Option {
	return func(o *Options) {
		o.maxOpenConns = maxOpenConns
	}
}

func OptLogLevel(lv log.LogLevel) Option {
	return func(o *Options) {
		o.logLevel = lv
	}
}

func OptShowSQL(showSQL bool) Option {
	return func(o *Options) {
		o.showSQL = showSQL
	}
}

func OptIsDefault(def bool) Option {
	return func(o *Options) {
		o.isDefault = def
	}
}

// NewEnginesFromFile initial engines from file.
func NewEnginesFromFile(file string, l logger.Logger) (map[string]transaction.Engine, error) {
	conf, err := config.NewConfigOptions(config.OptionFile(file))
	if err != nil {
		return nil, err
	}
	return NewEnginesWithConfig(conf, l)
}

// NewEnginesWithConfig initial engines with config options.
func NewEnginesWithConfig(cfg config.Config, l logger.Logger) (engines map[string]transaction.Engine, err error) {
	if cfg == nil {
		return nil, errcode.New("nil config")
	}

	es := make(map[string]transaction.Engine)

	defer func() {
		if err == nil {
			return
		}
		for _, engine := range es {
			engine.Close()
		}
	}()

	for _, key := range cfg.GetKeys() {
		engine, isDefault, e := newXORMEngineWithConfig(cfg, key, l)
		if e != nil {
			return nil, e
		}
		xEngine, e := newXEngine(engine)
		if e != nil {
			return nil, e
		}
		if isDefault {
			es[transaction.DefaultDatabase] = xEngine
		}
		es[key] = xEngine
	}

	return es, nil
}

// NewXEngine new XEngine from driver and dsn.
func NewXEngine(driver, dsn string, ops ...Option) (*XEngine, error) {
	engine, err := NewXORMEngine(driver, dsn, ops...)
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

// NewXORMEngine New XORM Engine
func NewXORMEngine(driver, dsn string, ops ...Option) (*xorm.Engine, error) {
	return NewXORMEngineWithDB(driver, dsn, nil, ops...)
}

// NewXORMEngineWithDB New XORM Engine with DB
func NewXORMEngineWithDB(driver, dsn string, db *core.DB, ops ...Option) (*xorm.Engine, error) {
	options := &Options{}
	for _, o := range ops {
		o(options)
	}
	options.driver = driver
	return newXormEngine(dsn, options, db)
}

// NewXORMEnginesFromFile initial xorm engine from file
func NewXORMEnginesFromFile(file string, l logger.Logger) (map[string]*xorm.Engine, error) {
	conf, err := config.NewConfigOptions(config.OptionFile(file))
	if err != nil {
		return nil, err
	}
	return NewXORMEngineWithConfig(conf, l)
}

func NewXORMEngineWithConfig(cfg config.Config, l logger.Logger) (engines map[string]*xorm.Engine, err error) {
	if cfg == nil {
		return nil, errcode.New("nil config")
	}

	es := make(map[string]*xorm.Engine)
	defer func() {
		if err == nil {
			return
		}
		for _, engine := range es {
			engine.Close()
		}
	}()

	for _, key := range cfg.GetKeys() {
		engine, isDefault, e := newXORMEngineWithConfig(cfg, key, l)
		if e != nil {
			return nil, e
		}
		if isDefault {
			es[transaction.DefaultDatabase] = engine
		}
		es[key] = engine
	}
	return es, nil
}

func newXORMEngineWithConfig(cfg config.Config, key string, l log.Logger) (*xorm.Engine, bool, error) {
	dConfig := cfg.GetValuesConfig(key)
	if dConfig == nil {
		return nil, false, errcode.Newf("not found config with key: %s", key)
	}
	options := configureToOptions(dConfig)
	// Stash the injected logger on options so configureEngine sets it before
	// applying ShowSQL/log level; otherwise those settings land on the default
	// logger and are lost when the logger is later replaced.
	if l != nil {
		options.logger = l
	}

	f, err := transaction.GetDSNFactory(options.driver)
	if err != nil {
		return nil, false, err
	}
	dsn, err := f(dConfig)
	if err != nil {
		return nil, false, err
	}

	engine, err := newXormEngine(dsn, options, nil)
	if err != nil {
		return nil, false, err
	}

	return engine, options.isDefault, nil
}

// newXormEngine create a new xorm engine with given options and core database connection.
func newXormEngine(dsn string, options *Options, coreDB *core.DB) (*xorm.Engine, error) {
	// if options is nil, use default options.
	if options == nil {
		options = defaultOptions
	}
	var err error
	// coreDB is not nil, use it as the core database connection.
	if coreDB == nil && options.coreDriver != "" {
		coreDB, err = core.Open(options.coreDriver, dsn)
		if err != nil {
			return nil, err
		}
	}

	// coreDB is not nil, use it as the core database connection.
	var engine *xorm.Engine
	if coreDB != nil {
		engine, err = xorm.NewEngineWithDB(options.driver, dsn, coreDB)
	} else {
		engine, err = xorm.NewEngine(options.driver, dsn)
	}
	if err != nil {
		return nil, err
	}
	// set engine options.
	configureEngine(engine, options)

	return engine, nil
}

// configureEngine configure the engine with given options.
// - engine: xorm object.
// - options: options for engine.
func configureEngine(engine *xorm.Engine, options *Options) {
	engine.SetMaxIdleConns(options.maxIdleConns)
	engine.SetMaxOpenConns(options.maxOpenConns)
	// Set the custom logger first so ShowSQL/log level apply to it rather than
	// the default logger (SetLogger replaces the logger without carrying settings over).
	if options.logger != nil {
		engine.SetLogger(options.logger)
	}
	engine.ShowSQL(options.showSQL)
	engine.Logger().SetLevel(options.logLevel)
}

// configureToOptions configure the options from given config.
func configureToOptions(cfg config.Config) *Options {
	return &Options{
		maxIdleConns: cfg.GetInt("max_idle_conns", defaultOptions.maxIdleConns),
		maxOpenConns: cfg.GetInt("max_open_conns", defaultOptions.maxOpenConns),
		showSQL:      cfg.GetBoolean("show_sql"),
		logLevel:     log.LogLevel(cfg.GetInt("log_level")),
		isDefault:    cfg.GetBoolean("is_default"),
		driver:       cfg.GetString("driver", defaultOptions.driver),
		coreDriver:   cfg.GetString("core_driver", defaultOptions.coreDriver),
	}
}

func (p *XEngine) TransactionDo(fn func(*xorm.Session) error) error {
	return TransactionDo(p, fn)
}

func (p *XEngine) NewSession() (any, error) {
	return p.NewXORMSession()
}

func (p *XEngine) NewXORMSession() (*xorm.Session, error) {
	return p.Engine.NewSession(), nil
}

func (p *XEngine) Exec(sql string, args ...any) (sql.Result, error) {
	sqlOrArgs := append([]any{sql}, args...)
	return p.Engine.Exec(sqlOrArgs...)
}

func (p *XEngine) BeginTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: true, engine: p.Engine, session: p.Engine.NewSession()}, nil
}

func (p *XEngine) BeginNonTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: false, engine: p.Engine}, nil
}
