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
	"crypto/rc4"
	"sync"
)

// ArcFour rc4
type ArcFour interface {
	Encryption(source []byte) []byte
}

type arcFour struct {
	locker sync.Mutex
	cipher *rc4.Cipher
}

// New return default arc four instance
func New(key []byte) (ArcFour, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return &arcFour{cipher: c}, nil
}

func (p *arcFour) Encryption(src []byte) []byte {
	_lenSrc := len(src)
	if _lenSrc == 0 {
		return nil
	}
	dst := make([]byte, _lenSrc)
	p.locker.Lock()
	// p.cipher.Reset()
	p.cipher.XORKeyStream(dst, src)
	p.locker.Unlock()
	return dst
}
