/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

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

package queue_test

import (
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"

	"github.com/go-trellis/common.v3/storage/data-structures/queue"
)

func TestQueue(t *testing.T) {
	q := queue.New()

	q.Push(1)

	i, ok := q.Front()
	testutils.Equals(t, true, ok, "not get")
	testutils.Equals(t, 1, i, "not get 1")

	i, ok = q.End()
	testutils.Equals(t, true, ok, "not get")
	testutils.Equals(t, 1, i, "not get 1")

	i, ok = q.Pop()
	testutils.Equals(t, true, ok, "not get")
	testutils.Equals(t, 1, i, "not get 1")

	q.Push(2)
	q.Push(3)
	q.Push(4)
	iArr, ok := q.PopMany(2)
	testutils.Equals(t, true, ok, "not get")
	testutils.Equals(t, []any{2, 3}, iArr, "not get 2,3")
	iArr, ok = q.PopMany(2)
	testutils.Equals(t, true, ok, "not get")
	testutils.Equals(t, []any{4}, iArr, "not get 4")
	iArr, ok = q.PopMany(2)
	testutils.Equals(t, false, ok, "not get")
	testutils.Equals(t, []any(nil), iArr, "get data")

	q.PushMany(5, 6, 7, 8)
	testutils.Equals(t, int64(4), q.Length(), "length is 4")
	iArr, ok = q.PopMany(4)
	testutils.Equals(t, true, ok, "not get")
	testutils.Equals(t, []any{5, 6, 7, 8}, iArr, "not get 5,6,7,8")
}
