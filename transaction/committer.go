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
	"reflect"

	"trellis.tech/trellis/common.v2"
)

// DefaultDatabase default database key
var (
	DefaultDatabase = common.FormatNamespaceString("xorm_ext:database")
)

// Committer 事务处理者
type Committer interface {
	TX(fn any, repos ...Repo) error
	TXWithName(fn any, name string, repos ...Repo) error
	NonTX(fn any, repos ...Repo) error
	NonTXWithName(fn any, name string, repos ...Repo) error
}

func NewCommitter(engines map[string]Engine) Committer {
	return &committer{
		engines: engines,
	}
}

type committer struct {
	engines map[string]Engine
}

// TX do transaction function by default database
func (p *committer) TX(fn any, repos ...Repo) error {
	return p.TXWithName(fn, DefaultDatabase, repos...)
}

// TXWithName do transaction function with name of database
func (p *committer) TXWithName(fn any, name string, repos ...Repo) error {
	return p.doCommit(fn, name, true, repos...)
}

// NonTX do non transaction function by default database
func (p *committer) NonTX(fn any, repos ...Repo) error {
	return p.NonTXWithName(fn, DefaultDatabase, repos...)
}

// NonTXWithName do non transaction function with name of database
func (p *committer) NonTXWithName(fn any, name string, repos ...Repo) error {
	return p.doCommit(fn, name, false, repos...)
}

func (p *committer) checkRepos(txFunc any, originRepos []Repo) error {
	if txFunc == nil {
		return ErrNotFoundFunction
	}

	if reposLen := len(originRepos); reposLen < 1 {
		return ErrAtLeastOneRepo
	}

	return nil
}

func (p *committer) createNewInstance(origin any) (any, error) {
	if repo, err := Derive(origin); err != nil {
		return nil, err
	} else if repo != nil {
		return repo, nil
	}

	newRepoV := reflect.New(reflect.ValueOf(reflect.Indirect(reflect.ValueOf(origin)).Interface()).Type())
	if !newRepoV.IsValid() {
		return nil, ErrFailToCreateRepo
	}

	newRepoI := newRepoV.Interface()

	if err := Inherit(newRepoI, origin); err != nil {
		return nil, err
	}

	return newRepoI, nil
}

func (p *committer) doCommit(fn any, name string, isTransaction bool, repos ...Repo) (err error) {
	engine, ok := p.engines[name]
	if !ok {
		return ErrNotFoundEngine
	}

	if err = p.checkRepos(fn, repos); err != nil {
		return err
	}

	var trans Transaction
	if isTransaction {
		trans, err = engine.BeginTransaction()
		if err != nil {
			return err
		}
	} else {
		trans, err = engine.BeginNonTransaction()
		if err != nil {
			return err
		}
	}

	var (
		_newRepos []any
	)
	for _, origin := range repos {

		_newRepoI, err := p.createNewInstance(origin)
		if err != nil {
			return err
		}
		_newRepos = append(_newRepos, _newRepoI)
	}

	return trans.Commit(fn, _newRepos...)
}
