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
	"flag"
	"fmt"
	"io"
	"os"

	"trellis.tech/trellis/common.v3/utils/flagext"
	"trellis.tech/trellis/common.v3/utils/types"

	"gopkg.in/yaml.v3"
)

//  gor main.go --trellis.name=aaa --trellis.age=1 --config.file=config.yaml --trellis.timeout=3s --trellis.strings="a" --trellis.strings="b"

const configFileOption = "config.file"

var configFile string

type Config struct {
	Name    string         `yaml:"name"`
	Age     int            `yaml:"age"`
	Timeout types.Duration `yaml:"timeout"`
	Strings types.Strings  `yaml:"strings"`

	Users []User `yaml:"users"`
}

type User struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func (c *Config) ParseFlags(f *flag.FlagSet) {
	c.ParseFlagsWithPrefix("", f)
}

func (c *Config) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.StringVar(&c.Name, prefix+"name", "", "")
	f.IntVar(&c.Age, prefix+"age", 0, "")
	f.Var(&c.Timeout, prefix+"timeout", "")
	f.Var(&c.Strings, prefix+"strings", "")
}

func main() {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&configFile, configFileOption, "", "")

	args := os.Args[1:]
	for len(args) > 0 {
		_ = fs.Parse(args)
		args = args[1:]
	}

	cnf := &Config{}

	flagext.ParseFlagsWithPrefix("trellis.", cnf)
	flagext.IgnoredFlag(flag.CommandLine, configFileOption, "Configuration file to load.")

	flag.CommandLine.Init(flag.CommandLine.Name(), flag.ContinueOnError)

	flag.CommandLine.Parse(os.Args[1:])
	buf, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(buf, cnf)
	if err != nil {
		panic(err)
	}
	fmt.Println(cnf, len(cnf.Strings))
}
