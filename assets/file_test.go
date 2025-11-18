/*
Copyright 2021 The Prometheus Authors
Licensed under the Apache License, Version 2.0 (the "License");

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

package assets

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"trellis.tech/trellis/common.v3/testutils"
)

func TestFile_Read(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	testutils.Ok(t, err)

	f, err := os.Open(testFile)
	testutils.Ok(t, err)
	defer f.Close()

	file := &File{
		file:    f,
		content: []byte("test content"),
		offset:  0,
		suffix:  "",
	}

	buf := make([]byte, 5)
	n, err := file.Read(buf)
	testutils.Ok(t, err)
	testutils.Equals(t, 5, n, "should read 5 bytes")
	testutils.Equals(t, "test ", string(buf), "should read correct content")
}

func TestFile_Read_EOF(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	testutils.Ok(t, err)

	f, err := os.Open(testFile)
	testutils.Ok(t, err)
	defer f.Close()

	file := &File{
		file:    f,
		content: []byte("test"),
		offset:  4,
		suffix:  "",
	}

	buf := make([]byte, 10)
	n, err := file.Read(buf)
	testutils.Equals(t, io.EOF, err, "should return EOF")
	testutils.Equals(t, 0, n, "should read 0 bytes")
}

func TestFile_Read_Partial(t *testing.T) {
	content := []byte("hello world")
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, content, 0644)
	testutils.Ok(t, err)

	f, err := os.Open(testFile)
	testutils.Ok(t, err)
	defer f.Close()

	file := &File{
		file:    f,
		content: content,
		offset:  6,
		suffix:  "",
	}

	buf := make([]byte, 20)
	n, err := file.Read(buf)
	testutils.Equals(t, io.EOF, err, "should return EOF when reading beyond content")
	testutils.Equals(t, 5, n, "should read remaining 5 bytes")
	testutils.Equals(t, "world", string(buf[:n]), "should read correct content")
}

func TestFile_Stat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt.gz")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	testutils.Ok(t, err)

	f, err := os.Open(testFile)
	testutils.Ok(t, err)
	defer f.Close()

	content := []byte("decompressed content")
	file := &File{
		file:    f,
		content: content,
		offset:  0,
		suffix:  ".gz",
	}

	info, err := file.Stat()
	testutils.Ok(t, err)
	testutils.Assert(t, info != nil, "Stat should return FileInfo")
	testutils.Assert(t, !strings.Contains(info.Name(), ".gz"), "Name should not contain suffix")
	testutils.Equals(t, int64(len(content)), info.Size(), "Size should return actual content size")
}

func TestFile_Close(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	testutils.Ok(t, err)

	f, err := os.Open(testFile)
	testutils.Ok(t, err)

	file := &File{
		file:    f,
		content: []byte("test"),
		offset:  0,
		suffix:  "",
	}

	err = file.Close()
	testutils.Ok(t, err)
}

func TestFileInfo_Name(t *testing.T) {
	mockInfo := &mockFileInfo{name: "test.txt.gz"}
	fi := FileInfo{
		fi:         mockInfo,
		actualSize: 100,
		suffix:     ".gz",
	}

	name := fi.Name()
	testutils.Equals(t, "test.txt", name, "Name should remove suffix")
}

func TestFileInfo_Name_NoSuffix(t *testing.T) {
	mockInfo := &mockFileInfo{name: "test.txt"}
	fi := FileInfo{
		fi:         mockInfo,
		actualSize: 100,
		suffix:     "",
	}

	name := fi.Name()
	testutils.Equals(t, "test.txt", name, "Name should return original if no suffix")
}

func TestFileInfo_Size(t *testing.T) {
	fi := FileInfo{
		fi:         &mockFileInfo{size: 50},
		actualSize: 100,
		suffix:     "",
	}

	testutils.Equals(t, int64(100), fi.Size(), "Size should return actualSize")
}

func TestFileInfo_Mode(t *testing.T) {
	fi := FileInfo{
		fi:         &mockFileInfo{mode: 0644},
		actualSize: 100,
		suffix:     "",
	}

	testutils.Equals(t, os.FileMode(0644), fi.Mode(), "Mode should return underlying mode")
}

func TestFileInfo_ModTime(t *testing.T) {
	now := time.Now()
	fi := FileInfo{
		fi:         &mockFileInfo{modTime: now},
		actualSize: 100,
		suffix:     "",
	}

	testutils.Equals(t, now, fi.ModTime(), "ModTime should return underlying time")
}

func TestFileInfo_IsDir(t *testing.T) {
	fi := FileInfo{
		fi:         &mockFileInfo{isDir: false},
		actualSize: 100,
		suffix:     "",
	}

	testutils.Assert(t, !fi.IsDir(), "IsDir should return underlying value")
}

func TestFileInfo_Sys(t *testing.T) {
	fi := FileInfo{
		fi:         &mockFileInfo{},
		actualSize: 100,
		suffix:     "",
	}

	testutils.Assert(t, fi.Sys() == nil, "Sys should return nil")
}

type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

