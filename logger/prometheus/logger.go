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
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"trellis.tech/trellis/common.v2/errcode"
	"trellis.tech/trellis/common.v2/logger"
)

type Config struct {
	FileName     string `yaml:"filename" json:"filename"`
	Level        string `yaml:"level" json:"level"`
	MoveFileType int    `yaml:"move_file_type" json:"move_file_type"`
	MaxLength    int64  `yaml:"max_length" json:"max_length"`
	MaxBackups   uint   `yaml:"max_backups" json:"max_backups"`
	Caller       bool   `yaml:"caller" json:"caller"`
}

type PrometheusLogger struct {
	Logger *logrus.Logger
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Config.
func (p *Config) UnmarshalYAML(unmarshal func(any) error) error {
	type plain Config
	return unmarshal((*plain)(p))
}

func New(config *Config) (log.Logger, error) {
	var options []rotatelogs.Option

	if config.FileName == "" {
		return nil, errcode.New("filename is empty")
	}
	_, filename := filepath.Split(config.FileName)
	options = append(options, rotatelogs.WithLinkName(filename))

	if config.MoveFileType > 0 {
		moveType := logger.MoveFileType(config.MoveFileType)

		switch moveType {
		case logger.MoveFileTypePerMinite:
			config.FileName += ".%Y%m%d%H%M"
		case logger.MoveFileTypeHourly:
			config.FileName += ".%Y%m%d%H"
		case logger.MoveFileTypeDaily:
			fallthrough
		default:
			moveType = logger.MoveFileTypeDaily
			config.FileName += ".%Y%m%d"
		}
		options = append(options, rotatelogs.WithRotationTime(moveType.Duration()))
	}

	if config.MaxLength > 0 {
		options = append(options, rotatelogs.WithRotationSize(config.MaxLength))
	}

	if config.MaxBackups > 0 {
		options = append(options, rotatelogs.WithRotationCount(config.MaxBackups))
	}

	rotator, err := logger.NewRotateLogs(config.FileName, options...)
	if err != nil {
		return nil, err
	}

	kitLog := log.NewJSONLogger(rotator)

	level.Error(kitLog).Log()

	return kitLog, nil
}
