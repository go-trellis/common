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

	"xorm.io/xorm/log"

	"github.com/sirupsen/logrus"
)

type XormLogger log.Logger

type XormLogrus struct {
	showSQL bool
	level   log.LogLevel
	Logger  *logrus.Logger
}

func ToXormLogger(l Logger) XormLogger {
	switch t := l.(type) {
	case nil:
		return discardXormLogger()
	case *logrus.Logger:
		if t == nil {
			return discardXormLogger()
		}
		return &XormLogrus{
			Logger: t,
			level:  logrusLevelToXorm(t.GetLevel()),
		}
	default:
		// InitLogger returns *logrus.Logger; other FieldLogger values cannot
		// drive SQL Infof safely, so fall back to a discard sink.
		return discardXormLogger()
	}
}

func discardXormLogger() XormLogger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return &XormLogrus{Logger: l, level: log.LOG_OFF}
}

func logrusLevelToXorm(lv logrus.Level) log.LogLevel {
	switch lv {
	case logrus.TraceLevel, logrus.DebugLevel:
		return log.LOG_DEBUG
	case logrus.InfoLevel:
		return log.LOG_INFO
	case logrus.WarnLevel:
		return log.LOG_WARNING
	case logrus.ErrorLevel:
		return log.LOG_ERR
	case logrus.FatalLevel, logrus.PanicLevel:
		return log.LOG_OFF
	default:
		return log.LOG_UNKNOWN
	}
}

func (p *XormLogrus) enabled(min log.LogLevel) bool {
	return p.level <= min
}

func (p *XormLogrus) Debug(v ...any) {
	if p.enabled(log.LOG_DEBUG) {
		p.Logger.Debug(v...)
	}
}
func (p *XormLogrus) Debugf(format string, v ...any) {
	if p.enabled(log.LOG_DEBUG) {
		p.Logger.Debugf(format, v...)
	}
}
func (p *XormLogrus) Info(v ...any) {
	if p.enabled(log.LOG_INFO) {
		p.Logger.Info(v...)
	}
}
func (p *XormLogrus) Infof(format string, v ...any) {
	if p.enabled(log.LOG_INFO) {
		p.Logger.Infof(format, v...)
	}
}
func (p *XormLogrus) Error(v ...any) {
	if p.enabled(log.LOG_ERR) {
		p.Logger.Error(v...)
	}
}
func (p *XormLogrus) Errorf(format string, v ...any) {
	if p.enabled(log.LOG_ERR) {
		p.Logger.Errorf(format, v...)
	}
}
func (p *XormLogrus) Warn(v ...any) {
	if p.enabled(log.LOG_WARNING) {
		p.Logger.Warn(v...)
	}
}
func (p *XormLogrus) Warnf(format string, v ...any) {
	if p.enabled(log.LOG_WARNING) {
		p.Logger.Warnf(format, v...)
	}
}

func (p *XormLogrus) Level() log.LogLevel {
	return p.level
}

// SetLevel only filters xorm SQL logs. Never change the shared logrus level —
// orm log_level: 4 is xorm LOG_OFF; the old code mapped that to logrus Fatal
// and silenced the whole process after the first engine was created.
func (p *XormLogrus) SetLevel(l log.LogLevel) {
	p.level = l
}

func (p *XormLogrus) ShowSQL(show ...bool) {
	if len(show) > 0 {
		p.showSQL = show[0]
		return
	}
	p.showSQL = true
}
func (p *XormLogrus) IsShowSQL() bool {
	return p.showSQL
}
