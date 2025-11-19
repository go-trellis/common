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

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	writerhook "github.com/sirupsen/logrus/hooks/writer"
	"trellis.tech/trellis/common.v3/utils/types"
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
	// LogPath is the base path for log files
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

	// LinkName is the symbolic link name pointing to the current log file
	LinkName string `yaml:"link_name" json:"link_name"`

	// ForceNewFile forces rotation even if the file doesn't exist
	ForceNewFile bool `yaml:"force_new_file" json:"force_new_file"`

	// Clock allows customization of the clock used for rotation
	// Not serialized to YAML/JSON as it's a runtime interface
	Clock rotatelogs.Clock `yaml:"-" json:"-"`

	// Handler allows customization of the rotation handler
	// Not serialized to YAML/JSON as it's a runtime interface
	Handler rotatelogs.Handler `yaml:"-" json:"-"`

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
		LinkName:      "",
		ForceNewFile:  false,
		WriterLevels:  []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel},
	}
}

// NewRotateLogsWriter creates a new rotatelogs writer based on the configuration
func NewRotateLogsWriter(config *RotateLogsConfig) (io.Writer, error) {
	if config == nil {
		return nil, nil
	}

	var options []rotatelogs.Option

	// Set rotation time based on mode
	switch config.RotateMode {
	case RotateModeHour:
		if config.RotationTime == 0 {
			config.RotationTime = time.Hour
		}
		options = append(options, rotatelogs.WithRotationTime(config.RotationTime))
	case RotateModeDay:
		if config.RotationTime == 0 {
			config.RotationTime = 24 * time.Hour
		}
		options = append(options, rotatelogs.WithRotationTime(config.RotationTime))
	default:
		// Default to day mode
		if config.RotationTime == 0 {
			config.RotationTime = 24 * time.Hour
		}
		options = append(options, rotatelogs.WithRotationTime(config.RotationTime))
	}

	// Add size-based rotation if MaxSize is set (works with both hour and day modes)
	if config.MaxSize > 0 {
		options = append(options, rotatelogs.WithRotationSize(config.MaxSize))
	}

	// Set max age
	if config.MaxAge > 0 {
		options = append(options, rotatelogs.WithMaxAge(config.MaxAge))
	}

	// Set rotation count
	if config.RotationCount > 0 {
		options = append(options, rotatelogs.WithRotationCount(config.RotationCount))
	}

	// Set link name
	if config.LinkName != "" {
		options = append(options, rotatelogs.WithLinkName(config.LinkName))
	}

	// Set force new file
	if config.ForceNewFile {
		options = append(options, rotatelogs.ForceNewFile())
	}

	// Set clock
	if config.Clock != nil {
		options = append(options, rotatelogs.WithClock(config.Clock))
	}

	// Set handler
	if config.Handler != nil {
		options = append(options, rotatelogs.WithHandler(config.Handler))
	}

	// Generate filename pattern based on mode
	var filenamePattern string
	switch config.RotateMode {
	case RotateModeHour:
		filenamePattern = config.LogPath + ".%Y%m%d%H"
	case RotateModeDay:
		filenamePattern = config.LogPath + ".%Y%m%d"
	default:
		filenamePattern = config.LogPath + ".%Y%m%d"
	}

	writer, err := rotatelogs.New(filenamePattern, options...)
	if err != nil {
		return nil, err
	}

	return writer, nil
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

	// Use writer hook to write logs to the rotated file
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

	// Create rotatelogs writer
	writer, err := NewRotateLogsWriter(config)
	if err != nil {
		return err
	}

	// Set the writer as the logger output
	logger.SetOutput(writer)

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

	// Parse MaxAge
	if aux.MaxAge != "" {
		duration := types.ParseStringTime(strings.ToLower(aux.MaxAge))
		if duration == 0 && aux.MaxAge != "0" && aux.MaxAge != "" {
			// Try standard time.ParseDuration as fallback
			if d, err := time.ParseDuration(aux.MaxAge); err == nil {
				duration = d
			} else {
				return fmt.Errorf("invalid max_age duration: %s", aux.MaxAge)
			}
		}
		r.MaxAge = duration
	}

	// Parse RotationTime
	if aux.RotationTime != "" {
		duration := types.ParseStringTime(strings.ToLower(aux.RotationTime))
		if duration == 0 && aux.RotationTime != "0" && aux.RotationTime != "" {
			// Try standard time.ParseDuration as fallback
			if d, err := time.ParseDuration(aux.RotationTime); err == nil {
				duration = d
			} else {
				return fmt.Errorf("invalid rotation_time duration: %s", aux.RotationTime)
			}
		}
		r.RotationTime = duration
	}

	// Parse WriterLevels
	if len(aux.WriterLevels) > 0 {
		levels := make([]logrus.Level, 0, len(aux.WriterLevels))
		for _, levelStr := range aux.WriterLevels {
			level, err := logrus.ParseLevel(strings.ToLower(levelStr))
			if err != nil {
				return fmt.Errorf("invalid log level %q: %w", levelStr, err)
			}
			levels = append(levels, level)
		}
		r.WriterLevels = levels
	}

	return nil
}

// MarshalYAML implements yaml.Marshaler for WriterLevels
func (r RotateLogsConfig) MarshalYAML() (any, error) {
	type Alias RotateLogsConfig
	aux := &struct {
		MaxAge       string   `yaml:"max_age"`
		RotationTime string   `yaml:"rotation_time"`
		WriterLevels []string `yaml:"writer_levels"`
		*Alias
	}{
		MaxAge:       r.MaxAge.String(),
		RotationTime: r.RotationTime.String(),
		Alias:        (*Alias)(&r),
	}

	// Convert WriterLevels to strings
	aux.WriterLevels = make([]string, 0, len(r.WriterLevels))
	for _, level := range r.WriterLevels {
		aux.WriterLevels = append(aux.WriterLevels, strings.ToLower(level.String()))
	}

	return aux, nil
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

	// Parse MaxAge
	if aux.MaxAge != "" {
		duration := types.ParseStringTime(strings.ToLower(aux.MaxAge))
		if duration == 0 && aux.MaxAge != "0" && aux.MaxAge != "" {
			// Try standard time.ParseDuration as fallback
			if d, err := time.ParseDuration(aux.MaxAge); err == nil {
				duration = d
			} else {
				return fmt.Errorf("invalid max_age duration: %s", aux.MaxAge)
			}
		}
		r.MaxAge = duration
	}

	// Parse RotationTime
	if aux.RotationTime != "" {
		duration := types.ParseStringTime(strings.ToLower(aux.RotationTime))
		if duration == 0 && aux.RotationTime != "0" && aux.RotationTime != "" {
			// Try standard time.ParseDuration as fallback
			if d, err := time.ParseDuration(aux.RotationTime); err == nil {
				duration = d
			} else {
				return fmt.Errorf("invalid rotation_time duration: %s", aux.RotationTime)
			}
		}
		r.RotationTime = duration
	}

	// Parse WriterLevels
	if len(aux.WriterLevels) > 0 {
		levels := make([]logrus.Level, 0, len(aux.WriterLevels))
		for _, levelStr := range aux.WriterLevels {
			level, err := logrus.ParseLevel(strings.ToLower(levelStr))
			if err != nil {
				return fmt.Errorf("invalid log level %q: %w", levelStr, err)
			}
			levels = append(levels, level)
		}
		r.WriterLevels = levels
	}

	return nil
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
		Alias:        (*Alias)(&r),
	}

	// Convert WriterLevels to strings
	aux.WriterLevels = make([]string, 0, len(r.WriterLevels))
	for _, level := range r.WriterLevels {
		aux.WriterLevels = append(aux.WriterLevels, strings.ToLower(level.String()))
	}

	return json.Marshal(aux)
}
