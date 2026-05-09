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
	"strings"
	"testing"

	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestSetVerbose(t *testing.T) {
	SetVerbose(true)
	testutils.Assert(t, verbose, "verbose should be set to true")

	SetVerbose(false)
	testutils.Assert(t, !verbose, "verbose should be set to false")
}

func TestOutput_Simple(t *testing.T) {
	// Test with a simple command that should work on most systems
	result := Output("echo test")
	testutils.Assert(t, strings.Contains(result, "test") || result == "test", "Output should contain command output")
}

func TestCmdOutput_Simple(t *testing.T) {
	result, err := CmdOutput("echo", "test")
	testutils.Ok(t, err)
	testutils.Assert(t, len(result) > 0, "CmdOutput should return bytes")
}

func TestCmdOutput_InvalidCommand(t *testing.T) {
	_, err := CmdOutput("nonexistentcommand12345")
	testutils.NotOk(t, err, "should return error for invalid command")
}

func TestCmdRun_Simple(t *testing.T) {
	// Test with a simple command
	err := CmdRun("echo", "test")
	testutils.Ok(t, err)
}

func TestCmdRun_InvalidCommand(t *testing.T) {
	err := CmdRun("nonexistentcommand12345")
	testutils.NotOk(t, err, "should return error for invalid command")
}

func TestRunCommand_InvalidCommand(t *testing.T) {
	err := RunCommand("nonexistentcommand12345")
	testutils.NotOk(t, err, "should return error for invalid command")
}

func TestPrintCommand_Verbose(t *testing.T) {
	SetVerbose(true)
	// printCommand should print when verbose is true
	// We can't easily test stdout, but we can verify the function doesn't panic
	printCommand("test", "arg1", "arg2")
	SetVerbose(false)
}

func TestPrintCommand_NotVerbose(t *testing.T) {
	SetVerbose(false)
	// printCommand should not print when verbose is false
	printCommand("test", "arg1", "arg2")
}
