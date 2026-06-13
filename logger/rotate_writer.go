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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type rotatingFileWriter struct {
	mu sync.Mutex

	logPath       string
	moveType      MoveFileType
	rotationTime  time.Duration
	maxSize       int64
	rotationCount uint

	file      *os.File
	curPeriod string
	now       func() time.Time
}

func newRotatingFileWriterFromRotateConfig(cfg *RotateConfig, defaultDaily bool) (*rotatingFileWriter, error) {
	if cfg == nil || cfg.FileName == "" {
		return nil, fmt.Errorf("filename is empty")
	}

	moveType := cfg.MoveFileType
	switch moveType {
	case MoveFileTypePerMinite, MoveFileTypeHourly, MoveFileTypeDaily:
	case MoveFileTypeNone:
		if defaultDaily {
			moveType = MoveFileTypeDaily
		}
	default:
		if defaultDaily {
			moveType = MoveFileTypeDaily
		} else {
			moveType = MoveFileTypeNone
		}
	}

	return &rotatingFileWriter{
		logPath:       cfg.FileName,
		moveType:      moveType,
		rotationTime:  moveType.Duration(),
		maxSize:       cfg.RotationSize,
		rotationCount: cfg.RotationCount,
		now:           time.Now,
	}, nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureOpen(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *rotatingFileWriter) ensureOpen() error {
	period := w.periodKey(w.now())

	if w.file == nil {
		if info, err := os.Stat(w.logPath); err == nil && info.Size() > 0 {
			filePeriod := w.periodKey(info.ModTime())
			if w.shouldRotateByPeriod(period, filePeriod) {
				w.curPeriod = filePeriod
				return w.rotateLocked(period)
			}
		}
		if err := w.openActiveFile(); err != nil {
			return err
		}
		w.curPeriod = period
		return nil
	}

	if w.shouldRotateByPeriod(period, w.curPeriod) {
		return w.rotateLocked(period)
	}

	if w.maxSize > 0 {
		info, err := w.file.Stat()
		if err != nil {
			return err
		}
		if info.Size() >= w.maxSize {
			return w.rotateLocked(period)
		}
	}

	return nil
}

func (w *rotatingFileWriter) shouldRotateByPeriod(currentPeriod, previousPeriod string) bool {
	if w.rotationTime <= 0 {
		return false
	}
	return currentPeriod != previousPeriod
}

func (w *rotatingFileWriter) rotateLocked(newPeriod string) error {
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		if err != nil {
			return err
		}
	}

	archivePeriod := w.curPeriod
	if archivePeriod == "" {
		archivePeriod = w.now().Format("20060102150405")
	}

	if err := w.archiveActiveFile(archivePeriod); err != nil {
		return err
	}

	if err := w.openActiveFile(); err != nil {
		return err
	}

	w.curPeriod = newPeriod

	return w.cleanup()
}

func (w *rotatingFileWriter) openActiveFile() error {
	dir := filepath.Dir(w.logPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create log directory: %w", err)
		}
	}

	f, err := os.OpenFile(w.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	w.file = f
	return nil
}

func (w *rotatingFileWriter) archiveActiveFile(period string) error {
	info, err := os.Stat(w.logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("log path %q is a directory", w.logPath)
	}
	if info.Size() == 0 {
		return os.Remove(w.logPath)
	}

	archiveName, err := w.nextArchiveName(period)
	if err != nil {
		return err
	}
	return os.Rename(w.logPath, archiveName)
}

func (w *rotatingFileWriter) nextArchiveName(period string) (string, error) {
	base := w.logPath + "." + period
	name := base
	for i := 0; ; i++ {
		if i > 0 {
			name = fmt.Sprintf("%s.%d", base, i)
		}
		if _, err := os.Stat(name); os.IsNotExist(err) {
			return name, nil
		} else if err != nil {
			return "", err
		}
	}
}

func (w *rotatingFileWriter) periodKey(t time.Time) string {
	if w.rotationTime <= 0 {
		return ""
	}

	base := truncateToRotationTime(t, w.rotationTime)
	switch w.moveType {
	case MoveFileTypePerMinite:
		return base.Format("200601021504")
	case MoveFileTypeHourly:
		return base.Format("2006010215")
	default:
		return base.Format("20060102")
	}
}

func truncateToRotationTime(t time.Time, rotationTime time.Duration) time.Time {
	if rotationTime <= 0 {
		return t
	}
	if t.Location() != time.UTC {
		base := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
		base = base.Truncate(rotationTime)
		return time.Date(base.Year(), base.Month(), base.Day(), base.Hour(), base.Minute(), base.Second(), base.Nanosecond(), t.Location())
	}
	return t.Truncate(rotationTime)
}

func (w *rotatingFileWriter) cleanup() error {
	if w.rotationCount == 0 {
		return nil
	}

	matches, err := filepath.Glob(w.logPath + ".*")
	if err != nil {
		return err
	}

	type archivedFile struct {
		path    string
		modTime time.Time
	}

	files := make([]archivedFile, 0, len(matches))
	for _, path := range matches {
		if strings.HasSuffix(path, "_lock") || strings.HasSuffix(path, "_symlink") {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		files = append(files, archivedFile{path: path, modTime: info.ModTime()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	if uint(len(files)) <= w.rotationCount {
		return nil
	}

	excess := len(files) - int(w.rotationCount)
	for i := 0; i < excess; i++ {
		if err := os.Remove(files[i].path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

var _ io.Writer = (*rotatingFileWriter)(nil)
var _ io.Closer = (*rotatingFileWriter)(nil)
