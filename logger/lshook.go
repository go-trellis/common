/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

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

	"github.com/go-trellis/common/errcode"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type RotateConfig struct {
	FileName      string       `yaml:"filename" json:"filename"`
	MoveFileType  MoveFileType `yaml:"move_file_type" json:"move_file_type"`
	RotationSize  int64        `yaml:"rotation_size" json:"rotation_size"`
	RotationCount uint         `yaml:"rotation_count" json:"rotation_count"`
	// Async moves file I/O off the logging goroutine: lines are queued in a
	// bounded channel and written in batches by a background goroutine.
	// When the queue is full Debug/Info/Trace lines are dropped, Warn/Error
	// wait briefly, and Fatal/Panic are written synchronously (only when used
	// through NewLugrusHook, which knows the level). Call Flush or Shutdown
	// before exiting so queued lines are not lost.
	Async bool `yaml:"async" json:"async"`
	// QueueSize is the async queue capacity in lines (default 8192).
	QueueSize int `yaml:"async_queue_size" json:"async_queue_size"`
	// Caller is not used by the rotating file writer (it is only an io.Writer).
	// For go-kit (logger/prometheus.New) set Config.Caller so the logger is
	// wrapped with log.With(..., "caller", log.DefaultCaller).
	// For logrus InitLogger, ReportCaller is set from the top-level "caller" key.
	Caller bool `yaml:"caller" json:"caller"`
}

type LugrusHookConfig struct {
	*RotateConfig
	Formatter logrus.Formatter
}

type LugrusRotateConfig struct {
	Levels []Level
	*RotateConfig
}

type LugrusRotateConfigs []*LugrusRotateConfig

type LugrusLevelConfig struct {
	Levels []Level
	Writer io.Writer
}

type LugrusLevelConfigs []*LugrusLevelConfig

func NewRotateLogsWithConfig(cfg *RotateConfig) (io.Writer, error) {
	return newRotateWriter(cfg, true)
}

func NewRotateLogs(cfg *RotateConfig) (io.Writer, error) {
	return newRotateWriter(cfg, false)
}

// newRotateWriter returns the sync rotating writer, wrapped in an asyncWriter
// when cfg.Async is set. The result is always an io.Closer.
func newRotateWriter(cfg *RotateConfig, defaultDaily bool) (io.Writer, error) {
	w, err := newRotatingFileWriterFromRotateConfig(cfg, defaultDaily)
	if err != nil {
		return nil, err
	}
	if cfg.Async {
		return newAsyncWriter(w, cfg.QueueSize), nil
	}
	return w, nil
}

// levelWriterFor returns the writer to register for lvl: for async writers a
// per-level view with that level's drop policy, otherwise w itself.
func levelWriterFor(w io.Writer, lvl Level) io.Writer {
	if a, ok := w.(*asyncWriter); ok {
		return &levelWriter{a: a, mode: writeModeForLevel(lvl)}
	}
	return w
}

func NewLFShook(output any, formatter logrus.Formatter) logrus.Hook {
	return lfshook.NewHook(output, formatter)
}

func NewLugrusHookWithConfig(cfg *LugrusHookConfig) (logrus.Hook, error) {
	rotate, err := NewRotateLogsWithConfig(cfg.RotateConfig)
	if err != nil {
		return nil, err
	}
	// Without a level there is no per-level drop policy; the async writer's
	// default (block with timeout) applies.
	return NewLFShook(rotate, cfg.Formatter), nil
}

func NewLugrusHook(formatter logrus.Formatter, cfgs []any) (logrus.Hook, error) {
	if len(cfgs) == 0 {
		return nil, nil
	}
	mapLevelWriter := make(lfshook.WriterMap)

	for _, v := range cfgs {
		switch t := v.(type) {
		case *LugrusRotateConfig:
			rotate, err := NewRotateLogsWithConfig(t.RotateConfig)
			if err != nil {
				return nil, errcode.Newf("NewRotateLogs(%+v) failed: %+v", t.RotateConfig, err)
			}
			for _, lvl := range t.Levels {
				mapLevelWriter[logrus.Level(lvl)] = levelWriterFor(rotate, lvl)
			}
		case LugrusRotateConfigs:
			for _, cfg := range t {
				rotate, err := NewRotateLogsWithConfig(cfg.RotateConfig)
				if err != nil {
					return nil, errcode.Newf("NewRotateLogs(%+v) failed: %+v", cfg.RotateConfig, err)
				}
				for _, lvl := range cfg.Levels {
					mapLevelWriter[logrus.Level(lvl)] = levelWriterFor(rotate, lvl)
				}
			}
		case *LugrusLevelConfig:
			for _, lvl := range t.Levels {
				mapLevelWriter[logrus.Level(lvl)] = t.Writer
			}
		case LugrusLevelConfigs:
			for _, cfg := range t {
				for _, lvl := range cfg.Levels {
					mapLevelWriter[logrus.Level(lvl)] = cfg.Writer
				}
			}
		}
	}

	return NewLFShook(mapLevelWriter, formatter), nil
}
