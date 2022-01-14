/*
Copyright © 2020 Henry Huang <hhh@rutcode.com>

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

package files

import (
	"io"
	"os"
)

// FileMode
const (
	FileModeReadOnly  os.FileMode = 0444
	FileModeReadWrite os.FileMode = 0666

	FileFlagReadOnly  = os.O_RDONLY
	FileFlagReadWrite = os.O_RDWR | os.O_APPEND | os.O_CREATE
)

func Read(name string, opts ...Option) (b []byte, n int64, err error) {
	fi, err := OpenReadFile(name)
	if err != nil {
		return nil, 0, err
	}

	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if options.ReadBufferLength <= 0 {
		options.ReadBufferLength = DefaultReadBufferLength
	}

	var offset int64 = 0

	for {
		buf := make([]byte, options.ReadBufferLength)
		m, e := fi.ReadAt(buf, offset)
		if e != nil && e != io.EOF {
			err = e
			return
		}
		offset += int64(m)
		b = append(b, buf[:m]...)
		if m < int(options.ReadBufferLength) {
			break
		}
	}
	return
}

func OpenReadFile(name string) (*os.File, error) {
	return OpenFile(name, FileFlagReadOnly, FileModeReadOnly)
}

func OpenWriteFile(name string) (*os.File, error) {
	return OpenFile(name, FileFlagReadWrite, FileModeReadWrite)
}

func OpenFile(name string, flag int, fileMode os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, fileMode)
}
