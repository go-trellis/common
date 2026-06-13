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
	mode          RotateMode
	rotationTime  time.Duration
	maxSize       int64
	maxAge        time.Duration
	rotationCount uint
	forceNewFile  bool

	file      *os.File
	curPeriod string
	now       func() time.Time
}

func newRotatingFileWriter(config *RotateLogsConfig) (*rotatingFileWriter, error) {
	if config.LogPath == "" {
		return nil, fmt.Errorf("log_path is required")
	}
	if config.MaxAge > 0 && config.RotationCount > 0 {
		return nil, fmt.Errorf("max_age and rotation_count cannot both be set")
	}

	rotationTime := config.RotationTime
	switch config.RotateMode {
	case RotateModeHour:
		if rotationTime == 0 {
			rotationTime = time.Hour
		}
	case RotateModeDay:
		if rotationTime == 0 {
			rotationTime = 24 * time.Hour
		}
	default:
		if rotationTime == 0 {
			rotationTime = 24 * time.Hour
		}
	}

	return &rotatingFileWriter{
		logPath:       config.LogPath,
		mode:          config.RotateMode,
		rotationTime:  rotationTime,
		maxSize:       config.MaxSize,
		maxAge:        config.MaxAge,
		rotationCount: config.RotationCount,
		forceNewFile:  config.ForceNewFile,
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
			if filePeriod != period || w.forceNewFile {
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

	if period != w.curPeriod {
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

func (w *rotatingFileWriter) rotateLocked(newPeriod string) error {
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		if err != nil {
			return err
		}
	}

	if err := w.archiveActiveFile(w.curPeriod); err != nil {
		return err
	}

	if err := w.openActiveFile(); err != nil {
		return err
	}

	w.curPeriod = newPeriod

	if err := w.cleanup(); err != nil {
		return err
	}

	return nil
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
	base := truncateToRotationTime(t, w.rotationTime)
	switch w.mode {
	case RotateModeHour:
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
	if w.maxAge <= 0 && w.rotationCount == 0 {
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

	toRemove := make(map[string]struct{})

	if w.maxAge > 0 {
		cutoff := w.now().Add(-w.maxAge)
		for _, f := range files {
			if f.modTime.Before(cutoff) {
				toRemove[f.path] = struct{}{}
			}
		}
	}

	if w.rotationCount > 0 && uint(len(files)) > w.rotationCount {
		excess := len(files) - int(w.rotationCount)
		for i := 0; i < excess; i++ {
			toRemove[files[i].path] = struct{}{}
		}
	}

	for path := range toRemove {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

var _ io.Writer = (*rotatingFileWriter)(nil)
var _ io.Closer = (*rotatingFileWriter)(nil)
