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

type Transaction interface {
	Session() any
	IsTransaction() bool

	Commit(fn any, repos ...any) error
}

type Repo interface {
	SetSession(any) error
}

// Derivative derive function
type Derivative interface {
	Derive() (repo any, err error)
}

// Derive derive from developer function
func Derive(origin any) (any, error) {
	if d, ok := origin.(Derivative); ok {
		return d.Derive()
	}
	return nil, nil
}

// Inheritor inherit function
type Inheritor interface {
	Inherit(repo any) error
}

// Inherit new repository from origin repository
func Inherit(new, origin any) error {
	if inheritor, ok := new.(Inheritor); ok {
		return inheritor.Inherit(origin)
	}
	return nil
}
