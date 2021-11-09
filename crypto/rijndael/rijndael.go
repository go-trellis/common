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
	"crypto/aes"
	"crypto/cipher"
)

// AESECBPKCSEncrypt AES/ECB/PKCSPadding Encrypt
func AESECBPKCSEncrypt(key, source []byte) ([]byte, error) {
	_ecb, err := NewECBEncrypter(key)
	if err != nil {
		return nil, err
	}
	_origData, err := PKCSPadding(source, _ecb.BlockSize())
	if err != nil {
		return nil, err
	}
	crypted := make([]byte, len(_origData))
	_ecb.CryptBlocks(crypted, _origData)
	return crypted, nil
}

// AESECBPKCSDecrypt AES/ECB/PKCSPadding Decrypt
func AESECBPKCSDecrypt(key, crypted []byte) ([]byte, error) {
	_ecb, err := NewECBDecrypter(key)
	if err != nil {
		return nil, err
	}
	origData := make([]byte, len(crypted))
	_ecb.CryptBlocks(origData, crypted)
	origData, err = PKCSUnPadding(origData, _ecb.BlockSize())
	if err != nil {
		return nil, err
	}
	return origData, nil
}

type ecb struct {
	b         cipher.Block
	blockSize int
}

type ecbDecrypter ecb
type ecbEncrypter ecb

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

// NewECBEncrypter returns a BlockMode which encrypts in electronic code book
// key, using the given AES key.
func NewECBEncrypter(key []byte) (cipher.BlockMode, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return (*ecbEncrypter)(newECB(b)), nil
}

// NewECBDecrypter returns a BlockMode which decrypts in electronic code book
// mode, using the given Block.
func NewECBDecrypter(key []byte) (cipher.BlockMode, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return (*ecbDecrypter)(newECB(b)), nil
}

func (p *ecbEncrypter) BlockSize() int { return p.blockSize }
func (p *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%p.blockSize != 0 {
		return
	}
	if len(dst) < len(src) {
		return
	}
	for len(src) > 0 {
		p.b.Encrypt(dst, src[:p.blockSize])
		src = src[p.blockSize:]
		dst = dst[p.blockSize:]
	}
}

func (p *ecbDecrypter) BlockSize() int { return p.blockSize }
func (p *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%p.blockSize != 0 {
		return
	}
	if len(dst) < len(src) {
		return
	}
	for len(src) > 0 {
		p.b.Decrypt(dst, src[:p.blockSize])
		src = src[p.blockSize:]
		dst = dst[p.blockSize:]
	}
}
