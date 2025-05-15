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
