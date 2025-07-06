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
	"time"

	"github.com/robfig/cron/v3"
	"trellis.tech/trellis/common.v2/logger"
	"trellis.tech/trellis/common.v2/plugin"
	"trellis.tech/trellis/common.v2/types"
)

type T struct{}

func (p *T) Do() error {
	fmt.Println("I am test T")
	return nil
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
	l := logger.Noop()
	p, err := plugin.NewPlugins(plugin.Logger(l))
	if err != nil {
		panic(err)
	}
	p.Start()
	time.Sleep(time.Minute)
	p.Stop()
	time.Sleep(time.Second * 5)
}
