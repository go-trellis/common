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
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"xorm.io/xorm/log"
)

var (
	_ Logger     = (*ZapLogger)(nil)
	_ log.Logger = (*ZapLogger)(nil)

	DefaultEncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "log",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
)

func NewWithZapLogger(l *zap.Logger) Logger {
	if l == nil {
		return &noop{}
	}
	return &ZapLogger{logger: l}
}

type ZapLogger struct {
	options   *LogConfig
	logger    *zap.Logger
	isShowSQL bool
	writer    zapcore.WriteSyncer
}

func NewLogger(opts ...Option) (*ZapLogger, error) {
	options := &LogConfig{}

	for _, o := range opts {
		o(options)
	}

	return NewLoggerWithConfig(options)
}

func NewLoggerWithConfig(c *LogConfig) (*ZapLogger, error) {
	if c == nil || (c.FileOptions.Filename == "" && len(c.StdPrinters) == 0) {
		return &ZapLogger{logger: zap.NewNop()}, nil
	}
	zl := &ZapLogger{
		options: c,
	}

	if err := zl.initLogger(); err != nil {
		return nil, err
	}
	return zl, nil
}

func (p *ZapLogger) initLogger() error {
	if !p.options.customEncoderConfig {
		p.options.EncoderConfig = DefaultEncoderConfig
	}

	if p.options.FileOptions.Separator != "" {
		p.options.EncoderConfig.ConsoleSeparator = p.options.FileOptions.Separator
	}

	level := zap.NewAtomicLevelAt(p.options.Level.ToZapLevel())

	var encoder zapcore.Encoder
	switch p.options.Encoding {
	case "", "console":
		encoder = zapcore.NewConsoleEncoder(p.options.EncoderConfig)
	case "json":
		encoder = zapcore.NewJSONEncoder(p.options.EncoderConfig)
	default:
		return errors.New("unknown encoding")
	}

	var ws []zapcore.WriteSyncer
	for _, op := range p.options.StdPrinters {
		switch op {
		case "stderr":
			ws = append(ws, zapcore.AddSync(os.Stderr))
		case "stdout":
			ws = append(ws, zapcore.AddSync(os.Stdout))
		default:
			return errors.New("unknown std printers")
		}
	}

	if p.options.FileOptions.Filename != "" {
		w, err := NewFileLoggerWithOptions(p.options.FileOptions)
		if err != nil {
			return err
		}
		ws = append(ws, w)
	}

	p.writer = zapcore.NewMultiWriteSyncer(ws...)
	core := zapcore.NewCore(encoder, p.writer, level)

	var options []zap.Option
	if p.options.CallerSkip != 0 {
		options = append(options, zap.AddCallerSkip(p.options.CallerSkip))
	}

	if p.options.StackTrace {
		options = append(options, zap.AddStacktrace(level))
	}

	if p.options.Caller {
		options = append(options, zap.AddCaller())
	}

	p.isShowSQL = p.options.ShowXormSQL

	p.logger = zap.New(core, options...)
	return nil
}

func (p *ZapLogger) GetZapLogger() *zap.Logger {
	return p.logger
}

// With (fields ...Field)
func (p *ZapLogger) With(kvs ...interface{}) Logger {
	newZL := &ZapLogger{
		options: p.options,
	}

	lenFields := len(kvs)
	var fields []zap.Field
	for i := 0; i < lenFields; i += 2 {
		k := kvs[i]
		var v interface{} = errors.New("MISSING VALUE")
		if i+1 < lenFields {
			v = kvs[i+1]
		}
		fields = append(fields, zap.Any(toString(k), v))
	}
	newZL.logger = p.logger.With(fields...)

	return newZL
}

// Log print log with kvs
func (p *ZapLogger) Log(kvs ...interface{}) error {
	p.Info(kvs...)
	return nil
}

// Debug 打印debug信息
func (p *ZapLogger) Debug(kvs ...interface{}) {
	p.DebugM("msg", kvs...)
}

// Debugf format打印debug信息
func (p *ZapLogger) Debugf(msg string, kvs ...interface{}) {
	p.DebugM(fmt.Sprintf(msg, kvs...))
}

// DebugM 打印debug信息
func (p *ZapLogger) DebugM(msg string, kvs ...interface{}) {
	fields := p.genKVs(kvs...)
	p.logger.Debug(msg, fields...)
}

// Info 打印Info信息
func (p *ZapLogger) Info(kvs ...interface{}) {
	p.InfoM("msg", kvs...)
}

// Infof format打印info信息
func (p *ZapLogger) Infof(msg string, kvs ...interface{}) {
	p.InfoM(fmt.Sprintf(msg, kvs...))
}

// InfoM 打印info信息
func (p *ZapLogger) InfoM(msg string, kvs ...interface{}) {
	fields := p.genKVs(kvs...)
	p.logger.Info(msg, fields...)
}

// Warn 打印warn信息
func (p *ZapLogger) Warn(kvs ...interface{}) {
	p.WarnM("msg", kvs...)
}

// Warnf format打印warn信息
func (p *ZapLogger) Warnf(msg string, kvs ...interface{}) {
	p.WarnM(fmt.Sprintf(msg, kvs...))
}

// WarnM 打印warn信息
func (p *ZapLogger) WarnM(msg string, kvs ...interface{}) {
	fields := p.genKVs(kvs...)
	p.logger.Warn(msg, fields...)
}

// Error 打印error信息
func (p *ZapLogger) Error(kvs ...interface{}) {
	p.ErrorM("msg", kvs...)
}

// Errorf format打印error信息
func (p *ZapLogger) Errorf(msg string, kvs ...interface{}) {
	p.ErrorM(fmt.Sprintf(msg, kvs...))
}

// ErrorM 打印error信息
func (p *ZapLogger) ErrorM(msg string, kvs ...interface{}) {
	fields := p.genKVs(kvs...)
	p.logger.Error(msg, fields...)
}

// Panic 打印panic信息
func (p *ZapLogger) Panic(kvs ...interface{}) {
	p.PanicM("msg", kvs...)
}

// Panicf format打印panic信息
func (p *ZapLogger) Panicf(msg string, kvs ...interface{}) {
	p.PanicM(fmt.Sprintf(msg, kvs...))
}

// PanicM 打印info信息
func (p *ZapLogger) PanicM(msg string, kvs ...interface{}) {
	fields := p.genKVs(kvs...)
	p.logger.Panic(msg, fields...)
}

// Fatal 打印panic信息
func (p *ZapLogger) Fatal(kvs ...interface{}) {
	p.PanicM("msg", kvs...)
}

// Fatalf format打印panic信息
func (p *ZapLogger) Fatalf(msg string, kvs ...interface{}) {
	p.PanicM(fmt.Sprintf(msg, kvs...))
}

// FatalM 打印info信息
func (p *ZapLogger) FatalM(msg string, kvs ...interface{}) {
	fields := p.genKVs(kvs...)
	p.logger.Fatal(msg, fields...)
}

func (p *ZapLogger) genKVs(kvs ...interface{}) []zap.Field {
	lenFields := len(kvs)
	if lenFields == 0 {
		return nil
	}
	n := 4 + (lenFields+1)/2*2

	logs := make([]zap.Field, 0, n)

	for i := 0; i < lenFields; i += 2 {
		k := kvs[i]
		var v interface{} = errors.New("MISSING VALUE")
		if i+1 < lenFields {
			v = kvs[i+1]
		}
		logs = append(logs, zap.Any(toString(k), v))
	}

	return logs
}

func (p *ZapLogger) Level() log.LogLevel {
	return 0
}

func (p *ZapLogger) SetLevel(l log.LogLevel) {
	level := p.logger.Level()
	switch l {
	case log.LOG_DEBUG:
		level.Set("DEBUG")
	case log.LOG_INFO:
		level.Set("INFO")
	case log.LOG_WARNING:
		level.Set("WARN")
	case log.LOG_ERR:
		level.Set("ERROR")
	case log.LOG_OFF:
		level.Set("PANIC")
	}
}
func (p *ZapLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		p.isShowSQL = show[0]
	} else {
		p.isShowSQL = true
	}
}
func (p *ZapLogger) IsShowSQL() bool {
	return p.isShowSQL
}

func (p *ZapLogger) Writer() io.Writer {
	return p.writer
}
