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
	"testing"

	"github.com/go-trellis/common/utils/testutils"
)

func TestNewErrors(t *testing.T) {
	err1 := fmt.Errorf("error_test1")
	err2 := fmt.Errorf("error_test2")
	err3 := fmt.Errorf("error_test3")

	errs := NewErrors(err1, err2)

	testutils.NotOk(t, errs)
	testutils.Assert(t, errs.Error() != "", "new errcode failed")
	testutils.Assert(t, errs.Error() == "error_test1;error_test2", "incorrect errcode")

	errs = errs.Append(err3)
	testutils.Assert(t, errs.Error() == "error_test1;error_test2;error_test3", "incorrect errcode:%s", errs.Error())
}

func TestNewErrors_Empty(t *testing.T) {
	errs := NewErrors()
	testutils.Assert(t, errs.Errors() == nil, "empty errors should return nil")
	testutils.Assert(t, errs.Error() == "", "empty errors should have empty error string")
}

func TestErrors_Errors(t *testing.T) {
	err1 := fmt.Errorf("error1")
	errs := NewErrors(err1)

	result := errs.Errors()
	testutils.NotOk(t, result, "should return error when not empty")

	emptyErrs := NewErrors()
	result = emptyErrs.Errors()
	testutils.Assert(t, result == nil, "should return nil for empty errors")
}

func TestErrors_ErrorString(t *testing.T) {
	err1 := fmt.Errorf("error1")
	err2 := fmt.Errorf("error2")
	errs := NewErrors(err1, err2)

	errorStr := errs.Error()
	testutils.Assert(t, errorStr != "", "error string should not be empty")
	testutils.Assert(t, len(errorStr) > 0, "error string should have content")
}

func TestErrors_Append(t *testing.T) {
	err1 := fmt.Errorf("error1")
	err2 := fmt.Errorf("error2")
	err3 := fmt.Errorf("error3")

	errs := NewErrors(err1)
	errs = errs.Append(err2, err3)

	testutils.Assert(t, len(errs) == 3, "should have 3 errors")
}

func TestErrorsString_ErrorCode(t *testing.T) {
	ec := NewErrorCode(OptionMessage("test error"), OptionCode(1001))
	errs := NewErrors(ec)

	errStr := errs.Error()
	testutils.Assert(t, errStr != "", "error string should not be empty")
	testutils.Assert(t, len(errStr) > 0, "error string should have content")
}

func TestErrorsString_SimpleError(t *testing.T) {
	se := New("test error")
	errs := NewErrors(se)

	errStr := errs.Error()
	testutils.Assert(t, errStr != "", "error string should not be empty")
	testutils.Assert(t, len(errStr) > 0, "error string should have content")
}

func TestErrorsString_Empty(t *testing.T) {
	errs := []error{}
	result := errorsString(errs...)
	testutils.Assert(t, result == nil, "should return nil for empty errors")
}

func TestErrorsString_Mixed(t *testing.T) {
	ec := NewErrorCode(OptionMessage("error code"))
	se := New("simple error")
	stdErr := fmt.Errorf("standard error")

	result := errorsString(ec, se, stdErr)
	testutils.Assert(t, len(result) == 3, "should have 3 error strings")
	testutils.Assert(t, result[0] != "", "first error string should not be empty")
	testutils.Assert(t, result[1] != "", "second error string should not be empty")
	testutils.Assert(t, result[2] != "", "third error string should not be empty")
}
