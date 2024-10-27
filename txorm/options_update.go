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
	"strings"
)

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