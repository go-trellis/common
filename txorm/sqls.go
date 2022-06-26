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
	"fmt"
	"reflect"

	"trellis.tech/trellis/common.v1/errcode"

	"xorm.io/xorm"
)

type BaseRepository interface {
	Insert(...interface{}) (int64, error)
	InsertMulti(beans interface{}, opts ...InsertMultiOption) (int64, error)
	Update(bean interface{}, opts ...UpdateOption) (int64, error)
	Delete(bean interface{}, opts ...DeleteOption) (int64, error)
	Get(bean interface{}, opts ...GetOption) (bool, error)
	Find(beans interface{}, opts ...GetOption) error
}

type BaseRepo struct {
	session *xorm.Session
}

func (p *BaseRepo) SetSession(x interface{}) error {
	session, err := CheckSession(x)
	if err != nil {
		return nil
	}
	p.session = session
	return nil
}

func NewBaseRepo() BaseRepository {
	return &BaseRepo{}
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

// Do to do transaction with customer function
func Do(engine *xorm.Engine, fn func(*xorm.Session) error) error {
	session := engine.NewSession()
	defer session.Close()
	return fn(session)
}

// TransactionDo to do transaction with customer function
func TransactionDo(engine *xorm.Engine, fn func(*xorm.Session) error) error {
	return TransactionDoWithSession(engine.NewSession(), fn)
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

/// Get Execute

type GetOption func(*GetOptions)
type GetOptions struct {
	Wheres interface{}
	Args   []interface{}

	Limit, Offset int
	OrderBy       string
}

func GetWheres(wheres interface{}) GetOption {
	return func(options *GetOptions) {
		options.Wheres = wheres
	}
}

func GetArgs(args ...interface{}) GetOption {
	return func(options *GetOptions) {
		options.Args = args
	}
}

func GetLimit(limit, offset int) GetOption {
	return func(options *GetOptions) {
		options.Limit = limit
		options.Offset = offset
	}
}

func GetOrderBy(order string) GetOption {
	return func(options *GetOptions) {
		options.OrderBy = order
	}
}

func Get(session *xorm.Session, bean interface{}, opts ...GetOption) (bool, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	session = session.Where(getOptions.Wheres, getOptions.Args...)
	if getOptions.Limit > 0 {
		session = session.Limit(getOptions.Limit, getOptions.Offset)
	}
	if len(getOptions.OrderBy) > 0 {
		session = session.OrderBy(getOptions.OrderBy)
	}
	return session.Get(bean)
}

func Find(session *xorm.Session, bean interface{}, opts ...GetOption) error {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	session = session.Where(getOptions.Wheres, getOptions.Args...)
	if getOptions.Limit > 0 {
		session = session.Limit(getOptions.Limit, getOptions.Offset)
	}
	if len(getOptions.OrderBy) > 0 {
		session = session.OrderBy(getOptions.OrderBy)
	}
	return session.Find(bean)
}

/// Update Execute

type UpdateOption func(*UpdateOptions)
type UpdateOptions struct {
	Wheres interface{}
	Args   []interface{}
	Cols   []string
}

func UpdateWheres(wheres interface{}) UpdateOption {
	return func(options *UpdateOptions) {
		options.Wheres = wheres
	}
}

func UpdateArgs(args ...interface{}) UpdateOption {
	return func(options *UpdateOptions) {
		options.Args = args
	}
}

func UpdateCols(cols ...string) UpdateOption {
	return func(options *UpdateOptions) {
		options.Cols = cols
	}
}

func Update(session *xorm.Session, bean interface{}, opts ...UpdateOption) (int64, error) {
	updateOptions := &UpdateOptions{}
	for _, opt := range opts {
		opt(updateOptions)
	}

	session = session.Where(updateOptions.Wheres, updateOptions.Args...)
	if len(updateOptions.Cols) > 0 {
		session = session.Cols(updateOptions.Cols...)
	}

	return session.Update(bean)
}

/// InsertMulti Execute

type InsertMultiOption func(*InsertMultiOptions)
type InsertMultiOptions struct {
	StepNumber int
}

func InsertMultiStepNumber(number int) InsertMultiOption {
	return func(options *InsertMultiOptions) {
		options.StepNumber = number
	}
}

func Insert(session *xorm.Session, beans ...interface{}) (int64, error) {
	return session.Insert(beans...)
}

// InsertMulti insert multi seperated slice data in a big slice with every step number
// default to insert the slice with no seperated.
func InsertMulti(session *xorm.Session, ones interface{}, opts ...InsertMultiOption) (int64, error) {
	options := &InsertMultiOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if options.StepNumber <= 0 {
		return session.InsertMulti(ones)
	}
	sliceOnes := reflect.Indirect(reflect.ValueOf(ones))
	switch sliceOnes.Kind() {
	case reflect.Slice, reflect.Array:
		onesLen := sliceOnes.Len()
		if onesLen == 0 {
			return 0, nil
		}

		switch sliceOnes.Index(0).Kind() {
		case reflect.Interface:
			session = session.NoAutoTime()
		}

		if onesLen <= options.StepNumber {
			return session.InsertMulti(ones)
		}

		loop, count, processNum := 0, 0, onesLen

		for i := 0; i < onesLen; i += options.StepNumber {
			if processNum > options.StepNumber {
				loop = i + options.StepNumber
			} else {
				loop = onesLen
			}
			var multi []interface{}
			for j := i; j < loop; j++ {
				multi = append(multi, sliceOnes.Index(j).Interface())
			}
			session = session.NoAutoTime()
			n, err := session.InsertMulti(multi)
			if err != nil {
				return int64(count) + n, err
			}
			count += int(n)
			processNum -= options.StepNumber
		}

		if count != onesLen {
			return 0, fmt.Errorf("insert number not %d, but %d", onesLen, count)
		}
		return int64(count), nil
	default:
		return session.InsertMulti(ones)
	}
}

/// Delete Execute

type DeleteOption func(*DeleteOptions)
type DeleteOptions struct {
	Wheres interface{}
	Args   []interface{}
}

func DeleteWheres(wheres interface{}) DeleteOption {
	return func(options *DeleteOptions) {
		options.Wheres = wheres
	}
}

func DeleteArgs(args ...interface{}) DeleteOption {
	return func(options *DeleteOptions) {
		options.Args = args
	}
}

func Delete(session *xorm.Session, bean interface{}, opts ...DeleteOption) (int64, error) {
	deleteOptions := &DeleteOptions{}
	for _, opt := range opts {
		opt(deleteOptions)
	}

	return session.Where(deleteOptions.Wheres, deleteOptions.Args...).Delete(bean)
}
