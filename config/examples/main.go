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

package main

import (
	"fmt"

	"github.com/go-trellis/common/config"
)

/*
include: "test.yaml"
app:
  name: "my-app"
  version: "1.0.0"
database:
  host: "localhost"
  port: 5432
users:
  - name: "user1"
    age: 30
  - name: "user2"
    age: 25
  - name: "${test.name}"
    age: "${test.age}"
*/

type Config struct {
	App      App    `yaml:"app"`
	Database User   `yaml:"database"`
	Users    []User `yaml:"users"`
}
type App struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Database struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type User struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func main() {
	cfg, err := config.NewConfig("config.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)

	cnf := &Config{}
	cfg.Object(cnf)
	fmt.Println(cnf)
}
