// Package strutil is a utility package for strings.
package strutil

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Title(s ...string) string {
	if len(s) == 0 {
		return ""
	}
	return cases.Title(language.English).String(strings.Join(s, " "))
}

func IsEmpty(s string) bool {
	return s == "" || len(strings.TrimSpace(s)) == 0
}

func IsNotEmpty(s string) bool {
	return s != "" && len(strings.TrimSpace(s)) > 0
}

// SplitSkipEmpty splits the string s by sep and returns a slice of non-empty substrings.
// If s is empty, it returns nil.
// Empty substrings and whitespace-only substrings (after trimming) are filtered out.
// Leading and trailing whitespace of each substring is removed before checking.
func SplitSkipEmpty(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if IsNotEmpty(p) {
			result = append(result, p)
		}
	}
	return result
}
