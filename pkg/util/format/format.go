package format

import (
	"strings"
	"text/template"
	"time"

	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tidwall/gjson"
)

// ReplaceString 替换字符串中的$为.
//
//		eg: {{ $labels.instance }} 的值大于 {{ $value }} {{ .labels.instance }} 的值大于 {{ .value }}
//	 如果{{}}中间存在$符号, 则替换成.符号
//func replaceString(str string) (s string) {
//	if str == "" {
//		return ""
//	}
//
//	// 正则表达式匹配 {{ $... }} 形式的子串
//	r := regexp.MustCompile(`\{\{\s*\$(.*?)\s*\}\}`)
//
//	// 使用 ReplaceAllStringFunc 函数替换匹配到的内容
//	s = r.ReplaceAllStringFunc(str, func(match string) string {
//		// 去掉 {{ 和 }} 符号，保留内部的变量名并替换 $
//		variable := strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}")
//		return fmt.Sprintf("{{ %s }}", strings.Replace(variable, "$", ".", 1))
//	})
//
//	return s
//}

// Formatter 格式化告警文案
func Formatter(format string, data any) (s string) {
	formatStr := format
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
		log.Error("template execute error:", err)
		return format
	}
	return resultIoWriter.String()
}

// FormatterWithErr 格式化告警文案
func FormatterWithErr(format string, data any) (s string, err error) {
	formatStr := format
	if formatStr == "" {
		return "", errors.New(400, "PARAMS_ERR", "请输入模板信息")
	}

	if types.IsNil(data) {
		return "", errors.New(400, "PARAMS_ERR", "模板数据为空，请检查你的查询语句")
	}

	defer after.RecoverX()
	// 创建一个模板对象，定义模板字符串
	t, err := template.New("alert").
		Funcs(templateFuncMap()).
		Parse(formatStr)
	if err != nil {
		return
	}
	tmpl := template.Must(t, err)
	// 执行模板并填充数据
	resultIoWriter := new(strings.Builder)
	if err = tmpl.Execute(resultIoWriter, data); err != nil {
		return
	}
	return resultIoWriter.String(), nil
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"now":             time.Now,
		"hasPrefix":       strings.HasPrefix,
		"hasSuffix":       strings.HasSuffix,
		"contains":        strings.Contains,
		"trimSpace":       strings.TrimSpace,
		"trimPrefix":      strings.TrimPrefix,
		"trimSuffix":      strings.TrimSuffix,
		"toUpper":         strings.ToUpper,
		"toLower":         strings.ToLower,
		"replace":         strings.Replace,
		"split":           strings.Split,
		"getObjectByPath": GetObjectByPath,
		"getObjectByKey":  GetObjectByKey,
	}
}

// GetObjectByPath 从json字符串中获取指定key的值
func GetObjectByPath(key string, value string) any {
	if types.TextIsNull(value) {
		return nil
	}
	return gjson.Get(value, key).Value()
}

// GetObjectByKey 从map中获取指定key的值
func GetObjectByKey(key string, value map[string]any) any {
	if types.IsNil(value) {
		return nil
	}
	return value[key]
}
