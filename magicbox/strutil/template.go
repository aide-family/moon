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
	funcMap := make(txtpl.FuncMap)
	for _, f := range funcs {
		for k, v := range f {
			funcMap[k] = v
		}
	}
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
	funcMap := make(htmlpl.FuncMap)
	for _, f := range funcs {
		for k, v := range f {
			funcMap[k] = v
		}
	}
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
