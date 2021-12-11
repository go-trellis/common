package types

import (
	"flag"
	"fmt"
)

var _ flag.Value = (*Strings)(nil)

// Strings array string
type Strings []string

func (x Strings) Len() int { return len(x) }

func (x Strings) Less(i, j int) bool { return x[i] < x[j] }

func (x Strings) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

// String implements flag.Value
func (x Strings) String() string {
	return fmt.Sprintf("%s", []string(x))
}

// Set implements flag.Value
func (x *Strings) Set(s string) error {
	*x = append(*x, s)
	return nil
}
