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

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestNewErrorCode(t *testing.T) {
	// Test with no options
	ec := NewErrorCode()
	testutils.Assert(t, ec != nil, "error code should not be nil")
	testutils.Assert(t, ec.Code() == 0, "default code should be 0")
	testutils.Assert(t, ec.ID() != "", "id should not be empty")
	testutils.Assert(t, ec.Namespace() != "", "namespace should not be empty")
	testutils.Assert(t, ec.Message() == "", "default message should be empty")

	// Test with all options
	customID := "test-id-123"
	customCode := uint64(1001)
	customNamespace := "TEST"
	customMessage := "test error message"
	customCtx := map[string]any{"key1": "value1", "key2": 42}

	ec = NewErrorCode(
		OptionID(customID),
		OptionCode(customCode),
		OptionNamespace(customNamespace),
		OptionMessage(customMessage),
		OptionContext(customCtx),
	)

	testutils.Equals(t, ec.ID(), customID)
	testutils.Equals(t, ec.Code(), customCode)
	testutils.Equals(t, ec.Namespace(), customNamespace)
	testutils.Equals(t, ec.Message(), customMessage)
	testutils.Assert(t, ec.Context() != nil, "context should not be nil")
}

func TestErrorCode_Append(t *testing.T) {
	ec := NewErrorCode(OptionMessage("base error"))
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")

	ec = ec.Append(err1, err2)
	errorStr := ec.Error()
	testutils.Assert(t, errorStr != "", "error string should not be empty")
	testutils.Assert(t, len(ec.(*errorCode).errors) == 2, "should have 2 appended errors")
}

func TestErrorCode_WithContext(t *testing.T) {
	ec := NewErrorCode(OptionMessage("test"))
	ec = ec.WithContext("key1", "value1")
	ec = ec.WithContext("key2", 42)

	ctx := ec.Context()
	testutils.Assert(t, ctx != nil, "context should not be nil")
	testutils.Equals(t, ctx["key1"], "value1")
	testutils.Equals(t, ctx["key2"], 42)
}

func TestErrorCode_Error(t *testing.T) {
	ec := NewErrorCode(OptionMessage("base error"))
	errorStr := ec.Error()
	testutils.Equals(t, errorStr, "base error")

	// Test with appended errors
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")
	ec = ec.Append(err1, err2)
	errorStr = ec.Error()
	testutils.Assert(t, errorStr != "", "error string should not be empty")
	testutils.Assert(t, len(ec.(*errorCode).errors) == 2, "should have 2 errors")
}

func TestErrorCode_FullError(t *testing.T) {
	ec := NewErrorCode(
		OptionCode(1001),
		OptionMessage("test error"),
		OptionContext(map[string]any{"key": "value"}),
	)
	fullError := ec.FullError()
	testutils.Assert(t, fullError != "", "full error should not be empty")
	testutils.Assert(t, len(fullError) > 0, "full error should have content")
}

func TestErrorCode_OptionErrs(t *testing.T) {
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")
	ec := NewErrorCode(
		OptionMessage("base"),
		OptionErrs(err1, err2),
	)

	testutils.Assert(t, len(ec.(*errorCode).errors) == 2, "should have 2 errors")
	errorStr := ec.Error()
	testutils.Assert(t, errorStr != "", "error string should not be empty")
}

func TestErrorContext_Error(t *testing.T) {
	// Test nil context
	var nilCtx ErrorContext
	testutils.Equals(t, nilCtx.Error(), "")

	// Test valid context
	ctx := ErrorContext{"key1": "value1", "key2": 42}
	ctxStr := ctx.Error()
	testutils.Assert(t, ctxStr != "", "context error string should not be empty")
}

func TestErrorCode_AllMethods(t *testing.T) {
	ec := NewErrorCode(
		OptionID("test-id"),
		OptionCode(1001),
		OptionNamespace("TEST"),
		OptionMessage("test message"),
	)

	testutils.Equals(t, ec.ID(), "test-id")
	testutils.Equals(t, ec.Code(), uint64(1001))
	testutils.Equals(t, ec.Namespace(), "TEST")
	testutils.Equals(t, ec.Message(), "test message")
	testutils.Assert(t, ec.Context() != nil, "context should not be nil")
}
