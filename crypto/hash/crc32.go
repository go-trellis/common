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

import "hash/crc32"

var (
	IEEETable = crc32.IEEETable
)

type CRC32Option func(*CRC32Options)

type CRC32Options struct {
	Table *crc32.Table
	Poly  uint32
}

func CRC32OptionTable(tab *crc32.Table) CRC32Option {
	return func(co *CRC32Options) {
		co.Table = tab
	}
}
func CRC32OptionPoly(poly uint32) CRC32Option {
	return func(co *CRC32Options) {
		co.Poly = poly
	}
}

// NewCRC32 get crc32 hash32repo
func NewCRC32(tab *crc32.Table) Hash32Repo {
	return &defHash32{
		Hash: crc32.New(tab),
	}
}

// NewCRCIEEE #discarded: get ieee hash32repo
func NewCRCIEEE() Hash32Repo {
	return &defHash32{
		Hash: crc32.NewIEEE(),
	}
}

// NewCRC32WithOptions get crc32 repo with options
func NewCRC32WithOptions(opts ...CRC32Option) (Hash32Repo, error) {
	options := CRC32Options{}
	for _, o := range opts {
		o(&options)
	}

	if options.Table == nil {
		options.Table = MakeCRC32Table(options.Poly)
	}

	return &defHash32{
		Hash: crc32.New(options.Table),
	}, nil
}

func MakeCRC32Table(poly uint32) *crc32.Table {
	return crc32.MakeTable(poly)
}
