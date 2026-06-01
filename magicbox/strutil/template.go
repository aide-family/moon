package strutil

import (
	"bytes"
	htmlpl "html/template"
	txtpl "text/template"
	"time"
)

func newTemplateName() string {
	return time.Now().Format("20060102150405000") + ".tmpl"
}

func ExecuteTextTemplate(tmpl string, data any, funcs ...txtpl.FuncMap) (string, error) {
	tmpl = NormalizeTextTemplate(tmpl)
	funcMap := cloneTextTemplateFuncMap(funcs...)
	t, err := txtpl.New(newTemplateName()).Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func ExecuteHTMLTemplateFile(tmpl string, data any, funcs ...htmlpl.FuncMap) (string, error) {
	tmpl = NormalizeTextTemplate(tmpl)
	funcMap := cloneHTMLTemplateFuncMap(funcs...)
	t, err := htmlpl.New(newTemplateName()).Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
