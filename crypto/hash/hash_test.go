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
	"crypto"
	"testing"

	"trellis.tech/trellis/common.v3/utils/testutils"
)

func TestNewHashRepo_MD5(t *testing.T) {
	repo := NewHashRepo(crypto.MD5)
	testutils.Assert(t, repo != nil, "NewHashRepo should return non-nil for MD5")
}

func TestNewHashRepo_SHA1(t *testing.T) {
	repo := NewHashRepo(crypto.SHA1)
	testutils.Assert(t, repo != nil, "NewHashRepo should return non-nil for SHA1")
}

func TestNewHashRepo_SHA256(t *testing.T) {
	repo := NewHashRepo(crypto.SHA256)
	testutils.Assert(t, repo != nil, "NewHashRepo should return non-nil for SHA256")
}

func TestNewHashRepo_SHA512(t *testing.T) {
	repo := NewHashRepo(crypto.SHA512)
	testutils.Assert(t, repo != nil, "NewHashRepo should return non-nil for SHA512")
}

func TestNewHashRepo_Invalid(t *testing.T) {
	repo := NewHashRepo(crypto.Hash(999))
	testutils.Assert(t, repo == nil, "NewHashRepo should return nil for invalid hash")
}

func TestNewMD5(t *testing.T) {
	repo := NewMD5()
	testutils.Assert(t, repo != nil, "NewMD5 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewSHA1(t *testing.T) {
	repo := NewSHA1()
	testutils.Assert(t, repo != nil, "NewSHA1 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewSHA224(t *testing.T) {
	repo := NewSHA224()
	testutils.Assert(t, repo != nil, "NewSHA224 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewSHA256(t *testing.T) {
	repo := NewSHA256()
	testutils.Assert(t, repo != nil, "NewSHA256 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewSHA384(t *testing.T) {
	repo := NewSHA384()
	testutils.Assert(t, repo != nil, "NewSHA384 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewSHA512(t *testing.T) {
	repo := NewSHA512()
	testutils.Assert(t, repo != nil, "NewSHA512 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewSHA512_224(t *testing.T) {
	repo := NewSHA512_224()
	testutils.Assert(t, repo != nil, "NewSHA512_224 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewSHA512_256(t *testing.T) {
	repo := NewSHA512_256()
	testutils.Assert(t, repo != nil, "NewSHA512_256 should return non-nil")

	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestDefaultHash_Sum(t *testing.T) {
	repo := NewMD5()
	result := repo.Sum("test")
	testutils.Assert(t, len(result) == 32, "MD5 should return 32 hex chars")
}

func TestDefaultHash_SumBytes(t *testing.T) {
	repo := NewMD5()
	result := repo.SumBytes([]byte("test"))
	testutils.Assert(t, len(result) == 32, "MD5 should return 32 hex chars")
}

func TestDefaultHash_SumTimes(t *testing.T) {
	repo := NewMD5()
	result := repo.SumTimes("test", 2)
	testutils.Assert(t, len(result) == 32, "SumTimes should return valid hash")

	result2 := repo.SumTimes("test", 0)
	testutils.Assert(t, result2 == "", "SumTimes(0) should return empty string")
}

func TestDefaultHash_Reset(t *testing.T) {
	repo := NewMD5()
	repo.Sum("test1")
	repo.Reset()
	result := repo.Sum("test2")
	testutils.Assert(t, len(result) > 0, "Sum after Reset should work")
}

func TestDefaultHash_Consistency(t *testing.T) {
	repo := NewMD5()
	result1 := repo.Sum("test")
	result2 := repo.Sum("test")
	testutils.Equals(t, result1, result2, "Same input should produce same hash")
}

func TestNewHashRepo_AllTypes(t *testing.T) {
	hashTypes := []crypto.Hash{
		crypto.MD5,
		crypto.SHA1,
		crypto.SHA224,
		crypto.SHA256,
		crypto.SHA384,
		crypto.SHA512,
		crypto.SHA512_224,
		crypto.SHA512_256,
	}

	for _, h := range hashTypes {
		repo := NewHashRepo(h)
		testutils.Assert(t, repo != nil, "NewHashRepo should return non-nil for %v", h)

		result := repo.Sum("test")
		testutils.Assert(t, len(result) > 0, "Sum should return non-empty result for %v", h)
	}
}

func TestDefaultHash_SumBytes_Error(t *testing.T) {
	// Test with empty bytes - should not error
	repo := NewMD5()
	result := repo.SumBytes([]byte{})
	testutils.Assert(t, len(result) > 0, "SumBytes with empty bytes should return hash")
}

func TestDefaultHash_SumTimes_Multiple(t *testing.T) {
	repo := NewMD5()
	result1 := repo.SumTimes("test", 1)
	result2 := repo.SumTimes("test", 2)
	testutils.Assert(t, result1 != result2, "SumTimes with different times should produce different results")
	testutils.Assert(t, len(result1) > 0, "SumTimes should return non-empty result")
	testutils.Assert(t, len(result2) > 0, "SumTimes should return non-empty result")
}

func TestDefaultHash_SumBytesTimes_Extended(t *testing.T) {
	repo := NewMD5()
	result := repo.SumBytesTimes([]byte("test"), 3)
	testutils.Assert(t, len(result) > 0, "SumBytesTimes should return non-empty result")
}

func TestNewHashRepo_SHA512_224(t *testing.T) {
	repo := NewHashRepo(crypto.SHA512_224)
	testutils.Assert(t, repo != nil, "NewHashRepo should return non-nil for SHA512_224")
	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}

func TestNewHashRepo_SHA512_256(t *testing.T) {
	repo := NewHashRepo(crypto.SHA512_256)
	testutils.Assert(t, repo != nil, "NewHashRepo should return non-nil for SHA512_256")
	result := repo.Sum("test")
	testutils.Assert(t, len(result) > 0, "Sum should return non-empty result")
}
