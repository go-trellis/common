/*
Copyright © 2021 Henry Huang <hhh@rutcode.com>

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

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"trellis.tech/trellis/common.v3/logger"
	"trellis.tech/trellis/common.v3/event-plugin/plugin"
	"trellis.tech/trellis/common.v3/utils/types"
	"xorm.io/xorm/log"
)

type T struct{}

func (p *T) Do() error {
	fmt.Println("I am test T")
	return nil
}

var _ logger.Logger = (*Logger)(nil)

type Logger struct{}

func (l *Logger) Log(kvs ...any) error {
	return nil
}

func (l *Logger) Debug(kvs ...any) {
	fmt.Println(append([]any{"debug"}, kvs...)...)
}

func (l *Logger) Debugf(msg string, kvs ...any) {
	fmt.Printf("debug: "+msg+" %v\n", kvs)
}

func (l *Logger) Info(kvs ...any) {
	fmt.Println(append([]any{"info"}, kvs...)...)
}

func (l *Logger) Infof(msg string, kvs ...any) {
	fmt.Printf("info: "+msg+" %v\n", kvs)
}

func (l *Logger) Warn(kvs ...any) {
	fmt.Println(append([]any{"warn"}, kvs...)...)
}

func (l *Logger) Warnf(msg string, kvs ...any) {
	fmt.Printf("warn: "+msg+" %v\n", kvs)
}

func (l *Logger) Error(kvs ...any) {
	fmt.Println(append([]any{"error"}, kvs...)...)
}

func (l *Logger) Errorf(msg string, kvs ...any) {
	fmt.Printf("error: "+msg+" %v\n", kvs)
}

func (l *Logger) Panic(kvs ...any) {
	fmt.Println(append([]any{"panic"}, kvs...)...)
	panic("panic")
}

func (l *Logger) Panicf(msg string, kvs ...any) {
	fmt.Printf("panic: "+msg+" %v\n", kvs)
	panic("panic")
}

func (l *Logger) Fatal(kvs ...any) {
	fmt.Println(append([]any{"fatal"}, kvs...)...)
	os.Exit(1)
}

func (l *Logger) Fatalf(msg string, kvs ...any) {
	fmt.Printf("fatal: "+msg+" %v\n", kvs)
	os.Exit(1)
}

func (l *Logger) With(kvs ...any) logger.Logger {
	// Simplified implementation, a more complex implementation may be needed in actual projects
	return l
}

func (l *Logger) Writer() io.Writer {
	return os.Stdout
}

func (l *Logger) Level() log.LogLevel {
	return log.LOG_INFO
}

func (l *Logger) SetLevel(level log.LogLevel) {
	// Simplified implementation, ignore setting
}

func (l *Logger) ShowSQL(show ...bool) {
	// Simplified implementation, ignore setting
}

func (l *Logger) IsShowSQL() bool {
	return false
}

func main() {
	t := &T{}
	plugin.RegisterPlugin("test", t.Do, plugin.OptionInterval(types.Duration(time.Second*10)))

	runner()
	configure()
}

func configure() {
	p, err := plugin.NewPlugins(plugin.ConfigFile("./sample.yaml"), plugin.CronOptions(cron.WithSeconds()))
	if err != nil {
		panic(err)
	}
	p.Start()
	time.Sleep(time.Minute * 5)
	p.Stop()
}

func runner() {
	l := &Logger{}
	p, err := plugin.NewPlugins(plugin.Logger(l))
	if err != nil {
		panic(err)
	}
	p.Start()
	time.Sleep(time.Minute)
	p.Stop()
	time.Sleep(time.Second * 5)
}
