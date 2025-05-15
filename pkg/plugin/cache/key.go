package cache

import (
	"fmt"
	"strings"
)

type K string

func (k K) Key(s ...any) string {
	ks := make([]string, 0, len(s))
	for _, a := range s {
		ks = append(ks, fmt.Sprintf("%v", a))
	}
	return strings.Join(append([]string{string(k)}, ks...), ":")
}
