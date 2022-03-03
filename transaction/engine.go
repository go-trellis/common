package transaction

import "database/sql"

type Engine interface {
	NewSession() (interface{}, error)
	Exec(sql string, args ...interface{}) (sql.Result, error)
	BeginTransaction() (Transaction, error)
	BeginNonTransaction() (Transaction, error)
	Close() error
}
