package flagext

import (
	"flag"
	"fmt"
)

var _ flag.Value = (*ignoredFlag)(nil)

type ignoredFlag struct {
	name string
}

func (p ignoredFlag) String() string {
	return fmt.Sprintf("ignored:" + p.name)
}

func (ignoredFlag) Set(string) error {
	return nil
}

// IgnoredFlag ignores set value, without any warning
func IgnoredFlag(f *flag.FlagSet, name, message string) {
	f.Var(ignoredFlag{name}, name, message)
}
