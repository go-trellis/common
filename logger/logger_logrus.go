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
	"errors"
	"io"
	"maps"

	"github.com/sirupsen/logrus"
	"xorm.io/xorm/log"
)

var (
	_ Logger     = (*LogrusLogger)(nil)
	_ log.Logger = (*LogrusLogger)(nil)
)

func NewWithLogrusLogger(l *logrus.Logger) Logger {
	if l == nil {
		return &noop{}
	}
	return &LogrusLogger{logger: l}
}

type LogrusLogger struct {
	logger    *logrus.Logger
	isShowSQL bool
}

func NewLogrusLogger() (*LogrusLogger, error) {
	// Create a null logger if no output is configured
	nullLogger := logrus.New()
	nullLogger.SetOutput(io.Discard)

	ll := &LogrusLogger{
		logger: nullLogger,
	}

	return ll, nil
}

// NewLogrusLoggerWithRotate creates a new logrus logger with file rotation
func NewLogrusLoggerWithRotate(config *RotateLogsConfig) (*LogrusLogger, error) {
	logger := logrus.New()
	
	if config != nil {
		if err := SetupRotateLogsLogger(logger, config); err != nil {
			return nil, err
		}
	} else {
		logger.SetOutput(io.Discard)
	}

	ll := &LogrusLogger{
		logger: logger,
	}

	return ll, nil
}

// SetRotateLogs sets up file rotation for the logger
func (p *LogrusLogger) SetRotateLogs(config *RotateLogsConfig) error {
	if p.logger == nil || config == nil {
		return nil
	}
	return SetupRotateLogsLogger(p.logger, config)
}

// AddRotateLogsHook adds a file rotation hook to the logger
// This allows logs to be written to both the default output and rotated files
func (p *LogrusLogger) AddRotateLogsHook(config *RotateLogsConfig) error {
	if p.logger == nil || config == nil {
		return nil
	}
	return AddRotateLogsHook(p.logger, config)
}

// With creates a child logger with specified fields
func (p *LogrusLogger) With(kvs ...any) Logger {
	lenFields := len(kvs)
	fields := make(logrus.Fields)

	for i := 0; i < lenFields; i += 2 {
		k := kvs[i]
		var v any = errors.New("MISSING VALUE")
		if i+1 < lenFields {
			v = kvs[i+1]
		}
		fields[toString(k)] = v
	}

	// Create a wrapper logger that uses WithFields for all log calls
	return &logrusLoggerWithFields{
		logger:    p.logger,
		fields:    fields,
		isShowSQL: p.isShowSQL,
	}
}

// logrusLoggerWithFields wraps a logger with fields applied to all log calls
type logrusLoggerWithFields struct {
	logger    *logrus.Logger
	fields    logrus.Fields
	isShowSQL bool
}

func (p *logrusLoggerWithFields) With(kvs ...any) Logger {
	// Merge new fields with existing fields
	lenFields := len(kvs)
	newFields := make(logrus.Fields)
	maps.Copy(newFields, p.fields)

	for i := 0; i < lenFields; i += 2 {
		k := kvs[i]
		var v any = errors.New("MISSING VALUE")
		if i+1 < lenFields {
			v = kvs[i+1]
		}
		newFields[toString(k)] = v
	}

	return &logrusLoggerWithFields{
		logger:    p.logger,
		fields:    newFields,
		isShowSQL: p.isShowSQL,
	}
}

func (p *logrusLoggerWithFields) Log(kvs ...any) error {
	p.logger.WithFields(p.fields).Info(kvs...)
	return nil
}

func (p *logrusLoggerWithFields) Debug(kvs ...any) {
	p.logger.WithFields(p.fields).Debug(kvs...)
}

func (p *logrusLoggerWithFields) Debugf(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Debugf(msg, kvs...)
}

func (p *logrusLoggerWithFields) Info(kvs ...any) {
	p.logger.WithFields(p.fields).Info(kvs...)
}

func (p *logrusLoggerWithFields) Infof(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Infof(msg, kvs...)
}

func (p *logrusLoggerWithFields) Warn(kvs ...any) {
	p.logger.WithFields(p.fields).Warn(kvs...)
}

func (p *logrusLoggerWithFields) Warnf(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Warnf(msg, kvs...)
}

func (p *logrusLoggerWithFields) Error(kvs ...any) {
	p.logger.WithFields(p.fields).Error(kvs...)
}

func (p *logrusLoggerWithFields) Errorf(msg string, kvs ...any) {
	p.logger.WithFields(p.fields).Errorf(msg, kvs...)
}

func (p *logrusLoggerWithFields) Level() log.LogLevel {
	if p.logger != nil {
		tempLogger := &LogrusLogger{logger: p.logger}
		return tempLogger.Level()
	}
	return log.LOG_DEBUG
}

func (p *logrusLoggerWithFields) SetLevel(l log.LogLevel) {
	if p.logger != nil {
		tempLogger := &LogrusLogger{logger: p.logger}
		tempLogger.SetLevel(l)
	}
}

func (p *logrusLoggerWithFields) ShowSQL(show ...bool) {
	p.isShowSQL = len(show) == 0 || (len(show) > 0 && show[0])
}

func (p *logrusLoggerWithFields) IsShowSQL() bool {
	return p.isShowSQL
}

func (p *logrusLoggerWithFields) Writer() io.Writer {
	if p.logger != nil {
		return p.logger.Out
	}
	return io.Discard
}

// Log prints log with kvs
func (p *LogrusLogger) Log(kvs ...any) error {
	p.Info(kvs...)
	return nil
}

// Debug prints debug information
func (p *LogrusLogger) Debug(kvs ...any) {
	p.logger.Debug(kvs...)
}

// Debugf format prints debug information
func (p *LogrusLogger) Debugf(msg string, kvs ...any) {
	p.logger.Debugf(msg, kvs...)
}

// Info prints info information
func (p *LogrusLogger) Info(kvs ...any) {
	p.logger.Info(kvs...)
}

// Infof format prints info information
func (p *LogrusLogger) Infof(msg string, kvs ...any) {
	p.logger.Infof(msg, kvs...)
}

// Warn prints warn information
func (p *LogrusLogger) Warn(kvs ...any) {
	p.logger.Warn(kvs...)
}

// Warnf format prints warn information
func (p *LogrusLogger) Warnf(msg string, kvs ...any) {
	p.logger.Warnf(msg, kvs...)
}

// Error prints error information
func (p *LogrusLogger) Error(kvs ...any) {
	p.logger.Error(kvs...)
}

// Errorf format prints error information
func (p *LogrusLogger) Errorf(msg string, kvs ...any) {
	p.logger.Errorf(msg, kvs...)
}

// Level returns current log level
func (p *LogrusLogger) Level() log.LogLevel {
	switch p.logger.GetLevel() {
	case logrus.DebugLevel:
		return log.LOG_DEBUG
	case logrus.InfoLevel:
		return log.LOG_INFO
	case logrus.WarnLevel:
		return log.LOG_WARNING
	case logrus.ErrorLevel:
		return log.LOG_ERR
	case logrus.PanicLevel, logrus.FatalLevel:
		return log.LOG_OFF
	default:
		return log.LOG_DEBUG
	}
}

// SetLevel sets log level
func (p *LogrusLogger) SetLevel(l log.LogLevel) {
	switch l {
	case log.LOG_DEBUG:
		p.logger.SetLevel(logrus.DebugLevel)
	case log.LOG_INFO:
		p.logger.SetLevel(logrus.InfoLevel)
	case log.LOG_WARNING:
		p.logger.SetLevel(logrus.WarnLevel)
	case log.LOG_ERR:
		p.logger.SetLevel(logrus.ErrorLevel)
	case log.LOG_OFF:
		p.logger.SetLevel(logrus.PanicLevel)
	}
}

// ShowSQL sets whether to show SQL
func (p *LogrusLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		p.isShowSQL = show[0]
	} else {
		p.isShowSQL = true
	}
}

// IsShowSQL returns whether SQL is shown
func (p *LogrusLogger) IsShowSQL() bool {
	return p.isShowSQL
}

func (p *LogrusLogger) Writer() io.Writer {
	return p.logger.Out
}
