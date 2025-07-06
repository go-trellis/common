/*
Copyright © 2025 Henry Huang <hhh@rutcode.com>

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

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"trellis.tech/trellis/common.v2/errcode"
)

type RotateConfig struct {
	FileName      string       `yaml:"filename" json:"filename"`
	MoveFileType  MoveFileType `yaml:"move_file_type" json:"move_file_type"`
	RotationSize  int64        `yaml:"rotation_size" json:"rotation_size"`
	RotationCount uint         `yaml:"rotation_count" json:"rotation_count"`
	Caller        bool         `yaml:"caller" json:"caller"`
}

func (p *RotateConfig) ToOptions() ([]rotatelogs.Option, error) {
	var options []rotatelogs.Option
	if p.FileName == "" {
		return nil, errcode.New("filename is empty")
	}
	_, filename := filepath.Split(p.FileName)
	options = append(options, rotatelogs.WithLinkName(filename))

	moveType := MoveFileType(p.MoveFileType)
	switch moveType {
	case MoveFileTypePerMinite:
		p.FileName += ".%Y%m%d%H%M"
	case MoveFileTypeHourly:
		p.FileName += ".%Y%m%d%H"
	case MoveFileTypeDaily:
		fallthrough
	default:
		moveType = MoveFileTypeDaily
		p.FileName += ".%Y%m%d"
	}
	options = append(options, rotatelogs.WithRotationTime(moveType.Duration()))

	if p.RotationSize > 0 {
		options = append(options, rotatelogs.WithRotationSize(p.RotationSize))
	}

	if p.RotationCount > 0 {
		options = append(options, rotatelogs.WithRotationCount(p.RotationCount))
	}

	return options, nil
}

type LugrusHookConfig struct {
	*RotateConfig
	Formatter logrus.Formatter
}

type LugrusRotateConfig struct {
	Levels []Level
	*RotateConfig
}

type LugrusRotateConfigs []*LugrusRotateConfig

type LugrusLevelConfig struct {
	Levels []Level
	Writer io.Writer
}

type LugrusLevelConfigs []*LugrusLevelConfig

func NewRotateLogsWithConfig(cfg *RotateConfig) (*rotatelogs.RotateLogs, error) {
	options, err := cfg.ToOptions()
	if err != nil {
		return nil, err
	}
	return NewRotateLogs(cfg.FileName, options...)
}

func NewRotateLogs(p string, options ...rotatelogs.Option) (*rotatelogs.RotateLogs, error) {
	return rotatelogs.New(p, options...)
}

func NewLFShook(output any, formatter logrus.Formatter) logrus.Hook {
	return lfshook.NewHook(output, formatter)
}

func NewLugrusHookWithConfig(cfg *LugrusHookConfig) (logrus.Hook, error) {
	options, err := cfg.ToOptions()
	if err != nil {
		return nil, err
	}
	rotate, err := NewRotateLogs(cfg.FileName, options...)
	if err != nil {
		return nil, err
	}
	return NewLFShook(rotate, cfg.Formatter), nil
}

func NewLugrusHook(formatter logrus.Formatter, cfgs []any) (logrus.Hook, error) {
	if len(cfgs) == 0 {
		return nil, nil
	}
	mapLevelWriter := make(lfshook.WriterMap)

	for _, v := range cfgs {
		switch t := v.(type) {
		case *LugrusRotateConfig:
			rotate, err := NewRotateLogsWithConfig(t.RotateConfig)
			if err != nil {
				return nil, errcode.Newf("NewRotateLogs(%+v) failed: %+v", t.RotateConfig, err)
			}
			for _, lvl := range t.Levels {
				mapLevelWriter[logrus.Level(lvl)] = rotate
			}
		case LugrusRotateConfigs:
			for _, cfg := range t {
				rotate, err := NewRotateLogsWithConfig(cfg.RotateConfig)
				if err != nil {
					return nil, errcode.Newf("NewRotateLogs(%+v) failed: %+v", cfg.RotateConfig, err)
				}
				for _, lvl := range cfg.Levels {
					mapLevelWriter[logrus.Level(lvl)] = rotate
				}
			}
		case *LugrusLevelConfig:
			for _, lvl := range t.Levels {
				mapLevelWriter[logrus.Level(lvl)] = t.Writer
			}
		case LugrusLevelConfigs:
			for _, cfg := range t {
				for _, lvl := range cfg.Levels {
					mapLevelWriter[logrus.Level(lvl)] = cfg.Writer
				}
			}
		}
	}

	return NewLFShook(mapLevelWriter, formatter), nil
}
