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

// NonTX do non transaction function by default database
func (p *Repo) NonTX(fn interface{}, repos ...interface{}) error {
	return p.NonTXWithName(fn, DefaultDatabase, repos...)
}

// NonTXWithName do non transaction function with name of database
func (p *Repo) NonTXWithName(fn interface{}, name string, repos ...interface{}) error {
	if err := p.checkRepos(fn, repos); err != nil {
		return err
	}

	var (
		_newRepos      []interface{}
		_newTXormRepos []*Repo
	)

	for _, origin := range repos {

		_repo := getRepo(origin)
		if _repo == nil {
			return ErrStructCombineWithRepo
		}

		_newTXorm, _newRepoI, err := createNewTXorm(origin)
		if err != nil {
			return err
		}

		_newTXorm.engines = _repo.engines
		_newTXorm.defEngine = _repo.defEngine

		if err := _newTXorm.beginNonTransaction(name); err != nil {
			return err
		}

		_newRepos = append(_newRepos, _newRepoI)
		_newTXormRepos = append(_newTXormRepos, _newTXorm)
	}

	return _newTXormRepos[0].commitNonTransaction(fn, _newRepos...)
}

func (p *Repo) beginNonTransaction(name string) error {
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

func (p *Repo) commitNonTransaction(txFunc interface{}, repos ...interface{}) error {
	if p.isTransaction {
		return ErrNonTransactionCantCommit
	}

	_fun := GetLogicFuncs(txFunc)

	var (
		_values []interface{}
		err     error
	)

	if _fun.BeforeLogic != nil {
		if _, err = CallFunc(_fun.BeforeLogic, _fun, repos); err != nil {
			return err
		}
	}

	if _fun.Logic != nil {
		if _values, err = CallFunc(_fun.Logic, _fun, repos); err != nil {
			return err
		}
	}

	if _fun.AfterLogic != nil {
		if _values, err = CallFunc(_fun.AfterLogic, _fun, repos); err != nil {
			return err
		}
	}

	if _fun.AfterCommit != nil {
		if _, err = CallFunc(_fun.AfterCommit, _fun, _values); err != nil {
			return err
		}
	}

	return nil
}
