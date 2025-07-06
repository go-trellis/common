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

/// Get Execute

func Get(session *xorm.Session, bean any, opts ...GetOption) (bool, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).Get(bean)
}

func Find(session *xorm.Session, bean any, opts ...GetOption) error {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).Find(bean)
}

func FindAndCount(session *xorm.Session, bean any, opts ...GetOption) (int64, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).FindAndCount(bean)
}

func Count(session *xorm.Session, bean any, opts ...GetOption) (int64, error) {
	getOptions := &GetOptions{}
	for _, opt := range opts {
		opt(getOptions)
	}

	return getOptions.Session(session).Count(bean)
}

// Update Execute

func Update(session *xorm.Session, bean any, opts ...UpdateOption) (int64, error) {
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

func Insert(session *xorm.Session, beans ...any) (int64, error) {
	return session.Insert(beans...)
}

// InsertMulti insert multi seperated slice data in a big slice with every step number
// default to insert the slice with no seperated.
func InsertMulti(session *xorm.Session, ones any, opts ...InsertMultiOption) (int64, error) {
	// 初始化选项
	options := &InsertMultiOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 如果步长小于等于0，直接插入
	if options.StepNumber <= 0 {
		return session.InsertMulti(ones)
	}

	// 获取ones的反射值
	sliceOnes := reflect.Indirect(reflect.ValueOf(ones))
	kind := sliceOnes.Kind()

	// 处理slice或array类型
	if kind == reflect.Slice || kind == reflect.Array {
		onesLen := sliceOnes.Len()
		if onesLen == 0 {
			return 0, nil
		}

		// 如果第一个元素是接口类型，禁用自动时间
		if sliceOnes.Index(0).Kind() == reflect.Interface {
			session = session.NoAutoTime()
		}

		// 如果长度小于等于步长，直接插入
		if onesLen <= options.StepNumber {
			return session.InsertMulti(ones)
		}

		// 分批插入
		var count int
		for i := 0; i < onesLen; i += options.StepNumber {
			end := i + options.StepNumber
			if end > onesLen {
				end = onesLen
			}

			// 构建当前批次的插入数据
			var multi []any
			for j := i; j < end; j++ {
				multi = append(multi, sliceOnes.Index(j).Interface())
			}

			// 插入当前批次数据
			session = session.NoAutoTime()
			n, err := session.InsertMulti(multi)
			if err != nil {
				return int64(count) + n, err
			}
			count += int(n)
		}

		// 检查插入数量是否一致
		if options.CheckNumber && count != onesLen {
			return 0, fmt.Errorf("insert number not %d, but %d", onesLen, count)
		}
		return int64(count), nil
	}

	// 默认处理
	return session.InsertMulti(ones)
}

/// Delete Execute

func Delete(session *xorm.Session, bean any, opts ...DeleteOption) (int64, error) {
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
