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

package transaction

import (
	"fmt"
	"reflect"

	"github.com/go-trellis/common.v3/errors/errcode"
)

// LogicFunc logic functions
type LogicFunc struct {
	BeforeLogic any
	AfterLogic  any
	OnError     any
	Logic       any
	AfterCommit any
}

// TXFunc transaction function type
type TXFunc func(repos ...any) error

// Function Flags
const (
	Logic = iota
	BeforeLogic
	AfterLogic
	OnError
	AfterCommit
)

// MapErrorTypes supported error types
var mapErrorTypes = map[reflect.Type]bool{
	// common error
	reflect.TypeOf((*error)(nil)).Elem(): true,
	// extended errors
	reflect.TypeOf((*errcode.ErrorCode)(nil)).Elem():   true,
	reflect.TypeOf((*errcode.SimpleError)(nil)).Elem(): true,
	reflect.TypeOf((*errcode.Errors)(nil)).Elem():      true,
}

// AddErrorTypes add new error types to supported list.
func AddErrorTypes(errType reflect.Type) {
	mapErrorTypes[errType] = true
}

// CallFunc execute transaction function with logic functions and args
func CallFunc(input any, args ...any) ([]any, error) {
	if input == nil {
		return nil, nil
	}

	switch _logicFunc := input.(type) {
	case TXFunc:
		{
			return nil, _logicFunc(args...)
		}
	case func(repos ...any) (err error):
		{
			return nil, _logicFunc(args...)
		}
	default:
		return call(input, args...)
	}
}

// GetLogicFunc get logic function from input any
func GetLogicFunc(input any) *LogicFunc {
	if input == nil {
		return nil
	}
	lFunc := &LogicFunc{}
	switch fn := input.(type) {
	case TXFunc, func(repos []any) (err error):
		{
			lFunc.Logic = fn
		}
	case map[int]any:
		if hookBeforeFn, exist := fn[BeforeLogic]; exist {
			lFunc.BeforeLogic = hookBeforeFn
		}
		if logicFn, exist := fn[Logic]; exist {
			lFunc.Logic = logicFn
		}
		if hookAfterFn, exist := fn[AfterLogic]; exist {
			lFunc.AfterLogic = hookAfterFn
		}
		if errFn, exist := fn[OnError]; exist {
			lFunc.OnError = errFn
		}
		if afterCommitFn, exist := fn[AfterCommit]; exist {
			lFunc.AfterCommit = afterCommitFn
		}
	default:
		lFunc.Logic = fn
	}

	return lFunc
}

func call(fn any, args ...any) ([]any, error) {
	v := reflect.ValueOf(fn)
	if !v.IsValid() {
		return nil, fmt.Errorf("call of nil")
	}
	typ := v.Type()
	if typ.Kind() != reflect.Func {
		return nil, fmt.Errorf("non-function of type %s", typ)
	}
	if err := goodFunc(typ); err != nil {
		return nil, err
	}
	numIn := typ.NumIn()
	var dddType reflect.Type
	if typ.IsVariadic() {
		if len(args) < numIn-1 {
			return nil, fmt.Errorf("wrong number of args: got %d want at least %d, type: %v", len(args), numIn-1, typ)
		}
		dddType = typ.In(numIn - 1).Elem()
	} else {
		if len(args) != numIn {
			return nil, fmt.Errorf("wrong number of args: got %d want %d, type: %v", len(args), numIn, typ)
		}
	}
	argv := make([]reflect.Value, len(args))
	for i, arg := range args {
		value := reflect.ValueOf(arg)
		// Compute the expected type. Clumsy because of variadics.
		var argType reflect.Type
		if !typ.IsVariadic() || i < numIn-1 {
			argType = typ.In(i)
		} else {
			argType = dddType
		}

		var err error
		if argv[i], err = prepareArg(value, argType); err != nil {
			return nil, fmt.Errorf("arg %d: %s", i, err)
		}
	}

	result := v.Call(argv)
	resultLen := len(result)

	var resultValues []any

	for _, v := range result {
		resultValues = append(resultValues, v.Interface())
	}

	if resultLen == 1 {
		if resultValues[0] != nil {
			return nil, resultValues[0].(error)
		}
	} else if resultLen > 1 {
		if resultValues[resultLen-1] != nil {
			return resultValues[0 : resultLen-1], resultValues[resultLen-1].(error)
		}
		return resultValues[0 : resultLen-1], nil
	}

	return nil, nil
}

// goodFunc checks if the function has a single return value or multiple return values where the last one is an error.
func goodFunc(typ reflect.Type) error {
	if typ.NumOut() == 0 ||
		(typ.NumOut() > 0 && mapErrorTypes[typ.Out(typ.NumOut()-1)]) {
		return nil
	}

	return errcode.New("function must have a single return value or multiple return values where the last one is an error")
}

// prepareArg prepares the argument for the function call. It ensures that the argument is valid and assignable to the expected type.
func prepareArg(value reflect.Value, argType reflect.Type) (reflect.Value, error) {
	if !value.IsValid() {
		if err := canBeNil(argType); err != nil {
			return reflect.Value{}, err
		}
		value = reflect.Zero(argType)
	}
	if !value.Type().AssignableTo(argType) {
		return reflect.Value{}, fmt.Errorf("value has type %s; should be %s", value.Type(), argType)
	}
	return value, nil
}

// canBeNil checks if the value can be nil.
func canBeNil(typ reflect.Type) error {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return nil
	}
	return fmt.Errorf("value is nil; should be of type %s", typ.String())
}
