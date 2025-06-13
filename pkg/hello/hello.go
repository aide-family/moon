package hello

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	github = "https://github.com/aide-family/moon"
	logo   = `
┌───────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                       │
│                        ███╗   ███╗ ██████╗  ██████╗ ███╗   ██╗                        │
│                        ████╗ ████║██╔═══██╗██╔═══██╗████╗  ██║                        │
│                        ██╔████╔██║██║   ██║██║   ██║██╔██╗ ██║                        │
│                        ██║╚██╔╝██║██║   ██║██║   ██║██║╚██╗██║                        │
│                        ██║ ╚═╝ ██║╚██████╔╝╚██████╔╝██║ ╚████║                        │
│                        ╚═╝     ╚═╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═══╝                        │
│                                  good luck and no bug                                 │`
	upLine   = `┌───────────────────────────────────────────────────────────────────────────────────────┐`
	downLine = `└───────────────────────────────────────────────────────────────────────────────────────┘`
)

var lineLen = utf8.RuneCountInString(upLine)

// formatLine 格式化一行内容，确保长度正确并添加右侧边框
func formatLine(prefix, content string) string {
	spaces := lineLen - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(content) - 1
	return prefix + content + strings.Repeat(" ", spaces) + "│"
}

// formatMetadataLine 格式化metadata行，确保正确对齐
func formatMetadataLine(prefix, key, value string) string {
	content := key + ": " + value
	spaces := lineLen - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(content) - 1
	return prefix + content + strings.Repeat(" ", spaces) + "│"
}

// Hello print logo and system info
func Hello() {
	fmt.Println(env.Name() + " service starting...")
	fmt.Println(logo)

	var detail strings.Builder

	// 添加基本信息行
	details := []struct {
		prefix string
		value  string
	}{
		{"├── Name:    ", env.Name()},
		{"├── Version: ", env.Version()},
		{"├── ID:      ", env.ID()},
		{"├── Env:     ", env.Env()},
		{"├── Github:  ", github},
	}

	for _, d := range details {
		detail.WriteString(formatLine(d.prefix, d.value) + "\n")
	}

	// 添加Metadata部分
	detail.WriteString(formatLine("├── Metadata: ", "") + "\n")
	meta := env.Metadata()
	lastIndex := len(meta)
	i := 1
	for k, v := range meta {
		prefix := "│   ├──"
		if i == lastIndex {
			prefix = "│   └──"
		}
		detail.WriteString(formatMetadataLine(prefix+" ", k, v) + "\n")
		i++
	}

	detail.WriteString(downLine)
	fmt.Println(detail.String())
}
