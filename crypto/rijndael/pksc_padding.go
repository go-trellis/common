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
	"bytes"

	"trellis.tech/common.v2"
	"trellis.tech/common.v2/errcode"
)

// Errors padding & unpadding can return
var (
	ErrorPaddingBadMultiple   = errcode.New(common.FormatNamespaceString("Bad PKCS#7 padding - bad multiple"))
	ErrorPaddingTooLong       = errcode.New(common.FormatNamespaceString("Bad PKCS#7 padding - too long"))
	ErrorPaddingTooShort      = errcode.New(common.FormatNamespaceString("Bad PKCS#7 padding - too short"))
	ErrorPaddingNotAllTheSame = errcode.New(common.FormatNamespaceString("Bad PKCS#7 padding - not all the same"))
)

// PKCSPadding buf using PKCS#7 to a multiple of n.
func PKCSPadding(buf []byte, n int) ([]byte, error) {
	if n <= 1 || n >= 256 {
		return nil, ErrorPaddingBadMultiple
	}
	_length := len(buf)
	if 0 == _length {
		return nil, nil
	}
	_padding := n - _length%n
	return append(buf, bytes.Repeat([]byte{byte(_padding)}, _padding)...), nil
}

// PKCS5Padding PKCS#5: buf using PKCS#7 to a multiple of 8.
func PKCS5Padding(buf []byte) ([]byte, error) {
	return PKCSPadding(buf, 8)
}

// PKCSUnPadding buf using PKCS#7 to a multiple of n.
func PKCSUnPadding(buf []byte, n int) ([]byte, error) {
	if n <= 1 || n >= 256 {
		return nil, ErrorPaddingBadMultiple
	}
	_length := len(buf)
	if _length == 0 {
		return nil, nil
	}
	if (_length % n) != 0 {
		return buf, nil
	}
	_padding := int(buf[_length-1])
	if _padding > n {
		return nil, ErrorPaddingTooLong
	}
	if _padding == 0 {
		return nil, ErrorPaddingTooShort
	}
	for i := 0; i < _padding; i++ {
		if buf[_length-1-i] != byte(_padding) {
			return nil, ErrorPaddingNotAllTheSame
		}
	}
	return buf[:_length-_padding], nil
}

// PKCS5UnPadding buf using PKCS#7 from a multiple of 8 returning a slice of
func PKCS5UnPadding(buf []byte) ([]byte, error) {
	return PKCSUnPadding(buf, 8)
}
