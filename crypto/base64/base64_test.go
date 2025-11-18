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

package base64

import (
	"encoding/base64"
	"strings"
	"testing"

	"trellis.tech/trellis/common.v3/testutils"
)

func TestNewEncoding(t *testing.T) {
	encoding := NewEncoding(EncodeStd)
	testutils.Assert(t, encoding != nil, "encoding should not be nil")
	testutils.Equals(t, base64.StdEncoding, encoding, "should return standard encoding")
}

func TestNewEncoding_Cached(t *testing.T) {
	encoding1 := NewEncoding(EncodeStd)
	encoding2 := NewEncoding(EncodeStd)
	testutils.Assert(t, encoding1 == encoding2, "should return same cached encoding")
}

func TestNewEncoding_WithPadding(t *testing.T) {
	encoding := NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/", Padding('='))
	testutils.Assert(t, encoding != nil, "encoding should not be nil")
}

func TestNewEncodingWithPadding(t *testing.T) {
	encoding := NewEncodingWithPadding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/", '=')
	testutils.Assert(t, encoding != nil, "encoding should not be nil")
}

func TestEncode(t *testing.T) {
	src := []byte("test data")
	result := Encode(EncodeStd, src)
	testutils.Assert(t, len(result) > 0, "encoded result should not be empty")

	// Verify it can be decoded
	expected := base64.StdEncoding.EncodeToString(src)
	testutils.Equals(t, expected, result, "encoded result should match standard encoding")
}

func TestEncodeString(t *testing.T) {
	src := "test data"
	result := EncodeString(EncodeStd, src)
	testutils.Assert(t, len(result) > 0, "encoded result should not be empty")

	expected := base64.StdEncoding.EncodeToString([]byte(src))
	testutils.Equals(t, expected, result, "encoded result should match standard encoding")
}

func TestDecode(t *testing.T) {
	src := []byte("dGVzdCBkYXRh") // base64 for "test data"
	result, err := Decode(EncodeStd, src)
	testutils.Ok(t, err)
	testutils.Assert(t, len(result) > 0, "decoded result should not be empty")
	testutils.Equals(t, []byte("test data"), result, "decoded result should match")
}

func TestDecodeString(t *testing.T) {
	src := "dGVzdCBkYXRh" // base64 for "test data"
	result, err := DecodeString(EncodeStd, src)
	testutils.Ok(t, err)
	testutils.Assert(t, len(result) > 0, "decoded result should not be empty")
	testutils.Equals(t, []byte("test data"), result, "decoded result should match")
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original := "test data for round trip"
	encoded := EncodeString(EncodeStd, original)
	decoded, err := DecodeString(EncodeStd, encoded)
	testutils.Ok(t, err)
	testutils.Equals(t, original, string(decoded), "round trip should preserve original data")
}

func TestEncodeDecode_AllEncoders(t *testing.T) {
	testCases := []struct {
		name     string
		encoder  string
		padding  rune
		testData string
	}{
		{name: "Std", encoder: EncodeStd, testData: "test data"},
		{name: "RawStd", encoder: EncodeRawStd, testData: "test data"},
		{name: "URL", encoder: EncodeURL, testData: "test data"},
		{name: "RawURL", encoder: EncodeRawURL, testData: "test data"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var encoded string
			if tc.padding != 0 {
				encoded = Encode(tc.encoder, []byte(tc.testData), Padding(tc.padding))
			} else {
				encoded = Encode(tc.encoder, []byte(tc.testData))
			}

			testutils.Assert(t, len(encoded) > 0, "encoded result should not be empty")

			var decoded []byte
			var err error
			if tc.padding != 0 {
				decoded, err = DecodeString(tc.encoder, encoded, Padding(tc.padding))
			} else {
				decoded, err = DecodeString(tc.encoder, encoded)
			}

			testutils.Ok(t, err)
			testutils.Equals(t, tc.testData, string(decoded), "decoded should match original")
		})
	}
}

func TestDecodeString_InvalidBase64(t *testing.T) {
	invalid := "invalid base64!!!"
	_, err := DecodeString(EncodeStd, invalid)
	testutils.NotOk(t, err, "should return error for invalid base64")
}

func TestNewEncoding_InvalidAlphabet(t *testing.T) {
	// This should not panic, but may return nil if encoding creation fails
	encoding := NewEncoding("invalid alphabet")
	// Encoding may be nil if alphabet is invalid and recover catches the panic
	_ = encoding
}

func TestEncode_EmptyData(t *testing.T) {
	result := Encode(EncodeStd, []byte{})
	testutils.Assert(t, len(result) == 0, "encoding empty data should return empty string")
}

func TestDecodeString_EmptyData(t *testing.T) {
	result, err := DecodeString(EncodeStd, "")
	testutils.Ok(t, err)
	testutils.Assert(t, len(result) == 0, "decoding empty string should return empty bytes")
}

func TestEncodeDecodeURLEncoding(t *testing.T) {
	original := "test+data/url"
	encoded := EncodeString(EncodeURL, original)
	testutils.Assert(t, !strings.Contains(encoded, "+"), "URL encoding should not contain +")
	testutils.Assert(t, !strings.Contains(encoded, "/"), "URL encoding should not contain /")

	decoded, err := DecodeString(EncodeURL, encoded)
	testutils.Ok(t, err)
	testutils.Equals(t, original, string(decoded), "round trip should work for URL encoding")
}

