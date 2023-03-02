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

package prometheus

import (
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.uber.org/zap/zapcore"
	"trellis.tech/trellis/common.v1/logger"
)

type Config struct {
	Level        string   `yaml:"level" json:"level"`
	FileName     string   `yaml:"filename" json:"filename"`
	MoveFileType int      `yaml:"move_file_type" json:"move_file_type"`
	MaxLength    int64    `yaml:"max_length" json:"max_length"`
	MaxBackups   int      `yaml:"max_backups" json:"max_backups"`
	StdPrinters  []string `yaml:"std_printers" json:"std_printers"`
	TimeFormat   string   `yaml:"time_format" json:"time_format"`
	Caller       bool     `yaml:"caller" json:"caller"`
	CallerSkip   int      `yaml:"caller_skip" json:"caller_skip"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Config.
func (p *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	return unmarshal((*plain)(p))
}

func New(config *Config) log.Logger {
	options := []logger.Option{
		logger.LogFileOption(
			logger.OptionFilename(config.FileName),
			logger.OptionMoveFileType(logger.MoveFileType(config.MoveFileType)),
			logger.OptionMaxLength(config.MaxLength),
			logger.OptionMaxBackups(config.MaxBackups),
			logger.OptionStdPrinters(config.StdPrinters),
		),
		logger.EncoderConfig(&zapcore.EncoderConfig{}),
		logger.CallerSkip(config.CallerSkip),
	}
	if config.Caller {
		options = append(options, logger.Caller())
	}

	stdLog, err := logger.NewLogger(options...)
	if err != nil {
		panic(err)
	}

	if config.TimeFormat == "" {
		config.TimeFormat = "2006-01-02T15:04:05.000Z07:00"
	}

	timestampFormat := log.TimestampFormat(
		func() time.Time { return time.Now() },
		config.TimeFormat,
	)
	l := level.NewFilter(stdLog, getLevel(config.Level))
	l = log.With(l, "ts", timestampFormat)

	return l
}

// Set get the value of the allowed level.
func getLevel(s string) level.Option {
	switch s {
	case "debug":
		return level.AllowDebug()
	case "info":
		return level.AllowInfo()
	case "warn":
		return level.AllowWarn()
	case "error":
		return level.AllowError()
	default:
		return level.AllowInfo()
	}
}
