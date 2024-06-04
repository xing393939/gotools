package callstack

import "strings"

// MultiString Define a custom type that implements flag.Value interface
type MultiString []string

// String method for MultiString
func (m *MultiString) String() string {
	return strings.Join(*m, ",")
}

// Set method for MultiString
func (m *MultiString) Set(value string) error {
	*m = append(*m, value)
	return nil
}
