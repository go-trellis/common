/*
Copyright © 2026 Henry Huang <hhh@rutcode.com>

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
	"strings"

	"github.com/go-trellis/common/config"

	"github.com/sirupsen/logrus"
)

var allLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
	TraceLevel,
}

// InitLogger builds a logger from config.
// Supported types: noop (default), console, file.
func InitLogger(cfg config.Config) (Logger, error) {
	if cfg == nil {
		return Noop(), nil
	}

	printers := cfg.GetStringList("std_printers")
	typ := strings.TrimSpace(cfg.GetString("type"))
	switch typ {
	case "noop", "":
		return Noop(), nil
	case "console":
		if len(printers) == 0 {
			printers = []string{"stdout"}
		}
		return NewLogger(baseLogrusConfig(cfg, stdWriters(printers), nil))
	case "file":
		filename := strings.TrimSpace(cfg.GetString("filename"))
		if filename == "" {
			return nil, fmt.Errorf("logger.filename is required when type is file")
		}

		defaultWriter := io.Writer(os.Stderr)
		if len(printers) > 0 {
			defaultWriter = stdWriters(printers)
		}

		logW, err := NewLogger(baseLogrusConfig(cfg, defaultWriter, []any{
			&LugrusRotateConfig{
				Levels: allLevels,
				RotateConfig: &RotateConfig{
					FileName:      filename,
					MoveFileType:  MoveFileType(cfg.GetInt("move_file_type", int(MoveFileTypeDaily))),
					RotationSize:  int64(cfg.GetInt("max_length", 100000000)),
					RotationCount: uint(cfg.GetInt("max_backups", 30)),
					Caller:        cfg.GetBoolean("caller", true),
				},
			},
		}))
		if err != nil {
			return nil, err
		}
		if err := ensureLogFileWritable(filename); err != nil {
			return nil, err
		}
		logW.WithField("filename", filename).Debug("file logger initialized")
		return logW, nil
	default:
		return nil, fmt.Errorf("unknown logger type: %q", typ)
	}
}

// ensureLogFileWritable creates parent dirs and verifies the active log path
// can be opened. Rotate writer opens lazily on first Write, so InitLogger must
// not rely on a log line (which may be filtered by level) to create the file.
func ensureLogFileWritable(filename string) error {
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create log directory for %s: %w", filename, err)
		}
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("logger file %s is not writable: %w", filename, err)
	}
	return f.Close()
}

func baseLogrusConfig(cfg config.Config, defaultWriter io.Writer, configs []any) *LogrusConfig {
	return &LogrusConfig{
		Level:         configLevel(cfg),
		ReportCaller:  cfg.GetBoolean("caller", true),
		Formatter:     logFormatter(cfg),
		DefaultWriter: defaultWriter,
		Configs:       configs,
	}
}

// configLevel returns configured log level, defaulting to InfoLevel when unset
// or out of range. AdapterConfig.GetInt does not apply defaults for missing keys.
func configLevel(cfg config.Config) Level {
	if cfg.GetInterface("level") == nil {
		return InfoLevel
	}
	lv := Level(cfg.GetInt("level"))
	if lv < PanicLevel || lv > TraceLevel {
		return InfoLevel
	}
	return lv
}

func logFormatter(cfg config.Config) logrus.Formatter {
	switch cfg.GetString("encoding", "text") {
	case "json":
		return &JSONFormatter{}
	default:
		return &TextFormatter{}
	}
}

func stdWriters(printers []string) io.Writer {
	var ws []io.Writer
	for _, p := range printers {
		switch p {
		case "stdout":
			ws = append(ws, os.Stdout)
		case "stderr":
			ws = append(ws, os.Stderr)
		}
	}
	switch len(ws) {
	case 0:
		return os.Stdout
	case 1:
		return ws[0]
	default:
		return io.MultiWriter(ws...)
	}
}
