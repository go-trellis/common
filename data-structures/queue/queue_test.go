package queue_test

import (
	"testing"

	"trellis.tech/trellis/common.v2/testutils"

	"trellis.tech/trellis/common.v2/data-structures/queue"
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
