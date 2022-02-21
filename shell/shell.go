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

package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Verbose print debug
var verbose = false

func SetVerbose(v bool) {
	verbose = v
}

// RunCommand executes a shell command.
func RunCommand(name string, args ...string) error {
	printCommand(name, args...)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printCommand(name string, args ...string) {
	if !verbose {
		return
	}
	params := append([]string{name}, args...)
	_, _ = fmt.Println(strings.Join(params, " "))
}

// Output executes a shell command and returns the trimmed output
func Output(cmd string) string {
	printCommand(cmd)
	args := strings.Fields(cmd)
	out, _ := exec.Command(args[0], args[1:]...).Output()
	return strings.Trim(string(out), " \n\r")
}

// CmdOutput executes a shell command and returns the bytes
func CmdOutput(cmd string, args ...string) ([]byte, error) {
	printCommand(cmd, args...)
	return exec.Command(cmd, args...).Output()
}

// CmdRun executes a shell command with no output
func CmdRun(cmd string, args ...string) error {
	return RunCommand(cmd, args...)
}
