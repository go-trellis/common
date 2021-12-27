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
	"reflect"

	"go.uber.org/zap/zapcore"
	"trellis.tech/trellis/common.v1/json"
)

type SimpleLogger interface {
	Log(kvs ...interface{}) error
}

// Logger 日志对象
type Logger interface {
	SimpleLogger

	With(kvs ...interface{}) Logger

	Debug(msg string, kvs ...interface{}) // Debug(msg string, fields ...Field)
	Debugf(msg string, kvs ...interface{})
	Info(msg string, kvs ...interface{}) // Info(msg string, fields ...Field)
	Infof(msg string, kvs ...interface{})
	Warn(msg string, kvs ...interface{}) // Warn(msg string, fields ...Field)
	Warnf(msg string, kvs ...interface{})
	Error(msg string, kvs ...interface{}) // Error(msg string, fields ...Field)
	Errorf(msg string, kvs ...interface{})
	Panic(msg string, kvs ...interface{})
	Panicf(msg string, kvs ...interface{}) // Panic(msg string, fields ...Field)
	Fatal(msg string, kvs ...interface{})
	Fatalf(msg string, kvs ...interface{}) // Fatal(msg string, fields ...Field)
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
