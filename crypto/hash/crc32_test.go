/*
Copyright © 2016 Henry Huang <hhh@rutcode.com>

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

package hash

import (
	"hash/crc32"
	"testing"

	"trellis.tech/trellis/common.v3/testutils"
)

func TestNewCRC32(t *testing.T) {
	table := crc32.IEEETable
	repo := NewCRC32(table)
	testutils.Assert(t, repo != nil, "NewCRC32 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewCRCIEEE(t *testing.T) {
	repo := NewCRCIEEE()
	testutils.Assert(t, repo != nil, "NewCRCIEEE should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewCRC32WithOptions_WithTable(t *testing.T) {
	table := crc32.IEEETable
	repo, err := NewCRC32WithOptions(CRC32OptionTable(table))
	testutils.Ok(t, err)
	testutils.Assert(t, repo != nil, "NewCRC32WithOptions should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewCRC32WithOptions_WithPoly(t *testing.T) {
	poly := uint32(crc32.IEEE)
	repo, err := NewCRC32WithOptions(CRC32OptionPoly(poly))
	testutils.Ok(t, err)
	testutils.Assert(t, repo != nil, "NewCRC32WithOptions should return non-nil with poly")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewCRC32WithOptions_Default(t *testing.T) {
	repo, err := NewCRC32WithOptions()
	testutils.Ok(t, err)
	testutils.Assert(t, repo != nil, "NewCRC32WithOptions should return non-nil with default")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestCRC32OptionTable(t *testing.T) {
	table := crc32.IEEETable
	opt := CRC32OptionTable(table)
	testutils.Assert(t, opt != nil, "CRC32OptionTable should return option function")

	options := &CRC32Options{}
	opt(options)
	testutils.Assert(t, options.Table == table, "CRC32OptionTable should set table")
}

func TestCRC32OptionPoly(t *testing.T) {
	poly := uint32(crc32.IEEE)
	opt := CRC32OptionPoly(poly)
	testutils.Assert(t, opt != nil, "CRC32OptionPoly should return option function")

	options := &CRC32Options{}
	opt(options)
	testutils.Equals(t, poly, options.Poly, "CRC32OptionPoly should set poly")
}

func TestMakeCRC32Table(t *testing.T) {
	poly := uint32(crc32.IEEE)
	table := MakeCRC32Table(poly)
	testutils.Assert(t, table != nil, "MakeCRC32Table should return non-nil table")
}

func TestHash32Repo_Sum32(t *testing.T) {
	repo := NewCRCIEEE()
	testutils.Assert(t, repo != nil, "NewCRCIEEE should return non-nil")

	hash32Repo, ok := repo.(Hash32Repo)
	testutils.Assert(t, ok, "should implement Hash32Repo")

	result, err := hash32Repo.Sum32([]byte("test"))
	testutils.Ok(t, err)
	testutils.Assert(t, result > 0, "Sum32 should return non-zero value")
}

func TestHash32Repo_SumBytes(t *testing.T) {
	repo := NewCRCIEEE()
	result := repo.SumBytes([]byte("test"))
	testutils.Assert(t, len(result) > 0, "SumBytes should return non-empty result")
}

func TestHash32Repo_SumTimes(t *testing.T) {
	repo := NewCRCIEEE()
	result := repo.SumTimes("test", 2)
	testutils.Assert(t, len(result) > 0, "SumTimes should return non-empty result")

	result2 := repo.SumTimes("test", 0)
	testutils.Assert(t, result2 == "", "SumTimes(0) should return empty string")
}

func TestHash32Repo_SumBytesTimes(t *testing.T) {
	repo := NewCRCIEEE()
	result := repo.SumBytesTimes([]byte("test"), 2)
	testutils.Assert(t, len(result) > 0, "SumBytesTimes should return non-empty result")
}

func TestHash32Repo_Reset(t *testing.T) {
	repo := NewCRCIEEE()
	repo.Sum("test1")
	repo.Reset()
	result := repo.Sum("test2")
	testutils.Assert(t, len(result) > 0, "Sum after Reset should work")
}

func TestHash32Repo_Consistency(t *testing.T) {
	repo := NewCRCIEEE()
	result1 := repo.Sum("test")
	result2 := repo.Sum("test")
	testutils.Equals(t, result1, result2, "Same input should produce same hash")
}

func TestHash32Repo_Sum32_Consistency(t *testing.T) {
	hash32Repo := NewCRCIEEE()

	result1, err1 := hash32Repo.Sum32([]byte("test"))
	testutils.Ok(t, err1)

	// Create a new repo for the second call since Sum32 modifies internal state
	hash32Repo2 := NewCRCIEEE()
	result2, err2 := hash32Repo2.Sum32([]byte("test"))
	testutils.Ok(t, err2)

	testutils.Equals(t, result1, result2, "Same input should produce same Sum32 result")
}

