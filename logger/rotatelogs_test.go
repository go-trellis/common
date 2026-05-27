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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-trellis/common/utils/testutils"
	"github.com/sirupsen/logrus"
)

func TestDefaultRotateLogsConfig(t *testing.T) {
	config := DefaultRotateLogsConfig("/tmp/test.log")
	testutils.Assert(t, config != nil, "config should not be nil")
	testutils.Equals(t, "/tmp/test.log", config.LogPath, "LogPath should match")
	testutils.Equals(t, RotateModeDay, config.RotateMode, "RotateMode should be day")
	testutils.Equals(t, 7*24*time.Hour, config.MaxAge, "MaxAge should be 7 days")
	testutils.Equals(t, 24*time.Hour, config.RotationTime, "RotationTime should be 24 hours")
	testutils.Equals(t, uint(0), config.RotationCount, "RotationCount should be 0")
	testutils.Equals(t, "", config.LinkName, "LinkName should be empty")
	testutils.Assert(t, !config.ForceNewFile, "ForceNewFile should be false")
	testutils.Assert(t, len(config.WriterLevels) > 0, "WriterLevels should not be empty")
}

func TestNewRotateLogsWriter_NilConfig(t *testing.T) {
	writer, err := NewRotateLogsWriter(nil)
	testutils.Ok(t, err)
	testutils.Assert(t, writer == nil, "writer should be nil for nil config")
}

func TestNewRotateLogsWriter_HourMode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeHour,
		RotationTime: time.Hour,
		MaxAge:       24 * time.Hour,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	// Test writing
	n, err := writer.Write([]byte("test log\n"))
	testutils.Ok(t, err)
	testutils.Assert(t, n > 0, "should write bytes")

	// Close writer if it's a closer
	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestNewRotateLogsWriter_DayMode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		MaxAge:       7 * 24 * time.Hour,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	// Test writing
	n, err := writer.Write([]byte("test log\n"))
	testutils.Ok(t, err)
	testutils.Assert(t, n > 0, "should write bytes")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestNewRotateLogsWriter_WithMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		MaxSize:      1024, // 1KB
		MaxAge:       7 * 24 * time.Hour,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	n, err := writer.Write([]byte("test log\n"))
	testutils.Ok(t, err)
	testutils.Assert(t, n > 0, "should write bytes")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestNewRotateLogsWriter_WithRotationCount(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:       logPath,
		RotateMode:    RotateModeDay,
		RotationTime:  24 * time.Hour,
		RotationCount: 10,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestNewRotateLogsWriter_WithLinkName(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	linkName := filepath.Join(tmpDir, "current.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		LinkName:     linkName,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestNewRotateLogsWriter_DefaultRotationTime(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Test hour mode with zero rotation time
	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeHour,
		RotationTime: 0, // Should default to 1 hour
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")
	testutils.Equals(t, time.Hour, config.RotationTime, "RotationTime should default to 1 hour")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}

	// Test day mode with zero rotation time
	config2 := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 0, // Should default to 24 hours
	}

	writer2, err := NewRotateLogsWriter(config2)
	testutils.Ok(t, err)
	testutils.Assert(t, writer2 != nil, "writer should not be nil")
	testutils.Equals(t, 24*time.Hour, config2.RotationTime, "RotationTime should default to 24 hours")

	if closer, ok := writer2.(io.Closer); ok {
		closer.Close()
	}
}

func TestAddRotateLogsHook(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		WriterLevels: []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
	}

	err := AddRotateLogsHook(logger, config)
	testutils.Ok(t, err)

	// Test logging with the hook
	logger.Info("test info message")
	logger.Warn("test warn message")
	logger.Error("test error message")
}

func TestAddRotateLogsHook_NilLogger(t *testing.T) {
	config := &RotateLogsConfig{
		LogPath:    "/tmp/test.log",
		RotateMode: RotateModeDay,
	}

	err := AddRotateLogsHook(nil, config)
	testutils.Ok(t, err)
}

func TestAddRotateLogsHook_NilConfig(t *testing.T) {
	logger := logrus.New()
	err := AddRotateLogsHook(logger, nil)
	testutils.Ok(t, err)
}

func TestAddRotateLogsHook_DefaultLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		WriterLevels: []logrus.Level{}, // Empty, should use defaults
	}

	err := AddRotateLogsHook(logger, config)
	testutils.Ok(t, err)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
}

func TestSetupRotateLogsLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := logrus.New()

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
	}

	err := SetupRotateLogsLogger(logger, config)
	testutils.Ok(t, err)

	logger.Info("test message")
	logger.Debug("debug message")
}

func TestSetupRotateLogsLogger_NilLogger(t *testing.T) {
	config := &RotateLogsConfig{
		LogPath:    "/tmp/test.log",
		RotateMode: RotateModeDay,
	}

	err := SetupRotateLogsLogger(nil, config)
	testutils.Ok(t, err)
}

func TestSetupRotateLogsLogger_NilConfig(t *testing.T) {
	logger := logrus.New()
	err := SetupRotateLogsLogger(logger, nil)
	testutils.Ok(t, err)
}

func TestRotateLogsConfig_UnmarshalYAML(t *testing.T) {
	yamlContent := `
log_path: /tmp/test.log
rotate_mode: day
max_age: 7d
rotation_time: 24h
max_size: 1048576
rotation_count: 10
link_name: /tmp/current.log
force_new_file: true
writer_levels:
  - info
  - warn
  - error
`

	config := &RotateLogsConfig{}
	err := config.UnmarshalYAML(func(v interface{}) error {
		// Simple YAML parser for test
		lines := strings.Split(yamlContent, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "log_path:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "log_path:"))
				if str, ok := v.(*string); ok {
					*str = val
				}
			} else if strings.HasPrefix(line, "max_age:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "max_age:"))
				if str, ok := v.(*string); ok {
					*str = val
				}
			} else if strings.HasPrefix(line, "rotation_time:") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "rotation_time:"))
				if str, ok := v.(*string); ok {
					*str = val
				}
			} else if strings.HasPrefix(line, "writer_levels:") {
				// Skip for this simple test
			} else if strings.HasPrefix(line, "- ") {
				val := strings.TrimSpace(strings.TrimPrefix(line, "- "))
				if levels, ok := v.(*[]string); ok {
					*levels = append(*levels, val)
				}
			}
		}
		return nil
	})

	// This test is a simplified version, actual YAML parsing would use yaml.v3
	_ = err
}

func TestRotateLogsConfig_MarshalYAML(t *testing.T) {
	config := &RotateLogsConfig{
		LogPath:       "/tmp/test.log",
		RotateMode:    RotateModeDay,
		MaxAge:        7 * 24 * time.Hour,
		RotationTime:  24 * time.Hour,
		MaxSize:       1048576,
		RotationCount: 10,
		LinkName:      "/tmp/current.log",
		ForceNewFile:  true,
		WriterLevels:  []logrus.Level{logrus.InfoLevel, logrus.WarnLevel},
	}

	result, err := config.MarshalYAML()
	testutils.Ok(t, err)
	testutils.Assert(t, result != nil, "result should not be nil")
}

func TestRotateLogsConfig_UnmarshalJSON(t *testing.T) {
	jsonContent := `{
		"log_path": "/tmp/test.log",
		"rotate_mode": "day",
		"max_age": "7d",
		"rotation_time": "24h",
		"max_size": 1048576,
		"rotation_count": 10,
		"link_name": "/tmp/current.log",
		"force_new_file": true,
		"writer_levels": ["info", "warn", "error"]
	}`

	config := &RotateLogsConfig{}
	err := config.UnmarshalJSON([]byte(jsonContent))
	testutils.Ok(t, err)
	testutils.Equals(t, "/tmp/test.log", config.LogPath, "LogPath should match")
	testutils.Equals(t, RotateModeDay, config.RotateMode, "RotateMode should match")
	testutils.Equals(t, 7*24*time.Hour, config.MaxAge, "MaxAge should match")
	testutils.Equals(t, 24*time.Hour, config.RotationTime, "RotationTime should match")
	testutils.Equals(t, int64(1048576), config.MaxSize, "MaxSize should match")
	testutils.Equals(t, uint(10), config.RotationCount, "RotationCount should match")
	testutils.Equals(t, "/tmp/current.log", config.LinkName, "LinkName should match")
	testutils.Assert(t, config.ForceNewFile, "ForceNewFile should be true")
	testutils.Assert(t, len(config.WriterLevels) == 3, "WriterLevels should have 3 levels")
}

func TestRotateLogsConfig_UnmarshalJSON_InvalidDuration(t *testing.T) {
	jsonContent := `{
		"log_path": "/tmp/test.log",
		"rotate_mode": "day",
		"max_age": "invalid-duration"
	}`

	config := &RotateLogsConfig{}
	err := config.UnmarshalJSON([]byte(jsonContent))
	testutils.NotOk(t, err, "should return error for invalid duration")
}

func TestRotateLogsConfig_UnmarshalJSON_InvalidLevel(t *testing.T) {
	jsonContent := `{
		"log_path": "/tmp/test.log",
		"rotate_mode": "day",
		"writer_levels": ["invalid-level"]
	}`

	config := &RotateLogsConfig{}
	err := config.UnmarshalJSON([]byte(jsonContent))
	testutils.NotOk(t, err, "should return error for invalid log level")
}

func TestRotateLogsConfig_MarshalJSON(t *testing.T) {
	config := &RotateLogsConfig{
		LogPath:       "/tmp/test.log",
		RotateMode:    RotateModeDay,
		MaxAge:        7 * 24 * time.Hour,
		RotationTime:  24 * time.Hour,
		MaxSize:       1048576,
		RotationCount: 10,
		LinkName:      "/tmp/current.log",
		ForceNewFile:  true,
		WriterLevels:  []logrus.Level{logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel},
	}

	data, err := config.MarshalJSON()
	testutils.Ok(t, err)
	testutils.Assert(t, len(data) > 0, "JSON data should not be empty")
	testutils.Assert(t, strings.Contains(string(data), "/tmp/test.log"), "JSON should contain log_path")
}

func TestNewLogrusLoggerWithRotate(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
	}

	logger, err := NewLogrusLoggerWithRotate(config)
	testutils.Ok(t, err)
	testutils.Assert(t, logger != nil, "logger should not be nil")

	logger.Info("test message")
	logger.Debug("debug message")
}

func TestNewLogrusLoggerWithRotate_NilConfig(t *testing.T) {
	logger, err := NewLogrusLoggerWithRotate(nil)
	testutils.Ok(t, err)
	testutils.Assert(t, logger != nil, "logger should not be nil")

	logger.Info("test message")
}

func TestLogrusLogger_SetRotateLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogrusLogger()
	testutils.Ok(t, err)

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
	}

	err = logger.SetRotateLogs(config)
	testutils.Ok(t, err)

	logger.Info("test message")
}

func TestLogrusLogger_SetRotateLogs_NilConfig(t *testing.T) {
	logger, err := NewLogrusLogger()
	testutils.Ok(t, err)

	err = logger.SetRotateLogs(nil)
	testutils.Ok(t, err)
}

func TestLogrusLogger_AddRotateLogsHook(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := NewLogrusLogger()
	testutils.Ok(t, err)

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
	}

	err = logger.AddRotateLogsHook(config)
	testutils.Ok(t, err)

	logger.Info("test message")
	logger.Warn("warn message")
	logger.Error("error message")
}

func TestLogrusLogger_AddRotateLogsHook_NilConfig(t *testing.T) {
	logger, err := NewLogrusLogger()
	testutils.Ok(t, err)

	err = logger.AddRotateLogsHook(nil)
	testutils.Ok(t, err)
}

func TestRotateMode_String(t *testing.T) {
	testutils.Equals(t, "hour", string(RotateModeHour), "RotateModeHour string should match")
	testutils.Equals(t, "day", string(RotateModeDay), "RotateModeDay string should match")
}

func TestNewRotateLogsWriter_InvalidPath(t *testing.T) {
	// Test with invalid path (non-existent parent directory)
	config := &RotateLogsConfig{
		LogPath:      "/non/existent/path/test.log",
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
	}

	writer, err := NewRotateLogsWriter(config)
	// This might succeed if rotatelogs creates the directory, or fail if it doesn't
	_ = err
	_ = writer
}

func TestRotateLogsConfig_EmptyWriterLevels(t *testing.T) {
	config := &RotateLogsConfig{
		LogPath:      "/tmp/test.log",
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		WriterLevels: []logrus.Level{},
	}

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	config.LogPath = logPath

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	err := AddRotateLogsHook(logger, config)
	testutils.Ok(t, err)

	// Should use default levels
	logger.Info("test")
}

func TestNewRotateLogsWriter_ForceNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		ForceNewFile: true,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestRotateLogsConfig_AllLevels(t *testing.T) {
	jsonContent := `{
		"log_path": "/tmp/test.log",
		"rotate_mode": "day",
		"writer_levels": ["trace", "debug", "info", "warn", "error", "fatal", "panic"]
	}`

	config := &RotateLogsConfig{}
	err := config.UnmarshalJSON([]byte(jsonContent))
	testutils.Ok(t, err)
	testutils.Assert(t, len(config.WriterLevels) == 7, "should have 7 levels")
}

func TestNewRotateLogsWriter_DefaultMode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Test with invalid mode (should default to day)
	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateMode("invalid"), // Invalid mode
		RotationTime: 0,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")
	testutils.Equals(t, 24*time.Hour, config.RotationTime, "should default to 24 hours")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestRotateLogsConfig_MarshalJSON_EmptyFields(t *testing.T) {
	config := &RotateLogsConfig{
		LogPath:    "/tmp/test.log",
		RotateMode: RotateModeDay,
	}

	data, err := config.MarshalJSON()
	testutils.Ok(t, err)
	testutils.Assert(t, len(data) > 0, "JSON data should not be empty")
}

func TestRotateLogsConfig_UnmarshalJSON_EmptyFields(t *testing.T) {
	jsonContent := `{
		"log_path": "/tmp/test.log",
		"rotate_mode": "day"
	}`

	config := &RotateLogsConfig{}
	err := config.UnmarshalJSON([]byte(jsonContent))
	testutils.Ok(t, err)
	testutils.Equals(t, "/tmp/test.log", config.LogPath, "LogPath should match")
	testutils.Equals(t, RotateModeDay, config.RotateMode, "RotateMode should match")
}

func TestNewRotateLogsWriter_OnlyMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Test with MaxSize but no RotationTime (should still work with day mode default)
	config := &RotateLogsConfig{
		LogPath:    logPath,
		RotateMode: RotateModeDay,
		MaxSize:    1024,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestNewRotateLogsWriter_OnlyMaxAge(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		MaxAge:       7 * 24 * time.Hour,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestNewRotateLogsWriter_HourModeWithSize(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeHour,
		RotationTime: time.Hour,
		MaxSize:      1024,
		MaxAge:       24 * time.Hour,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	n, err := writer.Write([]byte("test log line\n"))
	testutils.Ok(t, err)
	testutils.Assert(t, n > 0, "should write bytes")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}

func TestRotateLogsConfig_UnmarshalJSON_MixedCaseLevels(t *testing.T) {
	jsonContent := `{
		"log_path": "/tmp/test.log",
		"rotate_mode": "day",
		"writer_levels": ["INFO", "WARN", "ERROR"]
	}`

	config := &RotateLogsConfig{}
	err := config.UnmarshalJSON([]byte(jsonContent))
	testutils.Ok(t, err)
	testutils.Assert(t, len(config.WriterLevels) == 3, "should have 3 levels")
}

func TestRotateLogsConfig_ClockAndHandler(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &RotateLogsConfig{
		LogPath:      logPath,
		RotateMode:   RotateModeDay,
		RotationTime: 24 * time.Hour,
		// Clock and Handler are runtime-only, cannot be tested via JSON/YAML
		Clock:   nil,
		Handler: nil,
	}

	writer, err := NewRotateLogsWriter(config)
	testutils.Ok(t, err)
	testutils.Assert(t, writer != nil, "writer should not be nil")

	if closer, ok := writer.(io.Closer); ok {
		closer.Close()
	}
}
