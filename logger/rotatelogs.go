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
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-trellis/common/utils/types"
	"github.com/sirupsen/logrus"
	writerhook "github.com/sirupsen/logrus/hooks/writer"
)

// RotateMode defines the log rotation mode
type RotateMode string

const (
	// RotateModeHour rotates logs by hour (and optionally by size if MaxSize is set)
	RotateModeHour RotateMode = "hour"
	// RotateModeDay rotates logs by day (and optionally by size if MaxSize is set)
	RotateModeDay RotateMode = "day"
)

// RotateLogsConfig configures the file rotation logger
type RotateLogsConfig struct {
	// LogPath is the path of the active log file. Rotated archives are stored as LogPath.{period}.
	LogPath string `yaml:"log_path" json:"log_path"`

	// RotateMode defines how logs are rotated: "hour" or "day"
	// When MaxSize is set, logs will also be rotated by size in parallel with time-based rotation
	RotateMode RotateMode `yaml:"rotate_mode" json:"rotate_mode"`

	// MaxAge is the maximum age of log files to keep (0 means no limit)
	// Format: "1h", "24h", "7d", etc.
	MaxAge time.Duration `yaml:"max_age" json:"max_age"`

	// RotationTime is the rotation interval
	// For hour mode: typically 1 hour
	// For day mode: typically 24 hours
	// Format: "1h", "24h", etc.
	RotationTime time.Duration `yaml:"rotation_time" json:"rotation_time"`

	// MaxSize is the maximum size in bytes before rotation
	// If set, logs will be rotated when either MaxSize is reached OR RotationTime interval is reached
	// Works with both "hour" and "day" modes
	MaxSize int64 `yaml:"max_size" json:"max_size"`

	// RotationCount is the maximum number of rotated files to keep (0 means no limit)
	RotationCount uint `yaml:"rotation_count" json:"rotation_count"`

	// ForceNewFile forces rotation even if the file doesn't exist
	ForceNewFile bool `yaml:"force_new_file" json:"force_new_file"`

	// WriterLevels specifies which log levels should be written to this writer
	// Format: ["debug", "info", "warn", "error", "fatal", "panic"]
	WriterLevels []logrus.Level `yaml:"writer_levels" json:"writer_levels"`
}

// DefaultRotateLogsConfig returns a default configuration for rotate logs
func DefaultRotateLogsConfig(logPath string) *RotateLogsConfig {
	return &RotateLogsConfig{
		LogPath:       logPath,
		RotateMode:    RotateModeDay,
		MaxAge:        7 * 24 * time.Hour, // Keep logs for 7 days
		RotationTime:  24 * time.Hour,     // Rotate daily
		RotationCount: 0,                  // No limit
		ForceNewFile:  false,
		WriterLevels:  []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel},
	}
}

// NewRotateLogsWriter creates a new rotatelogs writer based on the configuration
func NewRotateLogsWriter(config *RotateLogsConfig) (io.Writer, error) {
	if config == nil {
		return nil, nil
	}
	return newRotatingFileWriter(config)
}

// AddRotateLogsHook adds a file rotation hook to the logrus logger
func AddRotateLogsHook(logger *logrus.Logger, config *RotateLogsConfig) error {
	if logger == nil || config == nil {
		return nil
	}

	writer, err := NewRotateLogsWriter(config)
	if err != nil {
		return err
	}

	levels := config.WriterLevels
	if len(levels) == 0 {
		levels = []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel}
	}

	hook := &writerhook.Hook{
		Writer:    writer,
		LogLevels: levels,
	}
	logger.AddHook(hook)

	return nil
}

// SetupRotateLogsLogger configures a logrus logger with file rotation
func SetupRotateLogsLogger(logger *logrus.Logger, config *RotateLogsConfig) error {
	if logger == nil {
		return nil
	}

	if config == nil {
		return nil
	}

	writer, err := NewRotateLogsWriter(config)
	if err != nil {
		return err
	}

	logger.SetOutput(writer)

	return nil
}

func parseRotateDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	duration := types.ParseStringTime(strings.ToLower(s))
	if duration == 0 && s != "0" {
		if d, err := time.ParseDuration(s); err == nil {
			return d, nil
		}
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	return duration, nil
}

func parseWriterLevels(levels []string) ([]logrus.Level, error) {
	if len(levels) == 0 {
		return nil, nil
	}
	parsed := make([]logrus.Level, 0, len(levels))
	for _, levelStr := range levels {
		level, err := logrus.ParseLevel(strings.ToLower(levelStr))
		if err != nil {
			return nil, fmt.Errorf("invalid log level %q: %w", levelStr, err)
		}
		parsed = append(parsed, level)
	}
	return parsed, nil
}

func formatWriterLevels(levels []logrus.Level) []string {
	if len(levels) == 0 {
		return nil
	}
	formatted := make([]string, 0, len(levels))
	for _, level := range levels {
		formatted = append(formatted, strings.ToLower(level.String()))
	}
	return formatted
}

func applyRotateLogsConfigFields(r *RotateLogsConfig, maxAge, rotationTime string, writerLevels []string) error {
	if maxAge != "" {
		duration, err := parseRotateDuration(maxAge)
		if err != nil {
			return fmt.Errorf("invalid max_age duration: %s", maxAge)
		}
		r.MaxAge = duration
	}

	if rotationTime != "" {
		duration, err := parseRotateDuration(rotationTime)
		if err != nil {
			return fmt.Errorf("invalid rotation_time duration: %s", rotationTime)
		}
		r.RotationTime = duration
	}

	if len(writerLevels) > 0 {
		levels, err := parseWriterLevels(writerLevels)
		if err != nil {
			return err
		}
		r.WriterLevels = levels
	}

	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for WriterLevels
func (r *RotateLogsConfig) UnmarshalYAML(unmarshal func(any) error) error {
	type Alias RotateLogsConfig
	aux := &struct {
		MaxAge       string   `yaml:"max_age"`
		RotationTime string   `yaml:"rotation_time"`
		WriterLevels []string `yaml:"writer_levels"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := unmarshal(aux); err != nil {
		return err
	}

	return applyRotateLogsConfigFields(r, aux.MaxAge, aux.RotationTime, aux.WriterLevels)
}

// MarshalYAML implements yaml.Marshaler for WriterLevels
func (r RotateLogsConfig) MarshalYAML() (any, error) {
	type Alias RotateLogsConfig
	return &struct {
		MaxAge       string   `yaml:"max_age"`
		RotationTime string   `yaml:"rotation_time"`
		WriterLevels []string `yaml:"writer_levels"`
		*Alias
	}{
		MaxAge:       r.MaxAge.String(),
		RotationTime: r.RotationTime.String(),
		WriterLevels: formatWriterLevels(r.WriterLevels),
		Alias:        (*Alias)(&r),
	}, nil
}

// UnmarshalJSON implements json.Unmarshaler for RotateLogsConfig
func (r *RotateLogsConfig) UnmarshalJSON(data []byte) error {
	type Alias RotateLogsConfig
	aux := &struct {
		MaxAge       string   `json:"max_age"`
		RotationTime string   `json:"rotation_time"`
		WriterLevels []string `json:"writer_levels"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	return applyRotateLogsConfigFields(r, aux.MaxAge, aux.RotationTime, aux.WriterLevels)
}

// MarshalJSON implements json.Marshaler for RotateLogsConfig
func (r RotateLogsConfig) MarshalJSON() ([]byte, error) {
	type Alias RotateLogsConfig
	aux := &struct {
		MaxAge       string   `json:"max_age"`
		RotationTime string   `json:"rotation_time"`
		WriterLevels []string `json:"writer_levels"`
		*Alias
	}{
		MaxAge:       r.MaxAge.String(),
		RotationTime: r.RotationTime.String(),
		WriterLevels: formatWriterLevels(r.WriterLevels),
		Alias:        (*Alias)(&r),
	}

	return json.Marshal(aux)
}
