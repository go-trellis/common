package tgorm

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"gorm.io/gorm/logger"
	"trellis.tech/trellis/common.v1/config"
	"trellis.tech/trellis/common.v1/errcode"
	"trellis.tech/trellis/common.v1/transaction"
)

var locker = &sync.Mutex{}

type DSNFactory func(config.Config) (string, error)

var driverDSN = map[string]DSNFactory{
	"mysql": transaction.GetMysqlDSNFromConfig,
}

func SetDSNFactory(name string, factory DSNFactory) error {
	if name == "" {
		return errcode.New("name must not be empty")
	}
	if factory == nil {
		return errcode.New("nil factory")
	}
	locker.Lock()
	defer locker.Unlock()
	driverDSN[name] = factory
	return nil
}

// NewEnginesFromFile initial xorm engine from file
func NewEnginesFromFile(file string) (map[string]transaction.Engine, error) {
	conf, err := config.NewConfigOptions(config.OptionFile(file))
	if err != nil {
		return nil, err
	}
	return NewEnginesFromConfig(conf)
}

// NewEnginesFromConfig initial xorm engine from config
func NewEnginesFromConfig(conf config.Config) (engines map[string]transaction.Engine, err error) {

	if conf == nil {
		return nil, fmt.Errorf("nil config")
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

	for _, key := range conf.GetKeys() {
		databaseConf := conf.GetValuesConfig(key)
		driver := databaseConf.GetString("driver", "mysql")

		f, ok := driverDSN[driver]
		if !ok {
			return nil, fmt.Errorf("unsupported driver: %s", driver)
		}

		dsn, err := f(databaseConf)
		if err != nil {
			return nil, err
		}

		logger.Default.LogMode(logger.LogLevel(conf.GetInt(key+".log_level", 0)))

		xEngine, err := NewEngine(driver, dsn, &gorm.Config{
			Logger: logger.Default,
		})
		if err != nil {
			return nil, err
		}

		sqlDB, err := xEngine.Engine.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxIdleConns(conf.GetInt(key+".max_idle_conns", 10))
		sqlDB.SetMaxOpenConns(conf.GetInt(key+".max_open_conns", 100))
		sqlDB.SetConnMaxLifetime(time.Hour)

		if _isD := conf.GetBoolean(key + ".is_default"); _isD {
			es[transaction.DefaultDatabase] = xEngine
		}

		es[key] = xEngine
	}

	return es, nil
}
