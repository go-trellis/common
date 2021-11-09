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
	"time"

	"go.uber.org/zap/zapcore"
)

// MoveFileType move file type
type MoveFileType int

func (p *MoveFileType) getMoveFileFlag(t time.Time) int {
	switch *p {
	case MoveFileTypePerMinite:
		return t.Minute()
	case MoveFileTypeHourly:
		return t.Hour()
	case MoveFileTypeDaily:
		return t.Day()
	default:
		return 0
	}
}

// MoveFileTypes
const (
	MoveFileTypeNothing   MoveFileType = iota // 不移动
	MoveFileTypePerMinite                     // 按分钟移动
	MoveFileTypeHourly                        // 按小时移动
	MoveFileTypeDaily                         // 按天移动
)

// FileOptions file options
type FileOptions struct {
	Filename    string   `yaml:"filename"`
	StdPrinters []string `yaml:"std_printers"`

	Separator string `yaml:"separator"`
	MaxLength int64  `yaml:"max_length"`

	MoveFileType MoveFileType `yaml:"move_file_type"`
	// 最大保留日志个数，如果为0则全部保留
	MaxBackups int `yaml:"max_backups"`
}

func (p *FileOptions) Check() error {
	if p == nil || p.Filename == "" {
		return errors.New("file name not exist")
	}

	return nil
}

type Option func(*LogConfig)
type LogConfig struct {
	Level      Level  `yaml:"level"`
	Encoding   string `yaml:"encoding,omitempty"` // json | console, default console
	CallerSkip int    `yaml:"caller_skip"`
	StackTrace bool   `yaml:"stack_trace"`
	Caller     bool   `yaml:"caller"`

	EncoderConfig *zapcore.EncoderConfig `yaml:",inline,omitempty"`

	FileOptions FileOptions `yaml:",inline"`
}

// Encoding 设置移动文件的类型
func Encoding(encoding string) Option {
	return func(f *LogConfig) {
		f.Encoding = encoding
	}
}

// LogLevel 设置等级
func LogLevel(lvl Level) Option {
	return func(f *LogConfig) {
		f.Level = lvl
	}
}

// CallerSkip 设置等级
func CallerSkip(cs int) Option {
	return func(f *LogConfig) {
		f.CallerSkip = cs
	}
}

// Caller 设置等级
func Caller() Option {
	return func(f *LogConfig) {
		f.Caller = true
	}
}

// StackTrace 设置等级
func StackTrace() Option {
	return func(f *LogConfig) {
		f.StackTrace = true
	}
}

// LogFileOptions 设置等级
func LogFileOptions(fos *FileOptions) Option {
	return func(f *LogConfig) {
		f.FileOptions = *fos
	}
}

// LogFileOption 设置等级
func LogFileOption(opts ...FileOption) Option {
	return func(f *LogConfig) {
		for _, o := range opts {
			o(&f.FileOptions)
		}
	}
}

// EncoderConfig 设置等级
func EncoderConfig(encoder *zapcore.EncoderConfig) Option {
	return func(f *LogConfig) {
		f.EncoderConfig = encoder
	}
}

// FileOption 操作配置函数
type FileOption func(*FileOptions)

// OptionSeparator 设置打印分隔符
func OptionSeparator(separator string) FileOption {
	return func(f *FileOptions) {
		f.Separator = separator
	}
}

// OptionFilename 设置文件名
func OptionFilename(name string) FileOption {
	return func(f *FileOptions) {
		f.Filename = name
	}
}

// OptionMaxLength 设置最大文件大小
func OptionMaxLength(length int64) FileOption {
	return func(f *FileOptions) {
		f.MaxLength = length
	}
}

// OptionMaxBackups 文件最大数量
func OptionMaxBackups(num int) FileOption {
	return func(f *FileOptions) {
		f.MaxBackups = num
	}
}

// OptionMoveFileType 设置移动文件的类型
func OptionMoveFileType(typ MoveFileType) FileOption {
	return func(f *FileOptions) {
		f.MoveFileType = typ
	}
}

// OptionStdPrinters 设置移动文件的类型
func OptionStdPrinters(ps []string) FileOption {
	return func(f *FileOptions) {
		f.StdPrinters = ps
	}
}
