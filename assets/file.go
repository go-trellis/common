// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package assets

import (
	"io"
	"io/fs"
	"time"
)

type File struct {
	file    fs.File
	content []byte
	offset  int
	suffix  string
}

// Stat implements the fs.File interface.
func (p *File) Stat() (fs.FileInfo, error) {
	stat, err := p.file.Stat()
	if err != nil {
		return stat, err
	}
	return FileInfo{stat, int64(len(p.content)), p.suffix}, nil
}

// Read implements the fs.File interface.
func (p *File) Read(buf []byte) (int, error) {
	if len(buf) > len(p.content)-p.offset {
		buf = buf[0:len(p.content[p.offset:])]
	}
	n := copy(buf, p.content[p.offset:])
	if n == len(p.content)-p.offset {
		return n, io.EOF
	}
	p.offset += n
	return n, nil
}

// Close implements the fs.File interface.
func (p *File) Close() error {
	return p.file.Close()
}

type FileInfo struct {
	fi         fs.FileInfo
	actualSize int64
	suffix     string
}

// Name implements the fs.FileInfo interface.
func (p FileInfo) Name() string {
	name := p.fi.Name()
	return name[:len(name)-len(p.suffix)]
}

// Size implements the fs.FileInfo interface.
func (p FileInfo) Size() int64 { return p.actualSize }

// Mode implements the fs.FileInfo interface.
func (p FileInfo) Mode() fs.FileMode { return p.fi.Mode() }

// ModTime implements the fs.FileInfo interface.
func (p FileInfo) ModTime() time.Time { return p.fi.ModTime() }

// IsDir implements the fs.FileInfo interface.
func (p FileInfo) IsDir() bool { return p.fi.IsDir() }

// Sys implements the fs.FileInfo interface.
func (p FileInfo) Sys() interface{} { return nil }
