package hello

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	github = "https://github.com/moon-monitor/moon"
)

const logo = `
┌───────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                       │
│                        ███╗   ███╗ ██████╗  ██████╗ ███╗   ██╗                        │
│                        ████╗ ████║██╔═══██╗██╔═══██╗████╗  ██║                        │
│                        ██╔████╔██║██║   ██║██║   ██║██╔██╗ ██║                        │
│                        ██║╚██╔╝██║██║   ██║██║   ██║██║╚██╗██║                        │
│                        ██║ ╚═╝ ██║╚██████╔╝╚██████╔╝██║ ╚████║                        │
│                        ╚═╝     ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝                        │
│                                  good luck and no bug                                 │`

// Hello print logo and system info
func Hello() {
	fmt.Println(env.Name() + " service starting...")

	fmt.Println(logo)

	line := `┌───────────────────────────────────────────────────────────────────────────────────────┐`
	lineLen := utf8.RuneCount([]byte(line))
	detail := ""
	name := fmt.Sprintf("├── Name:    %s", env.Name())
	detail += name + strings.Repeat(" ", lineLen-1-utf8.RuneCount([]byte(name))) + "│"
	version := fmt.Sprintf("\n├── Version: %s", env.Version())
	detail += version + strings.Repeat(" ", lineLen-utf8.RuneCount([]byte(version))) + "│"
	id := fmt.Sprintf("\n├── ID:      %s", env.ID())
	detail += id + strings.Repeat(" ", lineLen-utf8.RuneCount([]byte(id))) + "│"
	_env := fmt.Sprintf("\n├── Env:     %s", env.Env())
	detail += _env + strings.Repeat(" ", lineLen-utf8.RuneCount([]byte(_env))) + "│"
	_github := fmt.Sprintf("\n├── Github:  %s", github)
	detail += _github + strings.Repeat(" ", lineLen-utf8.RuneCount([]byte(_github))) + "│"
	detail += `
└───────────────────────────────────────────────────────────────────────────────────────┘`

	fmt.Println(detail)
}
