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
	"encoding/hex"
	"hash"
)

// Repo hash functions manager
type Repo interface {
	Sum(s string) string
	SumBytes(b []byte) string
	SumTimes(s string, times uint) string
	SumBytesTimes(b []byte, times uint) string
	Reset()
}

// Hash32Repo hash32 functions manager
type Hash32Repo interface {
	Sum(s string) string
	SumBytes(b []byte) string
	SumTimes(s string, times uint) string
	SumBytesTimes(b []byte, times uint) string
	Sum32(b []byte) (uint32, error)
	Reset()
}

// NewHashRepo get hash repo by crypto type
func NewHashRepo(h crypto.Hash) Repo {
	switch h {
	case crypto.MD5:
		return NewMD5()
	case crypto.SHA1:
		return NewSHA1()
	case crypto.SHA224:
		return NewSHA224()
	case crypto.SHA256:
		return NewSHA256()
	case crypto.SHA384:
		return NewSHA384()
	case crypto.SHA512:
		return NewSHA512()
	case crypto.SHA512_224:
		return NewSHA512_224()
	case crypto.SHA512_256:
		return NewSHA512_256()
	}

	return nil
}

type defaultHash struct {
	Hash hash.Hash
}

func (p *defaultHash) Sum(s string) string {
	return p.SumBytes([]byte(s))
}

func (p *defaultHash) SumBytes(data []byte) string {
	p.Hash.Reset()
	_, err := p.Hash.Write(data)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(p.Hash.Sum(nil))
}

func (p *defaultHash) SumTimes(s string, times uint) string {
	if times == 0 {
		return ""
	}
	for i := 0; i < int(times); i++ {
		s = p.Sum(s)
	}
	return s
}

func (p *defaultHash) SumBytesTimes(b []byte, times uint) string {
	return p.SumTimes(string(b), times)
}

func (p *defaultHash) Reset() {
	p.Hash.Reset()
}

// default hash 32
type defHash32 struct {
	Hash hash.Hash32
}

func (p *defHash32) Sum(s string) string {
	return p.SumBytes([]byte(s))
}

func (p *defHash32) SumBytes(data []byte) string {
	p.Hash.Reset()
	_, err := p.Hash.Write(data)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(p.Hash.Sum(nil))
}

func (p *defHash32) SumTimes(s string, times uint) string {
	if times == 0 {
		return ""
	}

	for i := 0; i < int(times); i++ {
		s = p.Sum(s)
	}
	return s
}

func (p *defHash32) SumBytesTimes(bs []byte, times uint) string {
	return p.SumTimes(string(bs), times)
}

func (p *defHash32) Sum32(b []byte) (uint32, error) {
	_, err := p.Hash.Write(b)
	if err != nil {
		return 0, err
	}
	return p.Hash.Sum32(), nil
}

func (p *defHash32) Reset() {
	p.Hash.Reset()
}
