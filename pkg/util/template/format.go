package template

import (
	html "html/template"
	"strings"
	text "text/template"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/timex"
)

func TextFormatterX(format string, data any) (s string) {
	f, err := TextFormatter(format, data)
	if err != nil {
		return format
	}
	return f
}

func HtmlFormatterX(format string, data any) (s string) {
	f, err := HtmlFormatter(format, data)
	if err != nil {
		return format
	}
	return f
}

func TextFormatter(format string, data any) (s string, err error) {
	if format == "" {
		return "", merr.ErrorInternalServerError("format is null")
	}
	if data == nil {
		return "", merr.ErrorInternalServerError("data is nil")
	}

	t, err := text.New("text/template").Funcs(templateFuncMap).Parse(format)
	if err != nil {
		return "", nil
	}
	tmpl := text.Must(t, err)
	resultIoWriter := new(strings.Builder)

	if err = tmpl.Execute(resultIoWriter, data); err != nil {
		return "", err
	}
	return resultIoWriter.String(), nil
}

func HtmlFormatter(format string, data any) (s string, err error) {
	if format == "" {
		return "", merr.ErrorInternalServerError("format is null")
	}
	if data == nil {
		return "", merr.ErrorInternalServerError("data is nil")
	}

	// 创建一个模板对象，定义模板字符串
	t, err := html.New("html/template").Funcs(templateFuncMap).Parse(format)
	if err != nil {
		return "", nil
	}
	tmpl := html.Must(t, err)
	// 执行模板并填充数据
	resultIoWriter := new(strings.Builder)

	if err = tmpl.Execute(resultIoWriter, data); err != nil {
		return "", err
	}
	return resultIoWriter.String(), nil
}

var templateFuncMap = map[string]any{
	"now":        timex.Now,
	"hasPrefix":  strings.HasPrefix,
	"hasSuffix":  strings.HasSuffix,
	"contains":   strings.Contains,
	"trimSpace":  strings.TrimSpace,
	"trimPrefix": strings.TrimPrefix,
	"trimSuffix": strings.TrimSuffix,
	"toUpper":    strings.ToUpper,
	"toLower":    strings.ToLower,
	"replace":    strings.Replace,
	"split":      strings.Split,
}
