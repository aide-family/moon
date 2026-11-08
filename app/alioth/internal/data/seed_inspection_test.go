package data

import (
	htmltmpl "html/template"
	"testing"
	texttmpl "text/template"
)

// TestInspectionSeedTemplatesParse 确认内置模板语法可用；渲染函数在 biz 层注入。
func TestInspectionSeedTemplatesParse(t *testing.T) {
	funcs := htmltmpl.FuncMap{
		"date": func(string, any) string { return "" },
	}
	if _, err := htmltmpl.New("report").Funcs(funcs).Parse(inspectionHTMLTemplate); err != nil {
		t.Fatalf("parse builtin html: %v", err)
	}
	if _, err := texttmpl.New("report").Funcs(texttmpl.FuncMap(funcs)).Parse(inspectionMarkdownTemplate); err != nil {
		t.Fatalf("parse builtin markdown: %v", err)
	}
}
