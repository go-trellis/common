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
	"compress/gzip"
	"embed"
	"io"
	"io/fs"
	"path/filepath"
)

type GzipFS struct {
	options *Options
	embed   embed.FS
}

func New(fs embed.FS, opts ...Option) GzipFS {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	options.init()
	return GzipFS{options, fs}
}

func (p GzipFS) Open(name string) (fs.File, error) {
	path := filepath.Join(p.options.Path, name)
	var f fs.File
	if f, err := p.embed.Open(path); err == nil {
		return f, nil
	}

	f, err := p.embed.Open(path + p.options.Suffix)
	if err != nil {
		return f, err
	}
	gr, err := gzip.NewReader(f)
	if err != nil {
		return f, err
	}
	defer gr.Close()

	c, err := io.ReadAll(gr)
	if err != nil {
		return f, err
	}
	return &File{file: f, content: c, suffix: p.options.Suffix}, nil
}
