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

package transaction

import (
	"context"
	"database/sql"
)

// Engine transaction engine interface.
type Engine interface {
	// NewSession creates a new session for transaction.
	NewSession() (any, error)
	// Exec executes the SQL statement.
	Exec(sql string, args ...any) (sql.Result, error)
	// BeginTransaction starts a new transaction.
	BeginTransaction() (Transaction, error)
	// BeginNonTransaction starts a non-transactional session.
	BeginNonTransaction() (Transaction, error)
	// Context returns an Engine whose NewSession/Begin*/Exec bind ctx
	// (SQL AfterSQL can read trace_id). Do not SetLogger per request.
	// Close on the returned Engine must not close the underlying shared engine.
	Context(ctx context.Context) Engine
	// AddHook attaches an ORM-specific hook. Concrete engines (e.g. txorm.XEngine)
	// assert the value to their hook type; pass any for ORM-agnostic callers.
	AddHook(hook any) error
	// Close closes the engine.
	Close() error
}
