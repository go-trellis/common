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

package arcfour

import (
	"testing"

	"github.com/go-trellis/common/utils/testutils"
)

func TestNew(t *testing.T) {
	key := []byte("test-key-1234567890123456")
	arcFour, err := New(key)
	testutils.Ok(t, err)
	testutils.Assert(t, arcFour != nil, "New should return non-nil")
}

func TestNew_EmptyKey(t *testing.T) {
	key := []byte{}
	_, err := New(key)
	testutils.NotOk(t, err, "RC4 should return error for empty key")
}

func TestArcFour_Encryption(t *testing.T) {
	key := []byte("test-key-1234567890123456")
	arcFour, err := New(key)
	testutils.Ok(t, err)

	source := []byte("hello world")
	encrypted := arcFour.Encryption(source)
	testutils.Assert(t, encrypted != nil, "Encryption should return non-nil")
	testutils.Assert(t, len(encrypted) == len(source), "Encrypted length should match source length")
}

func TestArcFour_Encryption_Empty(t *testing.T) {
	key := []byte("test-key-1234567890123456")
	arcFour, err := New(key)
	testutils.Ok(t, err)

	source := []byte{}
	encrypted := arcFour.Encryption(source)
	testutils.Assert(t, encrypted == nil, "Encryption should return nil for empty source")
}

func TestArcFour_Encryption_Decryption(t *testing.T) {
	key := []byte("test-key-1234567890123456")
	arcFour, err := New(key)
	testutils.Ok(t, err)

	source := []byte("hello world")
	encrypted := arcFour.Encryption(source)

	// RC4 is symmetric, so encrypting again should decrypt
	arcFour2, err := New(key)
	testutils.Ok(t, err)
	decrypted := arcFour2.Encryption(encrypted)

	testutils.Equals(t, source, decrypted, "Encryption should be symmetric")
}

func TestArcFour_Encryption_DifferentKeys(t *testing.T) {
	key1 := []byte("test-key-1234567890123456")
	key2 := []byte("different-key-123456789")

	arcFour1, err := New(key1)
	testutils.Ok(t, err)
	arcFour2, err := New(key2)
	testutils.Ok(t, err)

	source := []byte("hello world")
	encrypted1 := arcFour1.Encryption(source)
	encrypted2 := arcFour2.Encryption(source)

	testutils.Assert(t, string(encrypted1) != string(encrypted2), "Different keys should produce different encrypted output")
}

func TestArcFour_Encryption_Concurrent(t *testing.T) {
	key := []byte("test-key-1234567890123456")
	arcFour, err := New(key)
	testutils.Ok(t, err)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			source := []byte("test message")
			encrypted := arcFour.Encryption(source)
			testutils.Assert(t, encrypted != nil, "Encryption should work concurrently")
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestArcFour_Encryption_LongMessage(t *testing.T) {
	key := []byte("test-key-1234567890123456")
	arcFour, err := New(key)
	testutils.Ok(t, err)

	source := make([]byte, 1000)
	for i := range source {
		source[i] = byte(i % 256)
	}

	encrypted := arcFour.Encryption(source)
	testutils.Assert(t, encrypted != nil, "Encryption should handle long messages")
	testutils.Assert(t, len(encrypted) == len(source), "Encrypted length should match source length")
}
