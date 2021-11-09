/*
Copyright © 2018 Henry Huang <hhh@rutcode.com>

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

package builder

import (
	"fmt"
	"strings"

	"github.com/dimiro1/banner"
	"github.com/mattn/go-colorable"
)

// 编译信息
var (
	ProgramName     string
	ProgramVersion  string
	ProgramBranch   string
	ProgramRevision string
	CompilerVersion string
	BuildTime       string
	Author          string
)

const bannerLogo = `%s*******************************************************************
*******************************************************************
***  _______   _______   _______   _       _       _    ______  ***
*** (__   __) |  ___  \ /  _____) / \     / \     / \  /  ____) ***
***    | |    | |___| | | (____   | |     | |     | |  | (____  ***
***    | |    |  ___ /  |  ____)  | |     | |     | |  \____  \ ***
***    | |    | |  \ \  | (_____  | |__/\ | |__/\ | |   ____) | ***
***    \_/    \_/   \_/ \_______) \_____/ \_____/ \_/  (______/ ***
***                                                             ***
*******************************************************************
******************** Compile Environment **************************
*** Program Name     : %s
*** Program Version  : %s
*** Program Branch   : %s
*** Program Revision : %s
*** Compiler Version : %s
*** Build Time       : %s
*** Author           : %s
*******************************************************************
******************** Running Environment **************************
*** GO ROOT            : {{ .GOROOT }}
*** Go running version : {{ .GoVersion }}
*** Go compiler        : {{ .Compiler }}
*** Go running OS      : {{ .GOOS }} {{ .GOARCH }}
*** Go CPU Numbers     : {{ .NumCPU }}
*** Startup time       : {{ .Now "2006-01-02 15:04:05 (Monday)" }}
*******************************************************************
*******************************************************************
`

type Option func(*Options)
type Options struct {
	Color   string
	OnShow  bool
	OnColor bool
}

func Color(c string) Option {
	return func(o *Options) {
		o.Color = c
	}
}

func OnShow() Option {
	return func(o *Options) {
		o.OnShow = true
	}
}

func OnColor() Option {
	return func(o *Options) {
		o.OnColor = true
	}
}

// Show 显示项目信息
func Show(opts ...Option) {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.Color == "" {
		options.Color = "{{ .AnsiColor.Default }}"
	}

	newBanner := fmt.Sprintf(bannerLogo, options.Color,
		ProgramName, ProgramVersion,
		ProgramBranch, ProgramRevision,
		CompilerVersion, BuildTime, Author)

	banner.Init(colorable.NewColorableStdout(), options.OnShow, options.OnColor, strings.NewReader(newBanner))
}

// Version 版本信息
func Version() string {
	return fmt.Sprintf("%s, version: %s (branch: %s, revision: %s)",
		ProgramName, ProgramVersion, ProgramBranch, ProgramRevision,
	)
}

// BuildInfo returns goVersion, Author and buildTime information.
func BuildInfo() string {
	return fmt.Sprintf("(go=%s, user=%s, date=%s)", CompilerVersion, Author, BuildTime)
}
