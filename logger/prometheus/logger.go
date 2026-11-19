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
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
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
	if config == nil || config.FileName == "" {
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

	allow, err := allowOption(config.Level)
	if err != nil {
		return nil, err
	}

	// Filter first, then attach caller so log.DefaultCaller depth stays correct
	// when callers use level.Info/Debug(...).Log (see go-kit level.TestLevelContext).
	// Always stamp lines with the time; bind ts and caller in a single With so
	// they share one context and DefaultCaller keeps its expected stack depth.
	var l log.Logger = level.NewFilter(log.NewJSONLogger(rotator), allow)
	if config.Caller {
		l = log.With(l, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
	} else {
		l = log.With(l, "ts", log.DefaultTimestamp)
	}
	return l, nil
}

// allowOption maps Config.Level to a go-kit level filter option.
// Empty defaults to info. Aliases: warning→warn, trace→debug, fatal/panic→error.
func allowOption(name string) (level.Option, error) {
	switch strings.TrimSpace(strings.ToLower(name)) {
	case "", "info":
		return level.AllowInfo(), nil
	case "debug", "trace":
		return level.AllowDebug(), nil
	case "warn", "warning":
		return level.AllowWarn(), nil
	case "error", "fatal", "panic":
		return level.AllowError(), nil
	default:
		return nil, fmt.Errorf("unsupported log level %q, want debug|info|warn|error", name)
	}
}
