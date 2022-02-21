package txorm

import (
	"database/sql"

	"trellis.tech/trellis/common.v1/transaction"

	"xorm.io/xorm"
)

var _ transaction.Engine = (*xEngine)(nil)

type xEngine struct {
	*xorm.Engine
}

func NewEngine(driver string, dsn string) (*xEngine, error) {
	engine, err := xorm.NewEngine(driver, dsn)
	if err != nil {
		return nil, err
	}
	x := &xEngine{
		Engine: engine,
	}
	return x, nil
}

func (p *xEngine) NewSession() interface{} {
	return p.Engine.NewSession()
}

func (p *xEngine) Exec(sqlOrArgs ...interface{}) (sql.Result, error) {
	return p.Engine.Exec(sqlOrArgs...)
}

func (p *xEngine) BeginTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: true, engine: p.Engine, session: p.Engine.NewSession()}, nil
}

func (p *xEngine) BeginNonTransaction() (transaction.Transaction, error) {
	return &trans{isTrans: false, engine: p.Engine}, nil
}