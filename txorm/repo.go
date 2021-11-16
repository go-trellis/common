/*
Copyright © 2019 Henry Huang <hhh@rutcode.com>

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

import "xorm.io/xorm"

// Repo trellis xorm
type Repo struct {
	isTransaction bool
	txSession     *xorm.Session

	engines   map[string]*xorm.Engine
	defEngine *xorm.Engine
}

// SetEngines set xorm engines
func (p *Repo) SetEngines(engines map[string]*xorm.Engine) {
	if defEngine, exist := engines[DefaultDatabase]; exist {
		p.engines = engines
		p.defEngine = defEngine
	} else {
		panic(ErrNotFoundDefaultDatabase)
	}
}

// Session get session
func (p *Repo) Session() *xorm.Session {
	return p.txSession
}

// GetEngine get engine by name
func (p *Repo) getEngine(name string) (*xorm.Engine, error) {
	if engine, _exist := p.engines[name]; _exist {
		return engine, nil
	}
	return nil, ErrNotFoundXormEngine
}

func (p *Repo) checkRepos(txFunc interface{}, originRepos ...interface{}) error {
	if reposLen := len(originRepos); reposLen < 1 {
		return ErrAtLeastOneRepo
	}

	if txFunc == nil {
		return ErrNotFoundTransactionFunction
	}
	return nil
}

func (p *Repo) BeginNonTransaction(name string) error {
	if p.isTransaction {
		return ErrFailToConvertTXToNonTX
	}

	_engine, err := p.getEngine(name)
	if err != nil {
		return err
	}

	p.txSession = _engine.NewSession()

	return nil
}

func (p *Repo) BeginTransaction(name string) error {
	if !p.isTransaction {
		p.isTransaction = true
		_engine, err := p.getEngine(name)
		if err != nil {
			return err
		}
		p.txSession = _engine.NewSession()
		if p.txSession == nil {
			return ErrTransactionSessionIsNil
		}
		return nil
	}
	return ErrTransactionIsAlreadyBegin
}

func (p *Repo) commit(transaction bool, fun interface{}, repos ...interface{}) error {
	if p.txSession == nil {
		return ErrTransactionSessionIsNil
	}
	fn := GetLogicFunc(fun)
	if fn == nil || fn.Logic == nil {
		return ErrNotFoundTransactionFunction
	}

	if (!transaction && p.isTransaction) || (transaction && !p.isTransaction) {
		return ErrNonTransactionCantCommit
	}

	var (
		_values []interface{}
		err     error
	)

	if transaction {

		defer func() { _ = p.txSession.Close() }()

		if err := p.txSession.Begin(); err != nil {
			return err
		}

		defer func() {
			if err != nil {
				_ = p.txSession.Rollback()
			}
		}()
	}

	defer func() {
		if err != nil {
			CallFunc(fn.OnError, err)
		}
	}()

	if _, err = CallFunc(fn.BeforeLogic, repos...); err != nil {
		return err
	}

	if _values, err = CallFunc(fn.Logic, repos...); err != nil {
		return err
	}

	if _, err = CallFunc(fn.AfterLogic, repos...); err != nil {
		return err
	}

	if transaction {
		_ = p.txSession.Commit()
	}

	if _, err = CallFunc(fn.AfterCommit, _values); err != nil {
		return err
	}

	return nil
}
