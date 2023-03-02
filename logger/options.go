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
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"trellis.tech/trellis/common.v1/flagext"
	"trellis.tech/trellis/common.v1/types"

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

var _ flagext.Parser = (*FileOptions)(nil)

// FileOptions file options
type FileOptions struct {
	Filename string `yaml:"filename" json:"filename"`
	FileExt  string `yaml:"file_ext" json:"file_ext"`

	FileBasename string `yaml:"-" json:"-"`
	FileDir      string `yaml:"-" json:"-"`

	StdPrinters types.Strings `yaml:"std_printers" json:"std_printers"`

	Separator string `yaml:"separator" json:"separator"`
	MaxLength int64  `yaml:"max_length" json:"max_length"`

	MoveFileType MoveFileType `yaml:"move_file_type" json:"move_file_type"`
	// 最大保留日志个数，如果为0则全部保留
	MaxBackups int `yaml:"max_backups" json:"max_backups"`
}

func (p *FileOptions) ParseFlags(f *flag.FlagSet) {
	p.ParseFlagsWithPrefix("", f)
}

// ParseFlagsWithPrefix adds the flags required to config this to the given FlagSet.
func (p *FileOptions) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.StringVar(&p.Filename, prefix+"log.filename", "", "logger file")
	f.StringVar(&p.FileExt, prefix+"log.file_ext", ".log", "logger file ext, default filename ext, but .log")
	f.Var(&p.StdPrinters, prefix+"log.std_printers", "std printers")
	f.StringVar(&p.Separator, prefix+"log.separator", "", "log separator")
	f.Int64Var(&p.MaxLength, prefix+"log.max_length", 0, "the max bytes of data, default: no limited")
	f.IntVar(&p.MaxBackups, prefix+"log.max_backups", 0, "the max number of logged files, default: no limited")
	moveType := f.Int(prefix+"log.move_file_type", 0, "move log file type")
	p.MoveFileType = MoveFileType(*moveType)
}

func (p *FileOptions) Check() error {
	if p == nil || p.Filename == "" {
		return errors.New("file name not exist")
	}

	p.FileDir = filepath.Dir(p.Filename)
	p.Filename = filepath.Base(p.Filename)

	if ext := filepath.Ext(p.Filename); ext != "" {
		p.FileExt = ext
		p.FileBasename = strings.TrimRight(p.Filename, ext)
	} else {
		if p.FileExt == "" {
			p.FileExt = ".log"
		}
		p.FileBasename = p.Filename
		p.Filename = fmt.Sprintf("%s%s", p.Filename, p.FileExt)
	}

	return nil
}

type Option func(*LogConfig)
type LogConfig struct {
	Level       Level       `yaml:"level" json:"level"`
	Encoding    string      `yaml:"encoding" json:"encoding"` // json | console, default console
	CallerSkip  int         `yaml:"caller_skip" json:"caller_skip"`
	Caller      bool        `yaml:"caller" json:"caller"`
	StackTrace  bool        `yaml:"stack_trace" json:"stack_trace"`
	FileOptions FileOptions `yaml:",inline" json:",inline"`

	EncoderConfig *zapcore.EncoderConfig `yaml:",inline,omitempty" json:",inline,omitempty"`
}

func (p *LogConfig) ParseFlags(f *flag.FlagSet) {
	p.ParseFlagsWithPrefix("", f)
}

// ParseFlagsWithPrefix adds the flags required to config this to the given FlagSet.
func (p *LogConfig) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	logLevel := f.Int("log.level", 0, "")
	p.Level = Level(*logLevel)

	f.StringVar(&p.Encoding, prefix+"log.encoding", "", "logger encoding, default: console")
	f.IntVar(&p.CallerSkip, prefix+"log.caller_skip", 0, "caller skip")
	f.BoolVar(&p.Caller, prefix+"log.caller", false, "open the caller")
	f.BoolVar(&p.StackTrace, prefix+"log.stack_trace", false, "log trace")
	p.FileOptions.ParseFlagsWithPrefix(prefix, f)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for LogConfig.
func (p *LogConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain LogConfig
	return unmarshal((*plain)(p))
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
