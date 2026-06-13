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

package logger

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotatingFileWriter_StableActiveFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	writer, err := NewRotateLogsWithConfig(&RotateConfig{
		FileName:     logPath,
		MoveFileType: MoveFileTypeDaily,
	})
	if err != nil {
		t.Fatalf("NewRotateLogsWithConfig: %v", err)
	}

	if _, err := writer.Write([]byte("hello\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	info, err := os.Lstat(logPath)
	if err != nil {
		t.Fatalf("Lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("active log file should not be a symlink")
	}

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestRotatingFileWriter_RotateOnPeriodChange(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	current := time.Date(2025, 6, 13, 10, 0, 0, 0, time.Local)
	w, err := newRotatingFileWriterFromRotateConfig(&RotateConfig{
		FileName:     logPath,
		MoveFileType: MoveFileTypeDaily,
	}, true)
	if err != nil {
		t.Fatalf("newRotatingFileWriterFromRotateConfig: %v", err)
	}
	w.now = func() time.Time { return current }

	if _, err := w.Write([]byte("day one\n")); err != nil {
		t.Fatalf("Write day one: %v", err)
	}

	current = current.Add(24 * time.Hour)
	if _, err := w.Write([]byte("day two\n")); err != nil {
		t.Fatalf("Write day two: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	archivePath := logPath + ".20250613"
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("expected archive %s: %v", archivePath, err)
	}

	activeInfo, err := os.Lstat(logPath)
	if err != nil {
		t.Fatalf("Lstat active: %v", err)
	}
	if activeInfo.Mode()&os.ModeSymlink != 0 {
		t.Fatal("active log file should not be a symlink")
	}
}

func TestRotatingFileWriter_NoSymlinkAfterRotate(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	current := time.Date(2025, 6, 13, 10, 0, 0, 0, time.Local)
	w, err := newRotatingFileWriterFromRotateConfig(&RotateConfig{
		FileName:     logPath,
		MoveFileType: MoveFileTypeHourly,
	}, true)
	if err != nil {
		t.Fatalf("newRotatingFileWriterFromRotateConfig: %v", err)
	}
	w.now = func() time.Time { return current }

	if _, err := w.Write([]byte("hour one\n")); err != nil {
		t.Fatalf("Write hour one: %v", err)
	}

	current = current.Add(time.Hour)
	if _, err := w.Write([]byte("hour two\n")); err != nil {
		t.Fatalf("Write hour two: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("unexpected symlink: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
}

func TestNewRotateLogsWithConfig_EmptyFilename(t *testing.T) {
	_, err := NewRotateLogsWithConfig(&RotateConfig{})
	if err == nil {
		t.Fatal("expected error for empty filename")
	}
}
