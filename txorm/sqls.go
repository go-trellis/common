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
	"strings"

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

func (p *BaseRepo) SetSession(x interface{}) error {
	session, err := CheckSession(x)
	if err != nil {
		return nil
	}
	p.session = session
	return nil
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

func Args(args ...interface{}) []interface{} {
	var bs []interface{}
	return append(bs, args...)
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

type In struct {
	Column string
	Args   []interface{}
}

func InOpts(column string, args ...interface{}) *In {
	if column == "" {
		return nil
	}
	return &In{column, args}
}

/// Get Execute

type GetOption func(*GetOptions)
type GetOptions struct {
	Wheres interface{}
	Args   []interface{}

	Builders []*Builder

	InWheres    []*In
	NotInWheres []*In

	Limit, Offset int
	OrderBy       string
	GroupBy       string
	Having        string

	Cols     []string
	Distinct []string
}

func (p *GetOptions) Session(session *xorm.Session) *xorm.Session {
	for _, where := range p.InWheres {
		if where != nil {
			session = session.In(where.Column, where.Args...)
		}
	}
	for _, where := range p.NotInWheres {
		if where != nil {
			session = session.NotIn(where.Column, where.Args...)
		}
	}

	for _, b := range p.Builders {
		switch b.LinkType {
		case LinkTypeOR:
			session = session.Or(b.Where, b.Args...)
		case LinkTypeNotSet, LinkTypeAND:
			session = session.And(b.Where, b.Args...)
		default:
			panic(fmt.Errorf("not supported link type %s", b.LinkType))
		}
	}

	if p.Wheres != nil {
		session = session.Where(p.Wheres, p.Args...)
	}
	if p.Limit > 0 {
		session = session.Limit(p.Limit, p.Offset)
	}
	if p.Distinct != nil && len(p.Distinct) > 0 {
		session = session.Distinct(p.Distinct...)
	}
	if len(p.OrderBy) > 0 {
		session = session.OrderBy(p.OrderBy)
	}
	if len(p.GroupBy) > 0 {
		session = session.GroupBy(p.GroupBy)
	}
	if len(p.Having) > 0 {
		session = session.Having(p.Having)
	}
	if len(p.Cols) > 0 {
		session = session.Cols(p.Cols...)
	}

	return session
}

func (p *GetOptions) addMapWheres(maps map[string]interface{}) {
	if maps == nil {
		return
	}

	if p.Wheres == nil {
		p.Wheres = maps
		return
	}

	switch t := p.Wheres.(type) {
	case map[string]interface{}:
		for k, v := range maps {
			t[k] = v
		}
		p.Wheres = t
	default:
		panic(fmt.Errorf("not supported maps wheres type in: %s", reflect.TypeOf(t).String()))
	}
}

func (p *GetOptions) addStringWheres(where string) {
	if p.Wheres == nil {
		p.Wheres = where
		return
	}
	switch t := p.Wheres.(type) {
	case string:
		if t != "" {
			p.Wheres = t + " AND " + where
			return
		}
		p.Wheres = t
	default:
		panic(fmt.Errorf("not supported string wheres type in: %s", reflect.TypeOf(t).String()))
	}
}

func GetWheres(wheres interface{}) GetOption {
	return func(options *GetOptions) {
		switch ts := wheres.(type) {
		case string:
			options.addStringWheres(ts)
		case []string:
			options.addStringWheres(strings.Join(ts, " AND "))
		case map[string]interface{}:
			options.addMapWheres(ts)
		default:
			panic(fmt.Errorf("not supported wheres type: %s", reflect.TypeOf(ts).String()))
		}
	}
}

func GetIn(ins ...*In) GetOption {
	return func(options *GetOptions) {
		options.InWheres = append(options.InWheres, ins...)
	}
}

func GetNotIn(ins ...*In) GetOption {
	return func(options *GetOptions) {
		options.NotInWheres = append(options.NotInWheres, ins...)
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

func GetGroupBy(groupBy string) GetOption {
	return func(options *GetOptions) {
		options.GroupBy = groupBy
	}
}

func GetHaving(having string) GetOption {
	return func(options *GetOptions) {
		options.Having = having
	}
}

func GetCols(cols ...string) GetOption {
	return func(options *GetOptions) {
		options.Cols = cols
	}
}

func GetDistinct(args ...string) GetOption {
	return func(options *GetOptions) {
		options.Distinct = args
	}
}

func Get(session *xorm.Session, bean interface{}, opts ...GetOption) (bool, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).Get(bean)
}

func Find(session *xorm.Session, bean interface{}, opts ...GetOption) error {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).Find(bean)
}

func FindAndCount(session *xorm.Session, bean interface{}, opts ...GetOption) (int64, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).FindAndCount(bean)
}

func Count(session *xorm.Session, bean interface{}, opts ...GetOption) (int64, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).Count(bean)
}

/// Update Execute

type UpdateOption func(*UpdateOptions)
type UpdateOptions struct {
	Wheres interface{}
	Args   []interface{}
	Cols   []string

	InWheres    []*In
	NotInWheres []*In
}

func UpdateWheres(wheres interface{}) UpdateOption {
	return func(options *UpdateOptions) {
		switch ts := wheres.(type) {
		case []string:
			options.Wheres = GetWheres(strings.Join(ts, " AND "))
		default:
			options.Wheres = wheres
		}
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

func UpdateIn(ins ...*In) UpdateOption {
	return func(options *UpdateOptions) {
		options.InWheres = append(options.InWheres, ins...)
	}
}

func UpdateNotIn(ins ...*In) UpdateOption {
	return func(options *UpdateOptions) {
		options.NotInWheres = append(options.NotInWheres, ins...)
	}
}

func Update(session *xorm.Session, bean interface{}, opts ...UpdateOption) (int64, error) {
	updateOptions := &UpdateOptions{}
	for _, opt := range opts {
		opt(updateOptions)
	}

	for _, where := range updateOptions.InWheres {
		if where != nil {
			session = session.In(where.Column, where.Args...)
		}
	}
	for _, where := range updateOptions.NotInWheres {
		if where != nil {
			session = session.NotIn(where.Column, where.Args...)
		}
	}
	if updateOptions.Wheres != nil {
		session = session.Where(updateOptions.Wheres, updateOptions.Args...)
	}
	if len(updateOptions.Cols) > 0 {
		session = session.Cols(updateOptions.Cols...)
	} else {
		session = session.AllCols()
	}

	return session.Update(bean)
}

/// InsertMulti Execute

type InsertMultiOption func(*InsertMultiOptions)
type InsertMultiOptions struct {
	StepNumber  int
	CheckNumber bool
}

func InsertMultiStepNumber(number int) InsertMultiOption {
	return func(options *InsertMultiOptions) {
		options.StepNumber = number
	}
}

func InsertMultiCheckNumber(check ...bool) InsertMultiOption {
	return func(options *InsertMultiOptions) {
		if len(check) > 0 {
			options.CheckNumber = check[0]
		}
		options.CheckNumber = true
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

		if options.CheckNumber && count != onesLen {
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

	InWheres    []*In
	NotInWheres []*In
}

func DeleteWheres(wheres interface{}) DeleteOption {
	return func(options *DeleteOptions) {
		switch ts := wheres.(type) {
		case []string:
			options.Wheres = GetWheres(strings.Join(ts, " AND "))
		default:
			options.Wheres = wheres
		}
	}
}

func DeleteArgs(args ...interface{}) DeleteOption {
	return func(options *DeleteOptions) {
		options.Args = args
	}
}

func DeleteIn(ins ...*In) DeleteOption {
	return func(options *DeleteOptions) {
		options.InWheres = append(options.InWheres, ins...)
	}
}

func DeleteNotIn(ins ...*In) DeleteOption {
	return func(options *DeleteOptions) {
		options.NotInWheres = append(options.NotInWheres, ins...)
	}
}

func Delete(session *xorm.Session, bean interface{}, opts ...DeleteOption) (int64, error) {
	deleteOptions := &DeleteOptions{}
	for _, opt := range opts {
		opt(deleteOptions)
	}

	for _, where := range deleteOptions.InWheres {
		if where != nil {
			session = session.In(where.Column, where.Args...)
		}
	}
	for _, where := range deleteOptions.NotInWheres {
		if where != nil {
			session = session.NotIn(where.Column, where.Args...)
		}
	}
	if deleteOptions.Wheres != nil {
		session = session.Where(deleteOptions.Wheres, deleteOptions.Args...)
	}

	return session.Delete(bean)
}
