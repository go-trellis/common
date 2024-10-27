/*
Copyright © 2024 Henry Huang <hhh@rutcode.com>

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
	"trellis.tech/trellis/common.v2/errcode"
	"trellis.tech/trellis/common.v2/transaction"

	"xorm.io/xorm"
)

var _ BaseRepository = (*BaseRepo)(nil)

type BaseRepository interface {
	Get(bean interface{}, opts ...GetOption) (bool, error)
	Find(beans interface{}, opts ...GetOption) error
	FindAndCount(beans interface{}, opts ...GetOption) (int64, error)
	Count(beans interface{}, opts ...GetOption) (int64, error)

	ExecRepository
}

type ExecRepository interface {
	Insert(...interface{}) (int64, error)
	InsertMulti(beans interface{}, opts ...InsertMultiOption) (int64, error)
	Update(bean interface{}, opts ...UpdateOption) (int64, error)
	Delete(bean interface{}, opts ...DeleteOption) (int64, error)

	transaction.Repo
}

type BaseRepo struct {
	session *xorm.Session
}

func NewBaseRepo(ss ...*xorm.Session) *BaseRepo {
	r := &BaseRepo{}
	for _, s := range ss {
		if s != nil {
			r.session = s
			break
		}
	}
	return r
}

func (p *BaseRepo) SetSession(x any) error {
	session, err := CheckSession(x)
	if err != nil {
		return nil
	}
	p.session = session
	return nil
}

func CheckSession(x interface{}) (*xorm.Session, error) {
	switch t := x.(type) {
	case *xorm.Session:
		return t, nil
	case xorm.Session:
		return &t, nil
	default:
		return nil, errcode.New("not supported session type")
	}
}

func (p *BaseRepo) Insert(beans ...interface{}) (int64, error) {
	return Insert(p.session, beans...)
}

func (p *BaseRepo) InsertMulti(beans interface{}, opts ...InsertMultiOption) (int64, error) {
	return InsertMulti(p.session, beans, opts...)
}

func (p *BaseRepo) Update(bean interface{}, opts ...UpdateOption) (int64, error) {
	return Update(p.session, bean, opts...)
}

func (p *BaseRepo) Delete(bean interface{}, opts ...DeleteOption) (int64, error) {
	return Delete(p.session, bean, opts...)
}

func (p *BaseRepo) Get(bean interface{}, opts ...GetOption) (bool, error) {
	return Get(p.session, bean, opts...)
}

func (p *BaseRepo) Find(beans interface{}, opts ...GetOption) error {
	return Find(p.session, beans, opts...)
}

func (p *BaseRepo) FindAndCount(beans interface{}, opts ...GetOption) (int64, error) {
	return FindAndCount(p.session, beans, opts...)
}

func (p *BaseRepo) Count(beans interface{}, opts ...GetOption) (int64, error) {
	return Count(p.session, beans, opts...)
}
