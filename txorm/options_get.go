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
	"fmt"
	"reflect"
	"strings"

	"xorm.io/xorm"
)

type LinkType string

const (
	LinkTypeNotSet LinkType = ""
	LinkTypeAND    LinkType = "AND"
	LinkTypeOR     LinkType = "OR"
)

type Builder struct {
	LinkType LinkType
	Where    any
	Args     []any
}

type GetOption func(*GetOptions)
type GetOptions struct {
	Wheres any
	Args   []any

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

// GetBuilder 返回一个 GetOption 类型的函数，用于设置 GetOptions 的 Builders 字段
func GetBuilder(b *Builder) GetOption {
	return func(options *GetOptions) {
		// 如果 LinkType 未设置，则默认设置为 LinkTypeAND
		if b.LinkType == LinkTypeNotSet {
			b.LinkType = LinkTypeAND
		}

		// 检查 LinkType 是否为支持的类型，如果不是则抛出错误
		if b.LinkType != LinkTypeAND && b.LinkType != LinkTypeOR {
			panic(fmt.Errorf("not supported link type %s", b.LinkType))
		}

		// 将 Builder 添加到 GetOptions 的 Builders 列表中
		options.Builders = append(options.Builders, b)
	}
}

func GetWheres(wheres any) GetOption {
	return func(options *GetOptions) {
		switch ts := wheres.(type) {
		case string:
			options.addStringWheres(ts)
		case []string:
			options.addStringWheres(strings.Join(ts, " AND "))
		case map[string]any:
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

func GetArgs(args ...any) GetOption {
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
	if len(p.Distinct) > 0 {
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

func (p *GetOptions) addMapWheres(maps map[string]any) {
	if maps == nil {
		return
	}

	if p.Wheres == nil {
		p.Wheres = maps
		return
	}

	switch t := p.Wheres.(type) {
	case map[string]any:
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
