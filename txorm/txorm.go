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

import (
	"reflect"

	"xorm.io/xorm"
)

// Transaction 事务处理对象
// TODO: Engines 为一个option，可以支持redis等
type Transaction interface {
	SetEngines(engines map[string]*xorm.Engine)
	Session() *xorm.Session
	BeginTransaction(name string) error
	BeginNonTransaction(name string) error
}

// TXFunc Transaction function
type TXFunc func(repos ...interface{}) error

// Inheritor inherit function
type Inheritor interface {
	Inherit(repo interface{}) error
}

// Inherit new repository from origin repository
func Inherit(new, origin interface{}) error {
	if inheritor, ok := new.(Inheritor); ok {
		return inheritor.Inherit(origin)
	}
	return nil
}

// Derivative derive function
type Derivative interface {
	Derive() (repo interface{}, err error)
}

// Derive derive from developer function
func Derive(origin interface{}) (interface{}, error) {
	if d, ok := origin.(Derivative); ok {
		return d.Derive()
	}
	return nil, nil
}

// New get trellis xorm committer
func New() Transaction {
	return &Repo{}
}

func getRepo(v interface{}) *Repo {
	_deepRepo := DeepFields(v, reflect.TypeOf(new(Repo)), []reflect.Value{})
	if deepRepo, ok := _deepRepo.(*Repo); ok {
		return deepRepo
	}
	return nil
}

func createNewTXorm(origin interface{}) (*Repo, interface{}, error) {
	if repo, err := Derive(origin); err != nil {
		return nil, nil, err
	} else if repo != nil {
		return getRepo(repo), repo, nil
	}

	newRepoV := reflect.New(reflect.ValueOf(
		reflect.Indirect(reflect.ValueOf(origin)).Interface()).Type())
	if !newRepoV.IsValid() {
		return nil, nil, ErrFailToCreateRepo
	}

	newRepoI := newRepoV.Interface()

	if err := Inherit(newRepoI, origin); err != nil {
		return nil, nil, err
	}

	newTXorm := getRepo(newRepoI)

	if newTXorm == nil {
		return nil, nil, ErrFailToConvertTXToNonTX
	}
	return newTXorm, newRepoI, nil
}
