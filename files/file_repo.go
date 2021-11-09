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
	"os"
)

// FileRepo execute file functions
type FileRepo interface {
	Open(file string) (*FileInfo, error)
	OpenFile(file string, flag int, perm os.FileMode) (*FileInfo, error)
	// FileOpened judge if file is opening
	FileOpened(string) bool
	// FileInfo get information with file name
	FileInfo(name string) (os.FileInfo, error)
	// Close close the file
	Close(file string) error
	// CloseAll close all files which were opened
	CloseAll() error
	// Read file
	Read(string) (b []byte, n int, err error)
	// Write rewrite file with context
	Write(name, context string, opts ...WriteOption) (int, error)
	WriteBytes(name string, b []byte, opts ...WriteOption) (int, error)
	// WriteAppend append context to the file
	WriteAppend(name, context string, opts ...WriteOption) (int, error)
	WriteAppendBytes(name string, b []byte, opts ...WriteOption) (int, error)
	// Rename rename file
	Rename(oldPath, newPath string) error
	// SetReadBufLength set length of buffer to read file, default: 1024
	SetReadBufLength(int64) error
}

type Option func(*Options)
type Options struct {
	ReadBufferLength int64
	ConcurrencyRead  bool
}

func ReadBufferLength(rbuf int64) Option {
	return func(o *Options) {
		o.ReadBufferLength = rbuf
	}
}

func ConcurrencyRead() Option {
	return func(o *Options) {
		o.ConcurrencyRead = true
	}
}

type WriteOption func(*WriteOptions)
type WriteOptions struct {
	Flag *int
}

func WriteFlag(flag int) WriteOption {
	return func(o *WriteOptions) {
		if o.Flag == nil {
			o.Flag = &flag
		} else {
			*o.Flag = *o.Flag | flag
		}
	}
}

// DefaultReadBufferLength default reader buffer length
const (
	DefaultReadBufferLength = 1024
)
