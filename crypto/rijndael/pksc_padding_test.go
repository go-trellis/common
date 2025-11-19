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

package rijndael

import (
	"testing"

	"trellis.tech/trellis/common.v3/utils/testutils"
)

func TestPKCSPadding(t *testing.T) {
	buf := []byte("hello")
	n := 8
	padded, err := PKCSPadding(buf, n)
	testutils.Ok(t, err)
	testutils.Assert(t, len(padded) > len(buf), "padded data should be longer")
	testutils.Assert(t, len(padded)%n == 0, "padded data should be multiple of n")
}

func TestPKCSPadding_Empty(t *testing.T) {
	buf := []byte{}
	n := 8
	padded, err := PKCSPadding(buf, n)
	testutils.Ok(t, err)
	testutils.Assert(t, padded == nil, "empty buffer should return nil")
}

func TestPKCSPadding_AlreadyMultiple(t *testing.T) {
	buf := []byte("12345678") // 8 bytes, multiple of 8
	n := 8
	padded, err := PKCSPadding(buf, n)
	testutils.Ok(t, err)
	testutils.Assert(t, len(padded) > len(buf), "should still add padding")
	testutils.Assert(t, len(padded)%n == 0, "padded data should be multiple of n")
}

func TestPKCSPadding_InvalidN(t *testing.T) {
	buf := []byte("hello")
	n := 0
	_, err := PKCSPadding(buf, n)
	testutils.NotOk(t, err, "should return error for n <= 1")
}

func TestPKCSPadding_InvalidNTooLarge(t *testing.T) {
	buf := []byte("hello")
	n := 256
	_, err := PKCSPadding(buf, n)
	testutils.NotOk(t, err, "should return error for n >= 256")
}

func TestPKCS5Padding(t *testing.T) {
	buf := []byte("hello")
	padded, err := PKCS5Padding(buf)
	testutils.Ok(t, err)
	testutils.Assert(t, len(padded) > len(buf), "padded data should be longer")
	testutils.Assert(t, len(padded)%8 == 0, "padded data should be multiple of 8")
}

func TestPKCSUnPadding(t *testing.T) {
	buf := []byte("hello")
	n := 8
	padded, err := PKCSPadding(buf, n)
	testutils.Ok(t, err)

	unpadded, err := PKCSUnPadding(padded, n)
	testutils.Ok(t, err)
	testutils.Equals(t, buf, unpadded, "unpadded data should match original")
}

func TestPKCSUnPadding_Empty(t *testing.T) {
	buf := []byte{}
	n := 8
	unpadded, err := PKCSUnPadding(buf, n)
	testutils.Ok(t, err)
	testutils.Assert(t, unpadded == nil, "empty buffer should return nil")
}

func TestPKCSUnPadding_NotMultiple(t *testing.T) {
	buf := []byte("hello") // Not a multiple of 8
	n := 8
	unpadded, err := PKCSUnPadding(buf, n)
	testutils.Ok(t, err)
	testutils.Equals(t, buf, unpadded, "should return original if not multiple")
}

func TestPKCSUnPadding_InvalidN(t *testing.T) {
	buf := []byte("12345678")
	n := 0
	_, err := PKCSUnPadding(buf, n)
	testutils.NotOk(t, err, "should return error for n <= 1")
}

func TestPKCSUnPadding_InvalidNTooLarge(t *testing.T) {
	buf := []byte("12345678")
	n := 256
	_, err := PKCSUnPadding(buf, n)
	testutils.NotOk(t, err, "should return error for n >= 256")
}

func TestPKCSUnPadding_PaddingTooLong(t *testing.T) {
	buf := make([]byte, 16)
	buf[15] = 17 // Padding value > block size (8)
	n := 8
	_, err := PKCSUnPadding(buf, n)
	testutils.NotOk(t, err, "should return error for padding too long")
}

func TestPKCSUnPadding_PaddingZero(t *testing.T) {
	buf := make([]byte, 16)
	buf[15] = 0 // Padding value = 0
	n := 8
	_, err := PKCSUnPadding(buf, n)
	testutils.NotOk(t, err, "should return error for padding = 0")
}

func TestPKCSUnPadding_NotAllSame(t *testing.T) {
	buf := make([]byte, 16)
	buf[15] = 8 // Padding value
	buf[14] = 7 // Different value
	buf[13] = 8 // Different value
	n := 8
	_, err := PKCSUnPadding(buf, n)
	testutils.NotOk(t, err, "should return error for inconsistent padding")
}

func TestPKCS5UnPadding(t *testing.T) {
	buf := []byte("hello")
	padded, err := PKCS5Padding(buf)
	testutils.Ok(t, err)

	unpadded, err := PKCS5UnPadding(padded)
	testutils.Ok(t, err)
	testutils.Equals(t, buf, unpadded, "unpadded data should match original")
}
