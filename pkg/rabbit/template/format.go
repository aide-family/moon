package template

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// ReplaceString 替换字符串中的$为.
//
//		eg: {{ $labels.instance }} 的值大于 {{ $value }} {{ .labels.instance }} 的值大于 {{ .value }}
//	 如果{{}}中间存在$符号, 则替换成.符号
func replaceString(str string) (s string) {
	if str == "" {
		return ""
	}

	// 正则表达式匹配 {{ $... }} 形式的子串
	r := regexp.MustCompile(`\{\{\s*\$(.*?)\s*\}\}`)

	// 使用 ReplaceAllStringFunc 函数替换匹配到的内容
	s = r.ReplaceAllStringFunc(str, func(match string) string {
		// 去掉 {{ 和 }} 符号，保留内部的变量名并替换 $
		variable := strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}")
		return fmt.Sprintf("{{%s}}", strings.Replace(variable, "$", ".", -1))
	})

	return s
}

// Parser 解析模版并填充内容
func Parser(format string, in any, out io.Writer) error {
	formatStr := replaceString(format)
	if formatStr == "" || in == nil || out == nil {
		return fmt.Errorf("template, in, out cannot be empty")
	}

	// 创建一个模板对象，定义模板字符串
	t, err := template.New("alert").
		Funcs(templateFuncMap()).
		Parse(formatStr)
	if err != nil {
		return err
	}

	if err = t.Execute(out, in); err != nil {
		return err
	}
	return nil
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"now":        time.Now,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"contains":   strings.Contains,
		"TrimSpace":  strings.TrimSpace,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"toUpper":    strings.ToUpper,
		"toLower":    strings.ToLower,
		"replace":    strings.Replace,
		"split":      strings.Split,
	}
}
