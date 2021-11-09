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

// TX do transaction function by default database
func (p *Repo) TX(fn interface{}, repos ...interface{}) error {
	return p.TXWithName(fn, DefaultDatabase, repos...)
}

// TXWithName do transaction function with name of database
func (p *Repo) TXWithName(fn interface{}, name string, repos ...interface{}) error {
	if err := p.checkRepos(fn, repos); err != nil {
		return err
	}

	var (
		_newRepos      []interface{}
		_newTXormRepos []*Repo
	)

	for _, origin := range repos {

		repo := getRepo(origin)
		if repo == nil {
			return ErrStructCombineWithRepo
		}

		_newXorm, newRepoI, err := createNewTXorm(origin)
		if err != nil {
			return err
		}

		_newXorm.engines = repo.engines
		_newXorm.defEngine = repo.defEngine
		_newRepos = append(_newRepos, newRepoI)
		_newTXormRepos = append(_newTXormRepos, _newXorm)
	}

	if err := _newTXormRepos[0].beginTransaction(name); err != nil {
		return err
	}

	for i := range _newTXormRepos {
		_newTXormRepos[i].txSession = _newTXormRepos[0].txSession
		_newTXormRepos[i].isTransaction = _newTXormRepos[0].isTransaction
	}

	return _newTXormRepos[0].commitTransaction(fn, _newRepos...)
}

func (p *Repo) beginTransaction(name string) error {
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

func (p *Repo) commitTransaction(txFunc interface{}, repos ...interface{}) error {
	if !p.isTransaction {
		return ErrNonTransactionCantCommit
	}

	if p.txSession == nil {
		return ErrTransactionSessionIsNil
	}
	defer func() { _ = p.txSession.Close() }()

	if txFunc == nil {
		return ErrNotFoundTransactionFunction
	}

	isNeedRollBack := true

	if err := p.txSession.Begin(); err != nil {
		return err
	}

	defer func() {
		if isNeedRollBack {
			_ = p.txSession.Rollback()
		}
	}()

	_funcs := GetLogicFuncs(txFunc)

	var (
		_values []interface{}
		err     error
	)

	if _funcs.BeforeLogic != nil {
		if _, err = CallFunc(_funcs.BeforeLogic, _funcs, repos); err != nil {
			return err
		}
	}

	if _funcs.Logic != nil {
		if _values, err = CallFunc(_funcs.Logic, _funcs, repos); err != nil {
			return err
		}
	}

	if _funcs.AfterLogic != nil {
		if _values, err = CallFunc(_funcs.AfterLogic, _funcs, repos); err != nil {
			return err
		}
	}

	isNeedRollBack = false
	if err := p.txSession.Commit(); err != nil {
		return err
	}

	if _funcs.AfterCommit != nil {
		if _, err = CallFunc(_funcs.AfterCommit, _funcs, _values); err != nil {
			return err
		}
	}

	return nil
}
