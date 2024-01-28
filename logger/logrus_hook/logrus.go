/*
Copyright © 2023 Henry Huang <hhh@rutcode.com>

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

package logrushook

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"trellis.tech/common.v2/logger"
)

type LogrusConfig struct {
	Level         logrus.Level
	ReportCaller  bool
	DefaultWirter io.Writer
	Formatter     logrus.Formatter

	WriterInfo interface{}
}

type TextFormatter = logrus.TextFormatter
type JSONFormatter = logrus.JSONFormatter

type LogrusLevelFileConfigs []*LogrusLevelFileConfig

type LogrusLevelFileConfig struct {
	Levels  []logrus.Level     `json:"levels" yaml:"levels"`
	Options logger.FileOptions `json:",inline" yaml:",inline"`
}

func NewLogrus(c *LogrusConfig) (*logrus.Logger, error) {
	logrusLogger := logrus.New()
	logrusLogger.SetReportCaller(c.ReportCaller)

	ok := false
	for _, v := range logrus.AllLevels {
		if c.Level == v {
			ok = true
			break
		}
	}
	if ok {
		logrusLogger.SetLevel(c.Level)
	} else {
		logrusLogger.SetLevel(logrus.WarnLevel)
	}

	if c.Formatter != nil {
		logrusLogger.SetFormatter(c.Formatter)
	}

	if c.DefaultWirter != nil {
		logrusLogger.SetOutput(c.DefaultWirter)
	}

	// 空对象，则直接返回
	if c.WriterInfo == nil {
		return logrusLogger, nil
	}

	switch t := c.WriterInfo.(type) {
	case string:
		flog, err := logger.NewFileLogger(logger.OptionFilename(t))
		if err != nil {
			return nil, err
		}
		logrusWriterHook(logrusLogger, logrus.AllLevels, flog)
	case *logger.FileOptions:
		flog, err := logger.NewFileLoggerWithOptions(*t)
		if err != nil {
			return nil, err
		}
		logrusWriterHook(logrusLogger, logrus.AllLevels, flog)
	case map[logrus.Level]*logger.FileOptions:
		for l, options := range t {
			err := logrusWriterFileOptions(logrusLogger, *options, []logrus.Level{l})
			if err != nil {
				return nil, err
			}
		}
	case LogrusLevelFileConfigs:
		for _, options := range t {
			err := logrusWriterFileOptions(logrusLogger, options.Options, options.Levels)
			if err != nil {
				return nil, err
			}
		}
	case []*LogrusLevelFileConfig:
		for _, options := range t {
			err := logrusWriterFileOptions(logrusLogger, options.Options, options.Levels)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported writer info")
	}

	return logrusLogger, nil
}

func logrusWriterFileOptions(log *logrus.Logger, options logger.FileOptions, levels []logrus.Level) error {
	if len(levels) == 0 {
		return nil
	}
	flog, err := logger.NewFileLoggerWithOptions(options)
	if err != nil {
		return err
	}
	logrusWriterHook(log, levels, flog)
	return nil
}

func logrusWriterHook(logger *logrus.Logger, levels []logrus.Level, w io.Writer) {
	logger.AddHook(&writer.Hook{Writer: w, LogLevels: levels})
}
