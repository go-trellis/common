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
