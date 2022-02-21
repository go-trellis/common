package transaction

import "database/sql"

type Engine interface {
	NewSession() interface{}
	Exec(sqlOrArgs ...interface{}) (sql.Result, error)
	BeginTransaction() (Transaction, error)
	BeginNonTransaction() (Transaction, error)
	Close() error
}
