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
	"xorm.io/xorm/log"

	"github.com/sirupsen/logrus"
)

type XormLogger log.Logger

type XormLogrus struct {
	showSQL bool
	Logger  *logrus.Logger
}

func ToXormLogger(l *logrus.Logger) XormLogger {
	return &XormLogrus{
		Logger: l,
	}
}

func (p *XormLogrus) Debug(v ...any) {
	p.Logger.Debug(v...)
}
func (p *XormLogrus) Debugf(format string, v ...any) {
	p.Logger.Debugf(format, v...)
}
func (p *XormLogrus) Info(v ...any) {
	p.Logger.Info(v...)
}
func (p *XormLogrus) Infof(format string, v ...any) {
	p.Logger.Infof(format, v...)
}
func (p *XormLogrus) Error(v ...any) {
	p.Logger.Error(v...)
}
func (p *XormLogrus) Errorf(format string, v ...any) {
	p.Logger.Errorf(format, v...)
}
func (p *XormLogrus) Warn(v ...any) {
	p.Logger.Warn(v...)
}
func (p *XormLogrus) Warnf(format string, v ...any) {
	p.Logger.Warnf(format, v...)
}

func (p *XormLogrus) Level() log.LogLevel {
	switch p.Logger.Level {
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
func (p *XormLogrus) SetLevel(l log.LogLevel) {
	switch l {
	case log.LOG_DEBUG:
		p.Logger.SetLevel(logrus.DebugLevel)
	case log.LOG_WARNING:
		p.Logger.SetLevel(logrus.WarnLevel)
	case log.LOG_ERR:
		p.Logger.SetLevel(logrus.ErrorLevel)
	case log.LOG_OFF:
		p.Logger.SetLevel(logrus.FatalLevel)
	default:
	}
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
