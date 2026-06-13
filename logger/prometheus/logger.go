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
	"github.com/go-kit/log"
	"github.com/go-trellis/common/logger"
)

type Config struct {
	FileName     string `yaml:"filename" json:"filename"`
	Level        string `yaml:"level" json:"level"`
	MoveFileType int    `yaml:"move_file_type" json:"move_file_type"`
	MaxLength    int64  `yaml:"max_length" json:"max_length"`
	MaxBackups   uint   `yaml:"max_backups" json:"max_backups"`
	Caller       bool   `yaml:"caller" json:"caller"`
}

type PromeNoonLogger struct{}

func (p *PromeNoonLogger) Log(...any) error {
	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Config.
func (p *Config) UnmarshalYAML(unmarshal func(any) error) error {
	type plain Config
	return unmarshal((*plain)(p))
}

func New(config *Config) (log.Logger, error) {
	if config.FileName == "" {
		return &PromeNoonLogger{}, nil
	}

	rotateCfg := &logger.RotateConfig{
		FileName:      config.FileName,
		MoveFileType:  logger.MoveFileType(config.MoveFileType),
		RotationSize:  config.MaxLength,
		RotationCount: config.MaxBackups,
		Caller:        config.Caller,
	}

	rotator, err := logger.NewRotateLogs(rotateCfg)
	if err != nil {
		return nil, err
	}

	return log.NewJSONLogger(rotator), nil
}
