package format

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/aide-family/moon/pkg/util/after"
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
		return fmt.Sprintf("{{ %s }}", strings.Replace(variable, "$", ".", 1))
	})

	return s
}

// Formatter 格式化告警文案
func Formatter(format string, data any) (s string) {
	formatStr := replaceString(format)
	if formatStr == "" || data == nil {
		return ""
	}

	defer after.RecoverX()
	// 创建一个模板对象，定义模板字符串
	t, err := template.New("alert").
		Funcs(templateFuncMap()).
		Parse(formatStr)
	if err != nil {
		return format
	}
	tmpl := template.Must(t, err)
	// 执行模板并填充数据
	resultIoWriter := new(strings.Builder)

	if err = tmpl.Execute(resultIoWriter, data); err != nil {
		return format
	}
	return resultIoWriter.String()
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
