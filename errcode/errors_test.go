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

	"trellis.tech/trellis/common.v1/testutils"
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
