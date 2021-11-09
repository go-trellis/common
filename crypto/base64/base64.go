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
	"fmt"
	"sync"
)

// default base64 encoders
const (
	EncodeStd    = "trellis::algo::encodeStd"
	EncodeRawStd = "trellis::algo::encodeRawStd"
	EncodeURL    = "trellis::algo::encodeURL"
	EncodeRawURL = "trellis::algo::encodeRawURL"
)

type base struct {
	mapEncoders map[string]*base64.Encoding
	locker      sync.RWMutex
}

var defaultBase *base

func init() {
	defaultBase = &base{
		mapEncoders: map[string]*base64.Encoding{
			EncodeStd:    base64.StdEncoding,
			EncodeRawStd: base64.RawStdEncoding,
			EncodeURL:    base64.URLEncoding,
			EncodeRawURL: base64.RawURLEncoding,
		},
	}
}

type Option func(*Options)

type Options struct {
	Padding *rune
}

func (p *Options) getKey(encoder string) string {
	if p == nil || p.Padding == nil {
		return encoder
	}
	return getEncoderKey(encoder, *p.Padding)
}

func getEncoderKey(encoder string, padding rune) string {
	return fmt.Sprintf("%s::%d", encoder, padding)
}

func Padding(padding rune) Option {
	return func(o *Options) {
		o.Padding = &padding
	}
}

func (p *base) getEncoding(key string) (*base64.Encoding, bool) {
	p.locker.RLock()
	encoding, ok := p.mapEncoders[key]
	p.locker.RUnlock()
	return encoding, ok
}

func (p *base) setEncoding(key string, encoding *base64.Encoding) {
	p.locker.Lock()
	p.mapEncoders[key] = encoding
	p.locker.Unlock()
}

// NewEncoding get base64 encoding with input encoder
func NewEncoding(encoder string, opts ...Option) *base64.Encoding {
	options := Options{}
	for _, o := range opts {
		o(&options)
	}
	encoderKey := options.getKey(encoder)
	encoding, ok := defaultBase.getEncoding(encoderKey)
	if ok {
		return encoding
	}
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()
	encoding = base64.NewEncoding(encoder)
	if options.Padding != nil {
		encoding = encoding.WithPadding(*options.Padding)
	}
	defaultBase.setEncoding(encoderKey, encoding)
	return encoding
}

// NewEncodingWithPadding get encoding with encoder and padding
func NewEncodingWithPadding(encoder string, padding rune) *base64.Encoding {
	key := getEncoderKey(encoder, padding)
	encoding, ok := defaultBase.getEncoding(key)
	if ok {
		return encoding
	}
	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	encoding = base64.NewEncoding(encoder).WithPadding(padding)
	defaultBase.setEncoding(key, encoding)
	return encoding
}

// Encode encode bytes with encoder
func Encode(encoder string, src []byte, opts ...Option) string {
	encoding := NewEncoding(encoder, opts...)
	if encoding == nil {
		return ""
	}
	return encoding.EncodeToString(src)
}

// EncodeString encode string with encoder
func EncodeString(encoder string, src string, opts ...Option) string {
	return Encode(encoder, []byte(src), opts...)
}

// Decode decode bytes with encoder
func Decode(encoder string, src []byte, opts ...Option) ([]byte, error) {
	return DecodeString(encoder, string(src), opts...)
}

// DecodeString decode string with encoder
func DecodeString(encoder string, s string, opts ...Option) ([]byte, error) {
	encoding := NewEncoding(encoder, opts...)
	if encoding == nil {
		return nil, nil
	}
	return encoding.DecodeString(s)
}
