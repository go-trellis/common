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

package snowflake

import (
	"testing"

	"github.com/go-trellis/common/errors/errcode"
	"github.com/go-trellis/common/utils/testutils"
)

func TestNext(t *testing.T) {
	worker, err := NewWorker(NodeID(1))
	testutils.Ok(t, err)
	testutils.Assert(t, worker != nil, "worker should not be nil")

	for i := 0; i < 10; i++ {
		if i < 5 {
			id := worker.Next()
			testutils.Assert(t, id > 0, "id should be greater than 0")
		}
		if i > 5 {
			id := worker.NextSleep()
			testutils.Assert(t, id > 0, "id should be greater than 0")
		}
	}
}

func TestNewWorker_Default(t *testing.T) {
	worker, err := NewWorker()
	testutils.Ok(t, err)
	testutils.Assert(t, worker != nil, "worker should not be nil")

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

func TestNewWorker_WithOptions(t *testing.T) {
	worker, err := NewWorker(
		NodeID(10),
		Epoch(1504426600000),
		MaxBits(63),
		SequenceBits(12),
		NodesBits(10),
	)
	testutils.Ok(t, err)
	testutils.Assert(t, worker != nil, "worker should not be nil")

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

func TestNewWorker_InvalidNodeID(t *testing.T) {
	// Test nodeID greater than max
	_, err := NewWorker(NodeID(1025)) // Max nodeID is 1023 (2^10 - 1)
	testutils.NotOk(t, err, "should return error for invalid nodeID")

	// Test negative nodeID
	_, err = NewWorker(NodeID(-1))
	testutils.NotOk(t, err, "should return error for negative nodeID")
}

func TestNewWorker_InvalidBits(t *testing.T) {
	// Test bits sum greater than maxBits
	_, err := NewWorker(SequenceBits(40), NodesBits(40)) // 40+40=80 > 63
	testutils.NotOk(t, err, "should return error for bits sum greater than maxBits")
}

func TestNewWorker_InvalidEpoch(t *testing.T) {
	// Test epoch with invalid length
	_, err := NewWorker(Epoch(12345)) // Invalid length
	testutils.NotOk(t, err, "should return error for invalid epoch length")
}

func TestNewWorker_Epoch_Seconds(t *testing.T) {
	// Test epoch in seconds (10 digits)
	epoch := int64(1504426600) // 10 digits
	worker, err := NewWorker(Epoch(epoch))
	testutils.Ok(t, err)
	testutils.Assert(t, worker != nil, "worker should not be nil")

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

func TestNewWorker_Epoch_Microseconds(t *testing.T) {
	// Test epoch in microseconds (16 digits)
	// Use a more recent epoch to avoid negative timestamps
	epoch := int64(1609459200000000) // 2021-01-01 in microseconds
	worker, err := NewWorker(Epoch(epoch))
	testutils.Ok(t, err)
	testutils.Assert(t, worker != nil, "worker should not be nil")

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

func TestID_String(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	str := id.String()
	testutils.Assert(t, len(str) > 0, "String should not be empty")
}

func TestParseString(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	str := id.String()
	parsed, err := ParseString(str)
	testutils.Ok(t, err)
	testutils.Equals(t, id, parsed, "parsed ID should match original")
}

func TestParseString_Invalid(t *testing.T) {
	_, err := ParseString("invalid")
	testutils.NotOk(t, err, "should return error for invalid string")
}

func TestID_Base2(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	base2 := id.Base2()
	testutils.Assert(t, len(base2) > 0, "Base2 should not be empty")
}

func TestParseBase2(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	base2 := id.Base2()
	parsed, err := ParseBase2(base2)
	testutils.Ok(t, err)
	testutils.Equals(t, id, parsed, "parsed Base2 ID should match original")
}

func TestParseBase2_Invalid(t *testing.T) {
	_, err := ParseBase2("invalid")
	testutils.NotOk(t, err, "should return error for invalid Base2 string")
}

func TestID_Base64(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	base64Str := id.Base64()
	testutils.Assert(t, len(base64Str) > 0, "Base64 should not be empty")
}

func TestParseBase64(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	base64Str := id.Base64()
	parsed, err := ParseBase64(base64Str)
	testutils.Ok(t, err)
	testutils.Equals(t, id, parsed, "parsed Base64 ID should match original")
}

func TestParseBase64_Invalid(t *testing.T) {
	_, err := ParseBase64("invalid!!!")
	testutils.NotOk(t, err, "should return error for invalid Base64 string")
}

func TestID_Bytes(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	bytes := id.Bytes()
	testutils.Assert(t, len(bytes) > 0, "Bytes should not be empty")
}

func TestParseBytes(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	id := worker.Next()

	bytes := id.Bytes()
	parsed, err := ParseBytes(bytes)
	testutils.Ok(t, err)
	testutils.Equals(t, id, parsed, "parsed Bytes ID should match original")
}

func TestParseBytes_Invalid(t *testing.T) {
	_, err := ParseBytes([]byte("invalid"))
	testutils.NotOk(t, err, "should return error for invalid bytes")
}

func TestWorker_GetEpochTime(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	epochTime := worker.GetEpochTime()
	testutils.Assert(t, epochTime > 0, "epoch time should be greater than 0")
}

func TestWorker_Next_Unique(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	ids := make(map[ID]bool)

	for i := 0; i < 100; i++ {
		id := worker.Next()
		testutils.Assert(t, !ids[id], "ID should be unique")
		ids[id] = true
	}
}

func TestWorker_NextSleep_Unique(t *testing.T) {
	worker, _ := NewWorker(NodeID(1))
	ids := make(map[ID]bool)

	for i := 0; i < 100; i++ {
		id := worker.NextSleep()
		testutils.Assert(t, !ids[id], "ID should be unique")
		ids[id] = true
	}
}

func TestWorker_MaxNodeID(t *testing.T) {
	// Test with maximum valid nodeID
	worker, err := NewWorker(NodeID(1023), NodesBits(10)) // 2^10 - 1 = 1023
	testutils.Ok(t, err)

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

func TestWorker_ZeroNodeID(t *testing.T) {
	worker, err := NewWorker(NodeID(0))
	testutils.Ok(t, err)

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

func TestWorker_NoSequenceBits(t *testing.T) {
	worker, err := NewWorker(NodesBits(10), SequenceBits(0))
	testutils.Ok(t, err)

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

func TestWorker_NoNodesBits(t *testing.T) {
	worker, err := NewWorker(NodesBits(0), SequenceBits(12))
	testutils.Ok(t, err)

	id := worker.Next()
	testutils.Assert(t, id > 0, "id should be greater than 0")
}

// Benchmarks Presence Update event with fake data.
func BenchmarkNext(b *testing.B) {
	worker, _ := NewWorker(NodeID(1))

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		worker.Next()
	}
}

func BenchmarkNextMaxSequence(b *testing.B) {
	worker, _ := NewWorker(NodesBits(0), SequenceBits(22))

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		worker.Next()
	}
}

func BenchmarkNextNoSequence(b *testing.B) {
	worker, _ := NewWorker(NodesBits(10), SequenceBits(0))

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		worker.Next()
	}
}

func BenchmarkNextSleep(b *testing.B) {
	worker, _ := NewWorker(NodeID(1))

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		id := worker.NextSleep()
		if id <= 0 {
			panic(errcode.New("id must above 0"))
		}
	}
}

func BenchmarkNextSleepMaxSequence(b *testing.B) {
	worker, _ := NewWorker(NodesBits(0), SequenceBits(22))

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		worker.NextSleep()
	}
}

func BenchmarkNextSleepNoSequence(b *testing.B) {
	worker, _ := NewWorker(NodesBits(10), SequenceBits(0))

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		worker.NextSleep()
	}
}
