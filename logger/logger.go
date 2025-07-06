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
	"time"

	"github.com/sirupsen/logrus"
)

type MoveFileType int

const (
	MoveFileTypeNone MoveFileType = iota
	MoveFileTypePerMinite
	MoveFileTypeHourly
	MoveFileTypeDaily
)

func (p MoveFileType) Duration() time.Duration {
	switch p {
	case MoveFileTypePerMinite:
		return time.Minute
	case MoveFileTypeHourly:
		return time.Hour
	case MoveFileTypeDaily:
		return time.Hour * 24
	}
	return 0
}

// Level log level
type Level logrus.Level
type Logger logrus.FieldLogger

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

// ToLevelName convert level into string name
func ToLevelName(lvl Level) string {
	return logrus.Level(lvl).String()
}

// ToLevel convert string name into level
func ToLevel(lvl string) (Level, error) {
	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		return 0, err
	}
	return Level(level), nil
}

type LogrusConfig struct {
	Level         Level
	ReportCaller  bool
	Formatter     logrus.Formatter
	Configs       []any
	BufferPool    logrus.BufferPool
	DefaultWriter io.Writer
}

type TextFormatter = logrus.TextFormatter
type JSONFormatter = logrus.JSONFormatter

func NewLogger(c *LogrusConfig) (*logrus.Logger, error) {
	if c == nil {
		return Noop(), nil
	}

	logrusLogger := logrus.New()
	logrusLogger.SetReportCaller(c.ReportCaller)

	ok := false
	for _, v := range logrus.AllLevels {
		if c.Level == Level(v) {
			ok = true
			break
		}
	}
	if ok {
		logrusLogger.SetLevel(logrus.Level(c.Level))
	}

	if c.Formatter != nil {
		logrusLogger.SetFormatter(c.Formatter)
	} else {
		c.Formatter = logrusLogger.Formatter
	}

	writer, err := NewLugrusHook(c.Formatter, c.Configs)
	if err != nil {
		return nil, err
	}
	if writer != nil {
		logrusLogger.AddHook(writer)
	}

	if c.BufferPool != nil {
		logrusLogger.SetBufferPool(c.BufferPool)
	}

	if c.DefaultWriter != nil {
		logrusLogger.SetOutput(c.DefaultWriter)
	}

	return logrusLogger, nil
}
