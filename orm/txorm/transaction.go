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
	"context"

	"github.com/go-trellis/common/errors/errcode"
	"github.com/go-trellis/common/orm/transaction"

	"xorm.io/xorm"
)

type trans struct {
	isTrans bool
	engine  *xorm.Engine
	session *xorm.Session
	ctx     context.Context
}

func applySessionContext(session *xorm.Session, ctx context.Context) *xorm.Session {
	if session == nil || ctx == nil {
		return session
	}
	return session.Context(ctx)
}

func (p *trans) newSession() *xorm.Session {
	return applySessionContext(p.engine.NewSession(), p.ctx)
}

// Session returns the current session. If there is no active session, a new one will be created.
func (p *trans) Session() any {
	// check if the session is already created and active
	if p.isTrans {
		// session already exists and active, return it directly
		if p.session == nil {
			p.session = p.newSession()
		}
		// return the existing session
		return p.session
	}
	// if there is no active session, create a new one and return it
	return p.newSession()
}

// IsTransaction returns true if there is an active transaction.
func (p *trans) IsTransaction() bool {
	return p.isTrans
}

// Commit executes the logic function and commits the transaction. If there is an error during the execution of the logic function, the transaction will be rolled back. Otherwise, the transaction will be committed.
func (p *trans) Commit(fun any, repos ...any) (err error) {
	fn := transaction.GetLogicFunc(fun)

	// TX session must be closed on every exit path, including invalid Logic.
	if p.IsTransaction() {
		if p.session == nil {
			p.session = p.newSession()
		}
		defer p.session.Close()
	}

	if fn == nil || fn.Logic == nil {
		return transaction.ErrNotFoundFunction
	}

	var (
		_values   []any
		_newRepos []any
		sessions  []*xorm.Session
	)

	if p.IsTransaction() {
		if err = p.session.Begin(); err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.session.Rollback()
			}
		}()

		for _, repo := range repos {
			if err = setTransactionRepoSession(repo, p.session); err != nil {
				return err
			}
			_newRepos = append(_newRepos, repo)
		}
	} else {
		defer func() {
			for _, s := range sessions {
				_ = s.Close()
			}
		}()

		for _, repo := range repos {
			session := p.newSession()
			sessions = append(sessions, session)
			if err = setTransactionRepoSession(repo, session); err != nil {
				return err
			}
			_newRepos = append(_newRepos, repo)
		}
	}

	defer func() {
		if err != nil {
			transaction.CallFunc(fn.OnError, err)
		}
	}()

	if _, err = transaction.CallFunc(fn.BeforeLogic, _newRepos...); err != nil {
		return err
	}

	if _values, err = transaction.CallFunc(fn.Logic, _newRepos...); err != nil {
		return err
	}

	if _, err = transaction.CallFunc(fn.AfterLogic, _newRepos...); err != nil {
		return err
	}

	if p.isTrans {
		if err = p.session.Commit(); err != nil {
			return err
		}
	}

	if _, err = transaction.CallFunc(fn.AfterCommit, _values); err != nil {
		return err
	}

	return nil
}

func setTransactionRepoSession(repo any, session *xorm.Session) error {
	tRepo, ok := repo.(transaction.Repo)
	if !ok {
		return errcode.New("not transaction repo, check the repo implement transaction repo")
	}
	return tRepo.SetSession(session)
}

// Do to do transaction with customer function
func Do(engine transaction.Engine, fn func(*xorm.Session) error) error {
	xEngine, ctx, err := unwrapXEngine(engine)
	if err != nil {
		return err
	}
	session := applySessionContext(xEngine.Engine.NewSession(), ctx)
	defer session.Close()
	return fn(session)
}

// TransactionDo to do transaction with customer function
func TransactionDo(engine transaction.Engine, fn func(*xorm.Session) error) error {
	xEngine, ctx, err := unwrapXEngine(engine)
	if err != nil {
		return err
	}
	session := applySessionContext(xEngine.Engine.NewSession(), ctx)
	defer session.Close()
	return TransactionDoWithSession(session, fn)
}

func unwrapXEngine(engine transaction.Engine) (*XEngine, context.Context, error) {
	if engine == nil {
		return nil, nil, errcode.New("nil transaction engine")
	}
	switch e := engine.(type) {
	case *XEngine:
		return e, nil, nil
	case *ctxEngine:
		return e.XEngine, e.ctx, nil
	default:
		return nil, nil, errcode.New("not txorm XEngine")
	}
}

// TransactionDoWithSession to do transaction with customer function.
// Caller owns the session lifecycle.
func TransactionDoWithSession(s *xorm.Session, fn func(*xorm.Session) error) (err error) {
	if s == nil {
		return errcode.New("nil session")
	}
	if err = s.Begin(); err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = s.Rollback()
			return
		}
		err = s.Commit()
	}()
	err = fn(s)
	return
}
