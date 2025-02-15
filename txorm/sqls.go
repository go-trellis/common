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

	"xorm.io/xorm"
)

// Get retrieves a single record from the database.
func Get(session *xorm.Session, bean interface{}, opts ...GetOption) (ok bool, err error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	instrumentQuery(reflect.TypeOf(bean).Name(), "get", func() {
		ok, err = getOptions.Session(session).Get(bean)
	})
	return
}

// Find retrieves multiple records from the database.
func Find(session *xorm.Session, bean interface{}, opts ...GetOption) (err error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	instrumentQuery(reflect.TypeOf(bean).Name(), "find", func() {
		err = getOptions.Session(session).Find(bean)
	})
	return
}

// FindAndCount retrieves multiple records from the database and counts them.
func FindAndCount(session *xorm.Session, bean interface{}, opts ...GetOption) (c int64, err error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	instrumentQuery(reflect.TypeOf(bean).Name(), "find_and_count", func() {
		c, err = getOptions.Session(session).FindAndCount(bean)
	})
	return
}

// Count counts the number of records in the database.
func Count(session *xorm.Session, bean interface{}, opts ...GetOption) (c int64, err error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	instrumentQuery(reflect.TypeOf(bean).Name(), "count", func() {
		c, err = getOptions.Session(session).Count(bean)
	})
	return
}

// Update updates records in the database.
func Update(session *xorm.Session, bean interface{}, opts ...UpdateOption) (c int64, err error) {
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

	instrumentQuery(reflect.TypeOf(bean).Name(), "update", func() {
		c, err = session.Update(bean)
	})
	return
}

// InsertMultiOptions defines options for inserting multiple records.
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

// Insert insert data
func Insert(session *xorm.Session, beans ...interface{}) (c int64, err error) {
	instrumentQuery(reflect.TypeOf(beans).Name(), "insert", func() {
		c, err = session.Insert(beans...)
	})
	return
}

// InsertMulti insert data in batches with step number and check number
func InsertMulti(session *xorm.Session, ones interface{}, opts ...InsertMultiOption) (c int64, err error) {
	// 初始化选项
	options := &InsertMultiOptions{}
	for _, opt := range opts {
		opt(options)
	}
	instrumentQuery(reflect.TypeOf(ones).Name(), "insert_multi", func() {
		c, err = insertMulti(session, ones, options)
	})
	return
}

// insertMulti is the internal implementation of InsertMulti
func insertMulti(session *xorm.Session, ones interface{}, options *InsertMultiOptions) (int64, error) {

	// Extract the step number from the options if provided, otherwise use 1
	if options.StepNumber <= 0 {
		return session.InsertMulti(ones)
	}

	// get the length of the slice or array
	sliceOnes := reflect.Indirect(reflect.ValueOf(ones))
	kind := sliceOnes.Kind()

	// handle different types of slices or arrays
	if kind == reflect.Slice || kind == reflect.Array {
		onesLen := sliceOnes.Len()
		if onesLen == 0 {
			return 0, nil
		}

		// if the first element is an interface, disable auto-time
		if sliceOnes.Index(0).Kind() == reflect.Interface {
			session = session.NoAutoTime()
		}

		// if the length of the slice or array is less than or equal to the step number, insert all at once
		if onesLen <= options.StepNumber {
			return session.InsertMulti(ones)
		}

		// divide the slice or array into multiple batches and insert each batch separately
		var count int
		for i := 0; i < onesLen; i += options.StepNumber {
			end := i + options.StepNumber
			if end > onesLen {
				end = onesLen
			}

			// rebuild the slice or array for the current batch
			var multi []interface{}
			for j := i; j < end; j++ {
				multi = append(multi, sliceOnes.Index(j).Interface())
			}

			// insert the current batch separately
			session = session.NoAutoTime()
			n, err := session.InsertMulti(multi)
			if err != nil {
				return int64(count) + n, err
			}
			count += int(n)
		}

		// check if the number of inserted rows matches the expected number
		if options.CheckNumber && count != onesLen {
			return 0, fmt.Errorf("insert number not %d, but %d", onesLen, count)
		}
		return int64(count), nil
	}

	return session.InsertMulti(ones)
}

// Delete deletes the specified bean from the database. It supports various options such as WHERE conditions and batch size. The function returns the number of rows deleted and an error if any occurred.
func Delete(session *xorm.Session, bean interface{}, opts ...DeleteOption) (c int64, err error) {
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

	instrumentQuery(reflect.TypeOf(bean).Name(), "delete", func() {
		c, err = session.Delete(bean)
	})
	return
}
