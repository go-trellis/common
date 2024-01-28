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

	"trellis.tech/common.v2/errcode"
)

// Benchmarks Presence Update event with fake data.
func TestNext(t *testing.T) {

	worker, _ := NewWorker(NodeID(1))

	for i := 0; i < 10; i++ {
		if i < 5 {
			worker.Next()
		}
		if i > 5 {
			worker.NextSleep()
		}
	}
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
