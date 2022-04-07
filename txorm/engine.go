package txorm

import (
	"database/sql"

	"trellis.tech/trellis/common.v1/transaction"

	"xorm.io/xorm"
)

var _ transaction.Engine = (*XEngine)(nil)

type XEngine struct {
	*xorm.Engine
}

func NewEngine(driver string, dsn string) (*XEngine, error) {
	engine, err := xorm.NewEngine(driver, dsn)
	if err != nil {
		return nil, err
	}
	x := &XEngine{
		Engine: engine,
	}
	return x, nil
}

func (p *XEngine) NewSession() (interface{}, error) {
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
