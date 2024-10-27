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

	"trellis.tech/trellis/common.v2/config"
	"trellis.tech/trellis/common.v2/errcode"
	"trellis.tech/trellis/common.v2/logger"
	"trellis.tech/trellis/common.v2/transaction"

	"xorm.io/xorm"
	"xorm.io/xorm/core"
	"xorm.io/xorm/log"
)

var _ transaction.Engine = (*XEngine)(nil)

var (
	defaultOptions = &Options{
		maxIdleConns: 5,
		maxOpenConns: 10,
		showSQL:      false,
		logLevel:     log.LOG_INFO,
		driver:       "mysql",
	}
)

type XEngine struct {
	*xorm.Engine
}

var locker = &sync.Mutex{}

type Option func(*Options)
type Options struct {
	logger             logger.XormLogger
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

func OptLogger(l logger.XormLogger) Option {
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

// NewEnginesFromFile initial engines from file
func NewEnginesFromFile(file string, l logger.Logger) (map[string]transaction.Engine, error) {
	conf, err := config.NewConfigOptions(config.OptionFile(file))
	if err != nil {
		return nil, err
	}
	return NewEnginesWithConfig(conf, l)
}

// NewEnginesWithConfig initial engines from config
func NewEnginesWithConfig(cfg config.Config, l logger.Logger) (engines map[string]transaction.Engine, err error) {
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

	for _, key := range cfg.GetKeys() {
		engine, isDefault, err := newXORMEngineWithConfig(cfg, key, l)
		if err != nil {
			return nil, err
		}
		xEngine, err := newXEngine(engine)
		if err != nil {
			return nil, err
		}
		if isDefault {
			es[transaction.DefaultDatabase] = xEngine
		}
		es[key] = xEngine
	}

	return es, nil
}

// NewEngine New XEngine
// Deprecated: Use NewXEngine
func NewEngine(driver, dsn string, ops ...Option) (*XEngine, error) {
	return NewXEngine(driver, dsn, ops...)
}

// NewXEngine New XEngine
//   - error: 如果创建失败，返回错误信息。
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

func NewXORMEngine(driver, dsn string, ops ...Option) (*xorm.Engine, error) {
	return NewXORMEngineWithDB(driver, dsn, nil, ops...)
}

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

	for _, key := range cfg.GetKeys() {
		engine, isDefault, err := newXORMEngineWithConfig(cfg, key, l)
		if err != nil {
			return nil, err
		}
		if isDefault {
			engines[transaction.DefaultDatabase] = engine
		}
		engines[key] = engine
	}
	return engines, nil
}

func newXORMEngineWithConfig(cfg config.Config, key string, l log.Logger) (*xorm.Engine, bool, error) {
	dConfig := cfg.GetValuesConfig(key)
	if dConfig == nil {
		return nil, false, errcode.Newf("not found config with key: %s", key)
	}
	options := configureToOptions(dConfig)

	f, err := transaction.GetDSNFactory(options.driver)
	if err != nil {
		return nil, false, err
	}
	dsn, err := f(dConfig)
	if err != nil {
		return nil, false, err
	}

	options.logger = l

	engine, err := newXormEngine(dsn, options, nil)
	if err != nil {
		return nil, false, err
	}
	return engine, options.isDefault, nil
}

// newXormEngine 创建并返回一个新的 xorm.Engine 实例
// 参数:
// - driver: 数据库驱动名称
// - coreDriver: 核心数据库驱动名称
// - dsn: 数据库连接字符串
// - options: 配置选项
func newXormEngine(dsn string, options *Options, coreDB *core.DB) (*xorm.Engine, error) {
	// 如果 options 为 nil，则使用默认配置
	if options == nil {
		options = defaultOptions
	}
	var err error
	// 如果 coreDriver 不为空，则打开核心数据库连接
	if coreDB == nil && options.coreDriver != "" {
		coreDB, err = core.Open(options.coreDriver, dsn)
		if err != nil {
			return nil, err
		}
	}

	// 创建 XORM 引擎
	var engine *xorm.Engine
	if coreDB != nil {
		engine, err = xorm.NewEngineWithDB(options.driver, dsn, coreDB)
	} else {
		engine, err = xorm.NewEngine(options.driver, dsn)
	}
	if err != nil {
		return nil, err
	}
	// 设置引擎的连接池配置
	configureEngine(engine, options)

	return engine, nil
}

// configureEngine 配置 xorm.Engine 的连接池和日志选项
// 参数:
// - engine: xorm.Engine 实例
// - options: 配置选项
func configureEngine(engine *xorm.Engine, options *Options) {
	engine.SetMaxIdleConns(options.maxIdleConns)
	engine.SetMaxOpenConns(options.maxOpenConns)
	engine.ShowSQL(options.showSQL)
	// 如果提供了自定义日志器，则设置自定义日志器
	if options.logger != nil {
		engine.SetLogger(options.logger)
	}
	engine.Logger().SetLevel(options.logLevel)
}

// - *Options: 配置生成的 Options 对象
func configureToOptions(cfg config.Config) *Options {
	return &Options{
		maxIdleConns: cfg.GetInt("max_idle_conns", defaultOptions.maxIdleConns),
		maxOpenConns: cfg.GetInt("max_open_conns", defaultOptions.maxOpenConns),
		showSQL:      cfg.GetBoolean("show_sql"),
		logLevel:     log.LogLevel(cfg.GetInt("log_level")),
		isDefault:    cfg.GetBoolean("is_default"),
		driver:       cfg.GetString("driver", defaultOptions.driver),
		coreDriver:   cfg.GetString("core_driver"),
	}
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
