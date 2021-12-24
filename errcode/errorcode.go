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

package errcode

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"trellis.tech/trellis/common.v0.1"
	"trellis.tech/trellis/common.v0.1/json"
)

const (
	namespace = "E"
)

// ErrorCode Error functions
type ErrorCode interface {
	SimpleError
	Code() uint64
	Context() ErrorContext
	Append(err ...error) ErrorCode
	WithContext(k string, v interface{}) ErrorCode
}

// OptionFunc set params
type OptionFunc func(*ErrorOptions)

// ErrorOptions error parmas
type ErrorOptions struct {
	namespace string
	id        string
	code      uint64
	message   string

	errs []error

	ctx map[string]interface{}
}

// NewSimpleError new simple errcode by options
func (p *ErrorOptions) NewSimpleError() SimpleError {
	return &Error{namespace: p.namespace, id: p.id, message: p.message}
}

// OptionID set id into options
func OptionID(id string) OptionFunc {
	return func(p *ErrorOptions) {
		p.id = id
	}
}

// OptionCode set error code into options
func OptionCode(code uint64) OptionFunc {
	return func(p *ErrorOptions) {
		p.code = code
	}
}

// OptionNamespace set error code into options
func OptionNamespace(ns string) OptionFunc {
	return func(p *ErrorOptions) {
		p.namespace = ns
	}
}

// OptionMesssage set error message into options
func OptionMesssage(msg string) OptionFunc {
	return func(p *ErrorOptions) {
		p.message = msg
	}
}

// OptionErrs append errs
func OptionErrs(errs ...error) OptionFunc {
	return func(p *ErrorOptions) {
		p.errs = append(p.errs, errs...)
	}
}

// OptionContext set error context into options
func OptionContext(ctx map[string]interface{}) OptionFunc {
	return func(p *ErrorOptions) {
		p.ctx = ctx
	}
}

type errorCode struct {
	err  SimpleError
	code uint64

	context map[string]interface{}
	errors  []error
}

// NewErrorCode get a new error code
func NewErrorCode(ofs ...OptionFunc) ErrorCode {
	opts := &ErrorOptions{}
	for _, o := range ofs {
		o(opts)
	}
	if opts.namespace == "" {
		opts.namespace = common.FormatNamespaceString(namespace)
	}
	if opts.id == "" {
		opts.id = uuid.New().String()
	}

	if opts.ctx == nil {
		opts.ctx = make(map[string]interface{})
	}

	ec := &errorCode{
		err:     opts.NewSimpleError(),
		context: opts.ctx,
	}

	ec.errors = append(ec.errors, opts.errs...)

	return ec
}

func (p *errorCode) Append(errs ...error) ErrorCode {
	p.errors = append(p.errors, errs...)
	return p
}

func (p *errorCode) Code() uint64 {
	return p.code
}

func (p *errorCode) Message() string {
	return p.err.Message()
}

func (p *errorCode) Context() ErrorContext {
	return p.context
}

func (p *errorCode) Error() string {
	var errs = []string{p.err.Error()}
	for _, err := range p.errors {
		errs = append(errs, err.Error())
	}
	return strings.Join(errs, "; ")
}

func (p *errorCode) FullError() string {
	return strings.Join(
		append([]string{},
			fmt.Sprintf("ID:%s#%s", genErrorCodeKey(p.Namespace(), p.Code()), p.ID()),
			"Error:", p.Error(),
			"Context:", p.Context().Error(),
		), "\n")
}

func (p *errorCode) ID() string {
	return p.err.ID()
}

func (p *errorCode) Namespace() string {
	return p.err.Namespace()
}

func (p *errorCode) WithContext(key string, value interface{}) ErrorCode {
	p.context[key] = value
	return p
}

// ErrorContext map contexts
type ErrorContext map[string]interface{}

func (p ErrorContext) Error() string {
	if p == nil {
		return ""
	}

	if bs, e := json.Marshal(p); e == nil {
		return string(bs)
	}
	return ""
}

func genErrorCodeKey(namespace string, code uint64) string {
	return fmt.Sprintf("%s:%d", namespace, code)
}
