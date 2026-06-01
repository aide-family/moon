package strutil

import (
	"reflect"
	"strings"
	"time"

	htmlpl "html/template"
	txtpl "text/template"
)

// TextTemplateFuncMap returns built-in helpers for text/template rendering.
func TextTemplateFuncMap() txtpl.FuncMap {
	return txtpl.FuncMap{
		"default":    templateDefault,
		"empty":      templateEmpty,
		"ternary":    templateTernary,
		"now":        time.Now,
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
		"mask":       MaskString,
		"maskEmail":  MaskEmail,
		"maskPhone":  MaskPhone,
		"maskBankCard": MaskBankCard,
		"title":      Title,
	}
}

// HTMLTemplateFuncMap returns built-in helpers for html/template rendering.
func HTMLTemplateFuncMap() htmlpl.FuncMap {
	textFuncs := TextTemplateFuncMap()
	htmlFuncs := make(htmlpl.FuncMap, len(textFuncs))
	for name, fn := range textFuncs {
		htmlFuncs[name] = fn
	}
	return htmlFuncs
}

func cloneTextTemplateFuncMap(extra ...txtpl.FuncMap) txtpl.FuncMap {
	merged := make(txtpl.FuncMap, len(TextTemplateFuncMap()))
	for name, fn := range TextTemplateFuncMap() {
		merged[name] = fn
	}
	for _, funcs := range extra {
		for name, fn := range funcs {
			merged[name] = fn
		}
	}
	return merged
}

func cloneHTMLTemplateFuncMap(extra ...htmlpl.FuncMap) htmlpl.FuncMap {
	merged := make(htmlpl.FuncMap, len(HTMLTemplateFuncMap()))
	for name, fn := range HTMLTemplateFuncMap() {
		merged[name] = fn
	}
	for _, funcs := range extra {
		for name, fn := range funcs {
			merged[name] = fn
		}
	}
	return merged
}

// templateDefault returns the first non-empty override value, or defaultVal.
// Usage: {{ default "N/A" .Labels.alertname }}
func templateDefault(defaultVal any, override ...any) any {
	for _, value := range override {
		if !templateIsEmpty(value) {
			return value
		}
	}
	return defaultVal
}

func templateEmpty(value any) bool {
	return templateIsEmpty(value)
}

func templateTernary(trueVal, falseVal, condition any) any {
	if templateIsTruthy(condition) {
		return trueVal
	}
	return falseVal
}

func templateIsTruthy(value any) bool {
	if value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	}
	return !templateIsEmpty(value)
}

func templateIsEmpty(value any) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return v == ""
	case bool:
		return !v
	case int:
		return v == 0
	case int32:
		return v == 0
	case int64:
		return v == 0
	case float64:
		return v == 0
	case float32:
		return v == 0
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	case reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	default:
		return false
	}
}
