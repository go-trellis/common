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
	"fmt"
	"reflect"

	"xorm.io/xorm"
)

// TransactionDo to do transaction with customer function
func TransactionDo(engine xorm.Engine, fn func(*xorm.Session) error) (err error) {
	s := engine.NewSession()
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

type UpdateOption func(*UpdateOptions)

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

type UpdateOptions struct {
	Wheres interface{}
	Args   []interface{}
	Cols   []string
}

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

// InsertMulti insert multi seperated data in a big data for every step
func InsertMulti(session *xorm.Session, ones interface{}, stepNum int) (int64, error) {
	if stepNum <= 0 {
		return session.InsertMulti(ones)
	}
	sliceOnes := reflect.Indirect(reflect.ValueOf(ones))
	switch sliceOnes.Kind() {
	case reflect.Slice:
		onesLen := sliceOnes.Len()
		if onesLen == 0 {
			return 0, nil
		}

		switch sliceOnes.Index(0).Kind() {
		case reflect.Interface:
			session = session.NoAutoTime()
		}

		if onesLen <= stepNum {
			_, err := session.InsertMulti(ones)
			return 0, err
		}

		loop, count, processNum := 0, 0, onesLen

		for i := 0; i < onesLen; i += stepNum {
			if processNum > stepNum {
				loop = i + stepNum
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
				return 0, err
			}
			count += int(n)
			processNum -= stepNum
		}

		if count != onesLen {
			return 0, fmt.Errorf("insert number not %d, but %d", onesLen, count)
		}
		return int64(count), nil
	}
	return session.InsertMulti(ones)
}
