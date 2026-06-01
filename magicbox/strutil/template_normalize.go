package strutil

import (
	"strings"
)

// NormalizeTextTemplate adapts Helm/Alertmanager-style templates to Go text/template syntax.
func NormalizeTextTemplate(tmpl string) string {
	tmpl = strings.TrimSpace(tmpl)
	tmpl = unwrapTemplateDefineBlock(tmpl)
	return replaceInlineIfEqExpressions(tmpl)
}

func unwrapTemplateDefineBlock(tmpl string) string {
	defineStart, bodyStart := findTemplateDefineStart(tmpl)
	if defineStart < 0 {
		return tmpl
	}
	endIdx := findMatchingTemplateEnd(bodyStart, tmpl[bodyStart:])
	if endIdx < 0 {
		return tmpl
	}
	return strings.TrimSpace(tmpl[bodyStart : bodyStart+endIdx])
}

func findTemplateDefineStart(tmpl string) (defineStart int, bodyStart int) {
	searchFrom := 0
	for {
		idx := strings.Index(tmpl[searchFrom:], "{{")
		if idx < 0 {
			return -1, -1
		}
		idx += searchFrom
		action, next := parseTemplateAction(tmpl, idx)
		if action == "" {
			searchFrom = idx + 2
			continue
		}
		fields := strings.Fields(action)
		if len(fields) >= 2 && fields[0] == "define" {
			return idx, next
		}
		searchFrom = next
	}
}

func findMatchingTemplateEnd(bodyStart int, body string) int {
	depth := 1
	searchFrom := 0
	for {
		idx := strings.Index(body[searchFrom:], "{{")
		if idx < 0 {
			return -1
		}
		idx += searchFrom
		action, next := parseTemplateAction(body, idx)
		if action == "" {
			searchFrom = idx + 2
			continue
		}
		fields := strings.Fields(action)
		if len(fields) == 0 {
			searchFrom = next
			continue
		}
		switch fields[0] {
		case "define", "if", "range", "with":
			depth++
		case "end":
			depth--
			if depth == 0 {
				return idx
			}
		}
		searchFrom = next
	}
}

func parseTemplateAction(src string, start int) (action string, next int) {
	if !strings.HasPrefix(src[start:], "{{") {
		return "", start
	}
	closeRel := strings.Index(src[start:], "}}")
	if closeRel < 0 {
		return "", start + 2
	}
	closeIdx := start + closeRel + 2
	raw := src[start+2 : closeIdx-2]
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "-")
	raw = strings.TrimSuffix(raw, "-")
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "/*") {
		return "", closeIdx
	}
	return raw, closeIdx
}

func replaceInlineIfEqExpressions(tmpl string) string {
	const prefix = "(if eq "
	var b strings.Builder
	b.Grow(len(tmpl))

	i := 0
	for {
		idx := strings.Index(tmpl[i:], prefix)
		if idx < 0 {
			b.WriteString(tmpl[i:])
			break
		}
		idx += i
		b.WriteString(tmpl[i:idx])

		rest := tmpl[idx+len(prefix):]
		arg1, rest, ok := parseTemplateExpression(rest)
		if !ok {
			b.WriteString(prefix)
			i = idx + len(prefix)
			continue
		}
		arg2, rest, ok := parseTemplateExpression(rest)
		if !ok {
			b.WriteString(prefix)
			i = idx + len(prefix)
			continue
		}
		arg3, rest, ok := parseTemplateExpression(rest)
		if !ok {
			b.WriteString(prefix)
			i = idx + len(prefix)
			continue
		}
		arg4, rest, ok := parseTemplateExpression(rest)
		if !ok {
			b.WriteString(prefix)
			i = idx + len(prefix)
			continue
		}
		rest = strings.TrimSpace(rest)
		if !strings.HasPrefix(rest, ")") {
			b.WriteString(prefix)
			i = idx + len(prefix)
			continue
		}
		rest = strings.TrimSpace(rest[1:])

		b.WriteString("(ternary ")
		b.WriteString(arg3)
		b.WriteByte(' ')
		b.WriteString(arg4)
		b.WriteString(" (eq ")
		b.WriteString(arg1)
		b.WriteByte(' ')
		b.WriteString(arg2)
		b.WriteString("))")

		i = idx + len(prefix) + (len(tmpl[idx+len(prefix):]) - len(rest))
	}
	return b.String()
}

func parseTemplateExpression(src string) (expr string, rest string, ok bool) {
	src = strings.TrimSpace(src)
	if src == "" {
		return "", "", false
	}
	switch src[0] {
	case '"':
		return parseDoubleQuotedString(src)
	case '\'':
		return parseSingleQuotedString(src)
	case '(':
		return parseBalancedParens(src)
	default:
		return parseBareTemplateExpression(src)
	}
}

func parseDoubleQuotedString(src string) (expr string, rest string, ok bool) {
	end := 1
	for end < len(src) {
		if src[end] == '\\' && end+1 < len(src) {
			end += 2
			continue
		}
		if src[end] == '"' {
			return src[:end+1], strings.TrimSpace(src[end+1:]), true
		}
		end++
	}
	return "", "", false
}

func parseSingleQuotedString(src string) (expr string, rest string, ok bool) {
	end := 1
	for end < len(src) {
		if src[end] == '\\' && end+1 < len(src) {
			end += 2
			continue
		}
		if src[end] == '\'' {
			return src[:end+1], strings.TrimSpace(src[end+1:]), true
		}
		end++
	}
	return "", "", false
}

func parseBalancedParens(src string) (expr string, rest string, ok bool) {
	depth := 0
	inDouble := false
	inSingle := false
	escaped := false
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && (inDouble || inSingle) {
			escaped = true
			continue
		}
		if inDouble {
			if ch == '"' {
				inDouble = false
			}
			continue
		}
		if inSingle {
			if ch == '\'' {
				inSingle = false
			}
			continue
		}
		switch ch {
		case '"':
			inDouble = true
		case '\'':
			inSingle = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return src[:i+1], strings.TrimSpace(src[i+1:]), true
			}
		}
	}
	return "", "", false
}

func parseBareTemplateExpression(src string) (expr string, rest string, ok bool) {
	inDouble := false
	inSingle := false
	escaped := false
	for i := 0; i < len(src); i++ {
		ch := src[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && (inDouble || inSingle) {
			escaped = true
			continue
		}
		if inDouble {
			if ch == '"' {
				inDouble = false
			}
			continue
		}
		if inSingle {
			if ch == '\'' {
				inSingle = false
			}
			continue
		}
		switch ch {
		case '"':
			inDouble = true
		case '\'':
			inSingle = true
		case ' ', '\t', '\n', '\r':
			return strings.TrimSpace(src[:i]), strings.TrimSpace(src[i:]), true
		case ')':
			return strings.TrimSpace(src[:i]), strings.TrimSpace(src[i:]), true
		}
	}
	return strings.TrimSpace(src), "", true
}
