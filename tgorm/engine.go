package tgorm

import (
	"database/sql"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"trellis.tech/trellis/common.v1/errcode"
	"trellis.tech/trellis/common.v1/transaction"
)

var _ transaction.Engine = (*XEngine)(nil)

type XEngine struct {
	Engine *gorm.DB
}

func NewEngine(driver string, dsn string, opts ...gorm.Option) (*XEngine, error) {

	var d gorm.Dialector
	switch driver {
	case "mysql":
		d = mysql.Open(dsn)
	case "sqlite":
		d = sqlite.Open("gorm.db")
	default:
		return nil, errcode.New("not supported other engine")
	}
	engine, err := gorm.Open(d, opts...)
	if err != nil {
		return nil, err
	}

	return &XEngine{Engine: engine}, nil
}

func (p *XEngine) NewSession() (interface{}, error) {
	return p.Engine, nil
}

func (p *XEngine) Exec(query string, args ...interface{}) (sql.Result, error) {
	//return p.DB.(sqlOrArgs...)
	sql, err := p.Engine.DB()
	if err != nil {
		return nil, err
	}
	return sql.Exec(query, args...)
}

func (p *XEngine) BeginTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: true, engine: p.Engine, session: nil}, nil
}

func (p *XEngine) BeginNonTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: false, engine: p.Engine, session: nil}, nil
}

func (p *XEngine) Close() error {
	db, err := p.Engine.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
