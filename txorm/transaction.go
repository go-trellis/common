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
	"trellis.tech/trellis/common.v1/errcode"
	"trellis.tech/trellis/common.v1/transaction"

	"xorm.io/xorm"
)

type trans struct {
	isTrans bool
	engine  *xorm.Engine
	session *xorm.Session
}

func (p *trans) Session() interface{} {
	if p.isTrans {
		if p.session == nil {
			p.session = p.engine.NewSession()
		}
		return p.session
	}
	return p.engine.NewSession()
}

func (p *trans) IsTransaction() bool {
	return p.isTrans
}

func (p *trans) Commit(fun interface{}, repos ...interface{}) error {
	fn := transaction.GetLogicFunc(fun)
	if fn == nil || fn.Logic == nil {
		//return todo error
		return nil
	}

	var (
		_values   []interface{}
		_newRepos []interface{}
		err       error
	)

	if p.IsTransaction() {

		defer func() { _ = p.session.Close() }()

		if err := p.session.Begin(); err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.session.Rollback()
			}
		}()

		for _, repo := range repos {
			tRepo, ok := repo.(transaction.Repo)
			if !ok {
				return errcode.New("not transaction repo, check the repo implement transaction repo")
			}
			if err = tRepo.SetSession(p.session); err != nil {
				return err
			}
			_newRepos = append(_newRepos, tRepo)
		}
	} else {
		for _, repo := range repos {
			tRepo, ok := repo.(transaction.Repo)
			if !ok {
				return errcode.New("not transaction repo, check the repo implement transaction repo")
			}
			if err = tRepo.SetSession(p.engine.NewSession()); err != nil {
				return err
			}
			_newRepos = append(_newRepos, tRepo)
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
		_ = p.session.Commit()
	}

	if _, err = transaction.CallFunc(fn.AfterCommit, _values); err != nil {
		return err
	}

	return nil
}
