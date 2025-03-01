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
	"trellis.tech/trellis/common.v3/errcode"
	"trellis.tech/trellis/common.v3/transaction"

	"xorm.io/xorm"
)

type trans struct {
	isTrans bool
	engine  *xorm.Engine
	session *xorm.Session
}

// Session returns the current session. If there is no active session, a new one will be created.
func (p *trans) Session() any {
	// check if the session is already created and active
	if p.isTrans {
		// session already exists and active, return it directly
		if p.session == nil {
			p.session = p.engine.NewSession()
		}
		// return the existing session
		return p.session
	}
	// if there is no active session, create a new one and return it
	return p.engine.NewSession()
}

// IsTransaction returns true if there is an active transaction.
func (p *trans) IsTransaction() bool {
	return p.isTrans
}

// Commit executes the logic function and commits the transaction. If there is an error during the execution of the logic function, the transaction will be rolled back. Otherwise, the transaction will be committed.
func (p *trans) Commit(fun any, repos ...any) error {
	// get the logic function
	fn := transaction.GetLogicFunc(fun)
	if fn == nil || fn.Logic == nil {
		return errcode.New("logic function is not found")
	}

	var (
		_values   []any
		_newRepos []any
		err       error
	)

	if p.IsTransaction() {
		defer p.session.Close()

		if err := p.session.Begin(); err != nil {
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
		for _, repo := range repos {
			session := p.engine.NewSession()
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

	// execute before logic
	if _, err = transaction.CallFunc(fn.BeforeLogic, _newRepos...); err != nil {
		return err
	}

	// execute logic
	if _values, err = transaction.CallFunc(fn.Logic, _newRepos...); err != nil {
		return err
	}

	// execute after logic
	if _, err = transaction.CallFunc(fn.AfterLogic, _newRepos...); err != nil {
		return err
	}

	// commit transaction
	if p.isTrans {
		if err = p.session.Commit(); err != nil {
			return err
		}
	}

	// call after commit logic
	if _, err = transaction.CallFunc(fn.AfterCommit, _values); err != nil {
		return err
	}

	return nil
}

// setTransactionRepoSession sets the session for a transaction repo. It returns an error if the repository does not implement the transaction.Repo interface.
func setTransactionRepoSession(repo any, session *xorm.Session) error {
	tRepo, ok := repo.(transaction.Repo)
	if !ok {
		return errcode.New("not transaction repo, check the repo implement transaction repo")
	}
	return tRepo.SetSession(session)
}

// Do to do transaction with customer function
func Do(engine transaction.Engine, fn func(*xorm.Session) error) error {
	xEngine, err := assertXormEngine(engine)
	if err != nil {
		return err
	}
	session := xEngine.Engine.NewSession()
	defer session.Close()
	return fn(session)
}

// TransactionDo to do transaction with customer function
func TransactionDo(engine transaction.Engine, fn func(*xorm.Session) error) error {
	xEngine, err := assertXormEngine(engine)
	if err != nil {
		return err
	}
	return TransactionDoWithSession(xEngine.Engine.NewSession(), fn)
}

func assertXormEngine(engine transaction.Engine) (*XEngine, error) {
	if engine == nil {
		return nil, errcode.New("nil transaction engine")
	}
	xEngine, ok := engine.(*XEngine)
	if !ok {
		return nil, errcode.New("not txorm XEngine")
	}

	return xEngine, nil
}

// TransactionDoWithSession to do transaction with customer function
func TransactionDoWithSession(s *xorm.Session, fn func(*xorm.Session) error) (err error) {
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
