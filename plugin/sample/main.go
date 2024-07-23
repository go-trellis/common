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

	"trellis.tech/trellis/common.v2/plugin"
)

type T struct{}

func (p *T) Do() error {
	fmt.Println("I am test T")
	return nil
}

func main() {
	worker := &T{}
	plugin.RegisterPlugin("test", worker.Do, plugin.Interval(time.Second))

	//runner()
	configure()
}

func configure() {
	p, err := plugin.NewPlugins(plugin.ConfigFile("./sample.yaml"))
	if err != nil {
		panic(err)
	}
	p.Start()
	time.Sleep(time.Second * 4)
	p.Stop()
	time.Sleep(time.Second * 5)
}

func runner() {
	p, err := plugin.NewPlugins()
	if err != nil {
		panic(err)
	}
	p.Start()
	time.Sleep(time.Second * 3)
	p.Stop()
	time.Sleep(time.Second * 5)
}
