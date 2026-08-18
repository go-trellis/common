/*
Copyright © 2026 Henry Huang <hhh@rutcode.com>

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
	"context"
	"errors"
	"testing"

	"github.com/go-trellis/common/transaction"

	"xorm.io/xorm"
)

func testMySQLEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("mysql", "user:pass@tcp(127.0.0.1:1)/db?timeout=1ms&parseTime=true")
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	t.Cleanup(func() { _ = engine.Close() })
	return engine
}

func TestCommitNilLogicClosesTXSession(t *testing.T) {
	engine := testMySQLEngine(t)
	session := engine.NewSession()
	tr := &trans{isTrans: true, engine: engine, session: session}

	err := tr.Commit(nil)
	if err != transaction.ErrNotFoundFunction {
		t.Fatalf("Commit(nil) error = %v, want %v", err, transaction.ErrNotFoundFunction)
	}
	if !session.IsClosed() {
		t.Fatal("TX session should be closed when Logic is missing")
	}
}

func TestCommitNilLogicMapClosesTXSession(t *testing.T) {
	engine := testMySQLEngine(t)
	session := engine.NewSession()
	tr := &trans{isTrans: true, engine: engine, session: session}

	err := tr.Commit(map[int]any{transaction.BeforeLogic: func() {}})
	if err != transaction.ErrNotFoundFunction {
		t.Fatalf("Commit error = %v, want %v", err, transaction.ErrNotFoundFunction)
	}
	if !session.IsClosed() {
		t.Fatal("TX session should be closed when Logic key is missing")
	}
}

func TestCommitNonTXClosesSessions(t *testing.T) {
	engine := testMySQLEngine(t)
	tr := &trans{isTrans: false, engine: engine}
	repo := NewBaseRepo()

	var gotSession *xorm.Session
	err := tr.Commit(func(repos ...any) error {
		r := repos[0].(*BaseRepo)
		gotSession = r.session
		if gotSession == nil {
			t.Fatal("repo session should be set during Logic")
		}
		if gotSession.IsClosed() {
			t.Fatal("repo session should be open during Logic")
		}
		return nil
	}, repo)
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if gotSession == nil {
		t.Fatal("logic was not called")
	}
	if !gotSession.IsClosed() {
		t.Fatal("non-TX session should be closed after Commit")
	}
}

func TestTransactionDoClosesSessionOnBeginError(t *testing.T) {
	engine := testMySQLEngine(t)
	// Dial will fail quickly; session must still be closed by TransactionDo.
	err := TransactionDo(engine, func(s *xorm.Session) error {
		t.Fatal("logic should not run when Begin fails")
		return nil
	})
	if err == nil {
		t.Fatal("expected Begin error")
	}
}

func TestEngineContextBindsSession(t *testing.T) {
	engine := testMySQLEngine(t)
	xEngine := &XEngine{Engine: engine}
	ctx := context.WithValue(context.Background(), "trace_id", "trc-sess")

	bound := xEngine.Context(ctx)
	if bound == nil {
		t.Fatal("Engine.Context returned nil")
	}
	tr, err := bound.BeginNonTransaction()
	if err != nil {
		t.Fatalf("Context().BeginNonTransaction: %v", err)
	}
	tx, ok := tr.(*trans)
	if !ok || tx.ctx == nil {
		t.Fatal("trans should keep ctx")
	}
	if id, _ := tx.ctx.Value("trace_id").(string); id != "trc-sess" {
		t.Fatalf("trans ctx trace_id = %q", id)
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	sessAny, err := xEngine.Context(canceled).NewSession()
	if err != nil {
		t.Fatalf("Context().NewSession: %v", err)
	}
	sess, ok := sessAny.(*xorm.Session)
	if !ok {
		t.Fatalf("session type = %T", sessAny)
	}
	defer sess.Close()
	if err := sess.Ping(); !errors.Is(err, context.Canceled) {
		t.Fatalf("session.Ping want context.Canceled (ctx bound), got %v", err)
	}
}

func TestCtxEngineCloseDoesNotCloseEngine(t *testing.T) {
	engine := testMySQLEngine(t)
	xEngine := &XEngine{Engine: engine}
	if err := xEngine.Context(context.Background()).Close(); err != nil {
		t.Fatalf("Context().Close: %v", err)
	}
	sess, err := xEngine.NewXORMSession()
	if err != nil {
		t.Fatalf("NewXORMSession after wrapper Close: %v", err)
	}
	defer sess.Close()
}

func TestCtxEngineTransactionDoKeepsCtx(t *testing.T) {
	engine := testMySQLEngine(t)
	xEngine := &XEngine{Engine: engine}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	type hasTransactionDo interface {
		TransactionDo(func(*xorm.Session) error) error
	}
	td, ok := xEngine.Context(canceled).(hasTransactionDo)
	if !ok {
		t.Fatal("Context() engine should keep TransactionDo")
	}
	err := td.TransactionDo(func(*xorm.Session) error {
		t.Fatal("logic should not run when ctx is canceled")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("TransactionDo want context.Canceled, got %v", err)
	}
}
