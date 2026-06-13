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

	"github.com/go-trellis/common/utils/testutils"
)

func TestRotatingFileWriter_StableActiveFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	writer, err := NewRotateLogsWriter(&RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
	})
	testutils.Ok(t, err)

	_, err = writer.Write([]byte("hello\n"))
	testutils.Ok(t, err)

	info, err := os.Lstat(logPath)
	testutils.Ok(t, err)
	testutils.Assert(t, info.Mode()&os.ModeSymlink == 0, "active log file should not be a symlink")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestRotatingFileWriter_RotateOnPeriodChange(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	current := time.Date(2025, 6, 13, 10, 0, 0, 0, time.Local)
	w, err := newRotatingFileWriter(&RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
	})
	testutils.Ok(t, err)
	w.now = func() time.Time { return current }

	_, err = w.Write([]byte("day one\n"))
	testutils.Ok(t, err)

	current = current.Add(24 * time.Hour)
	_, err = w.Write([]byte("day two\n"))
	testutils.Ok(t, err)
	testutils.Ok(t, w.Close())

	archivePath := logPath + ".20250613"
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("expected archive %s: %v", archivePath, err)
	}

	activeInfo, err := os.Lstat(logPath)
	testutils.Ok(t, err)
	testutils.Assert(t, activeInfo.Mode()&os.ModeSymlink == 0, "active log file should not be a symlink")
}

func TestRotatingFileWriter_NoSymlinkAfterRotate(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	current := time.Date(2025, 6, 13, 10, 0, 0, 0, time.Local)
	w, err := newRotatingFileWriter(&RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeHour,
		RotationTime: time.Hour,
	})
	testutils.Ok(t, err)
	w.now = func() time.Time { return current }

	_, err = w.Write([]byte("hour one\n"))
	testutils.Ok(t, err)

	current = current.Add(time.Hour)
	_, err = w.Write([]byte("hour two\n"))
	testutils.Ok(t, err)
	testutils.Ok(t, w.Close())

	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			t.Fatalf("unexpected symlink: %s", path)
		}
		return nil
	})
	testutils.Ok(t, err)
}

func TestNewRotateLogsWriter_MaxAgeAndRotationCountConflict(t *testing.T) {
	_, err := NewRotateLogsWriter(&RotateLogsConfig{
		LogPath:       "/tmp/test.log",
		MaxAge:        time.Hour,
		RotationCount: 1,
	})
	testutils.NotOk(t, err, "should reject max_age and rotation_count together")
}
