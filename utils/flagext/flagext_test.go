/*
Copyright © 2016 Henry Huang <hhh@rutcode.com>

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

package flagext

import (
	"flag"
	"testing"

	"github.com/go-trellis/common/utils/testutils"
)

type mockParser struct {
	parseFlagsCalled           bool
	parseFlagsWithPrefixCalled bool
	prefix                     string
}

func (m *mockParser) ParseFlags(f *flag.FlagSet) {
	m.parseFlagsCalled = true
}

func (m *mockParser) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	m.parseFlagsWithPrefixCalled = true
	m.prefix = prefix
}

func TestParseFlags(t *testing.T) {
	mock1 := &mockParser{}
	mock2 := &mockParser{}

	ParseFlags(mock1, mock2)

	testutils.Assert(t, mock1.parseFlagsCalled, "ParseFlags should call ParseFlags on first parser")
	testutils.Assert(t, mock2.parseFlagsCalled, "ParseFlags should call ParseFlags on second parser")
}

func TestParseFlags_Empty(t *testing.T) {
	// Should not panic
	ParseFlags()
}

func TestParseFlagsWithPrefix(t *testing.T) {
	mock1 := &mockParser{}
	mock2 := &mockParser{}

	prefix := "test."
	ParseFlagsWithPrefix(prefix, mock1, mock2)

	testutils.Assert(t, mock1.parseFlagsWithPrefixCalled, "ParseFlagsWithPrefix should call ParseFlagsWithPrefix on first parser")
	testutils.Assert(t, mock2.parseFlagsWithPrefixCalled, "ParseFlagsWithPrefix should call ParseFlagsWithPrefix on second parser")
	testutils.Equals(t, prefix, mock1.prefix, "ParseFlagsWithPrefix should pass prefix to first parser")
	testutils.Equals(t, prefix, mock2.prefix, "ParseFlagsWithPrefix should pass prefix to second parser")
}

func TestParseFlagsWithPrefix_Empty(t *testing.T) {
	// Should not panic
	ParseFlagsWithPrefix("")
}

func TestDefaultValues(t *testing.T) {
	mock1 := &mockParser{}

	// Should not panic
	DefaultValues(mock1)

	testutils.Assert(t, mock1.parseFlagsCalled, "DefaultValues should call ParseFlags")
}

func TestDefaultValues_Empty(t *testing.T) {
	// Should not panic
	DefaultValues()
}

func TestDefaultValues_MultipleParsers(t *testing.T) {
	mock1 := &mockParser{}
	mock2 := &mockParser{}

	DefaultValues(mock1, mock2)

	testutils.Assert(t, mock1.parseFlagsCalled, "DefaultValues should call ParseFlags on first parser")
	testutils.Assert(t, mock2.parseFlagsCalled, "DefaultValues should call ParseFlags on second parser")
}

// Test with real Parser implementation
type testConfig struct {
	StringValue string
	IntValue    int
}

func (c *testConfig) ParseFlags(f *flag.FlagSet) {
	c.ParseFlagsWithPrefix("", f)
}

func (c *testConfig) ParseFlagsWithPrefix(prefix string, f *flag.FlagSet) {
	f.StringVar(&c.StringValue, prefix+"string", "default", "test string flag")
	f.IntVar(&c.IntValue, prefix+"int", 10, "test int flag")
}

func TestParseFlags_RealParser(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := &testConfig{}

	cfg.ParseFlags(fs)

	err := fs.Parse([]string{"-string", "test-value", "-int", "20"})
	testutils.Ok(t, err)
	testutils.Equals(t, "test-value", cfg.StringValue, "StringValue should be set")
	testutils.Equals(t, 20, cfg.IntValue, "IntValue should be set")
}

func TestParseFlagsWithPrefix_RealParser(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := &testConfig{}

	cfg.ParseFlagsWithPrefix("prefix.", fs)

	err := fs.Parse([]string{"-prefix.string", "prefixed-value", "-prefix.int", "30"})
	testutils.Ok(t, err)
	testutils.Equals(t, "prefixed-value", cfg.StringValue, "StringValue should be set with prefix")
	testutils.Equals(t, 30, cfg.IntValue, "IntValue should be set with prefix")
}
