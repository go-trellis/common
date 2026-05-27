/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

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

package logger

import (
	"fmt"
	"io"
	"reflect"

	"github.com/go-trellis/common/utils/json"
	"xorm.io/xorm/log"
)

// Logger is the logging interface
type Logger interface {
	log.Logger

	With(kvs ...any) Logger
	Writer() io.Writer
}

func toString(v any) string {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr, reflect.Struct, reflect.Map:
		bs, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(bs)
	case reflect.String:
		return v.(string)
	default:
		return fmt.Sprint(v)
	}
}
