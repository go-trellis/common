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
	"fmt"
	"io"
	"reflect"

	"go.uber.org/zap/zapcore"
	"trellis.tech/trellis/common.v1/json"
	"xorm.io/xorm/log"
)

type SimpleLogger interface {
	Log(kvs ...interface{}) error
}

// Logger 日志对象
type Logger interface {
	SimpleLogger

	With(kvs ...interface{}) Logger
	Writer() io.Writer

	Debug(kvs ...interface{})
	DebugM(msg string, kvs ...interface{})
	Debugf(msg string, kvs ...interface{})
	Info(kvs ...interface{})
	InfoM(msg string, kvs ...interface{})
	Infof(msg string, kvs ...interface{})
	Warn(kvs ...interface{})
	WarnM(msg string, kvs ...interface{})
	Warnf(msg string, kvs ...interface{})
	Error(kvs ...interface{})
	ErrorM(msg string, kvs ...interface{})
	Errorf(msg string, kvs ...interface{})
	Panic(kvs ...interface{})
	PanicM(msg string, kvs ...interface{})
	Panicf(msg string, kvs ...interface{})
	Fatal(kvs ...interface{})
	FatalM(msg string, kvs ...interface{})
	Fatalf(msg string, kvs ...interface{})
}

type XormLogger interface {
	Level() log.LogLevel
	SetLevel(l log.LogLevel)

	ShowSQL(show ...bool)
	IsShowSQL() bool
}

// Level log level
type Level int32

// define levels
const (
	TraceLevel = Level(iota)
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel

	LevelNameUnknown = "NULL"
	LevelNameTrace   = "TRAC"
	LevelNameDebug   = "DEBU"
	LevelNameInfo    = "INFO"
	LevelNameWarn    = "WARN"
	LevelNameError   = "ERRO"
	LevelNamePanic   = "PANC"
	LevelNameFatal   = "CRIT"

	levelColorDebug = "\033[32m%s\033[0m" // grenn
	levelColorInfo  = "\033[37m%s\033[0m" // white
	levelColorWarn  = "\033[34m%s\033[0m" // blue
	levelColorError = "\033[33m%s\033[0m" // yellow
	levelColorPanic = "\033[35m%s\033[0m" // perple
	levelColorFatal = "\033[31m%s\033[0m" // red
)

// ToZapLevel  convert level into zap level
func (p *Level) ToZapLevel() zapcore.Level {
	switch *p {
	case TraceLevel, DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}

// LevelColors printer's color
var LevelColors = map[Level]string{
	TraceLevel: levelColorDebug,
	DebugLevel: levelColorDebug,
	InfoLevel:  levelColorInfo,
	WarnLevel:  levelColorWarn,
	ErrorLevel: levelColorError,
	PanicLevel: levelColorPanic,
	FatalLevel: levelColorFatal,
}

// ToLevelName convert level into string name
func ToLevelName(lvl Level) string {
	switch lvl {
	case TraceLevel:
		return LevelNameTrace
	case DebugLevel:
		return LevelNameDebug
	case InfoLevel:
		return LevelNameInfo
	case WarnLevel:
		return LevelNameWarn
	case ErrorLevel:
		return LevelNameError
	case PanicLevel:
		return LevelNamePanic
	case FatalLevel:
		return LevelNameFatal
	default:
		return LevelNameUnknown
	}
}

func toString(v interface{}) string {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Ptr, reflect.Struct, reflect.Map:
		bs, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(bs)
	case reflect.String:
		return v.(string)
	default:
		return fmt.Sprint(v)
	}
}
