package hello

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	logo = `
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
func Hello(disableLogo ...bool) {
	log.Debug(Name() + " service starting...")
	if len(disableLogo) > 0 && disableLogo[0] {
		return
	}
	fmt.Println(logo)

	var detail strings.Builder

	// 添加基本信息行
	details := []struct {
		prefix string
		value  string
	}{
		{"├── Name:    ", Name()},
		{"├── Version: ", Version()},
		{"├── ID:      ", ID()},
		{"├── Env:     ", Env()},
		{"├── NodeID:  ", strconv.FormatInt(NodeID(), 10)},
	}

	for _, d := range details {
		detail.WriteString(formatLine(d.prefix, d.value) + "\n")
	}

	// 添加Metadata部分
	detail.WriteString(formatLine("├── Metadata: ", "") + "\n")
	meta := Metadata()
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
