package cache

import (
	"fmt"
	"strings"
)

// K is a key for the cache.
type K string

// NewKey returns a new key for the cache.
func NewKey(s ...any) K {
	ks := make([]string, 0, len(s))
	for _, a := range s {
		ks = append(ks, fmt.Sprintf("%v", a))
	}
	return K(strings.Join(ks, separator))
}

// Joins joins the key with the given strings.
func (k K) Joins(s ...any) K {
	ks := make([]string, 0, len(s))
	for _, a := range s {
		ks = append(ks, fmt.Sprintf("%v", a))
	}
	return K(strings.Join(append([]string{string(k)}, ks...), separator))
}

// String returns the string representation of the key.
func (k K) String() string {
	writer := strings.Builder{}
	writer.WriteString(prefix)
	writer.WriteString(string(k))
	writer.WriteString(suffix)
	return writer.String()
}

var (
	separator = ":"
	prefix    = ""
	suffix    = ""
)

// SetSeparator sets the separator for the cache.
func SetSeparator(s string) {
	separator = s
}

// SetPrefix sets the prefix for the cache.
func SetPrefix(p string) {
	prefix = p
}

// SetSuffix sets the suffix for the cache.
func SetSuffix(s string) {
	suffix = s
}
