package template

import (
	"encoding/json"
	html "html/template"
	"strings"
	text "text/template"

	"github.com/go-kratos/kratos/v2/errors"
	"gopkg.in/yaml.v3"

	"github.com/aide-family/moon/pkg/util/strutil"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

// TextFormatterX is a safe version of TextFormatter that returns the original format string
// if any error occurs during formatting.
//
// Example:
//
//	data := map[string]string{"name": "John"}
//	result := TextFormatterX("Hello {{.name}}!", data)
//	// Output: "Hello John!"
func TextFormatterX(format string, data any) (s string) {
	f, err := TextFormatter(format, data)
	if err != nil {
		return format
	}
	return f
}

// HtmlFormatterX is a safe version of HtmlFormatter that returns the original format string
// if any error occurs during formatting.
//
// Example:
//
//	data := map[string]string{"content": "<b>Hello</b>"}
//	result := HtmlFormatterX("<div>{{.content}}</div>", data)
//	// Output: "<div><b>Hello</b></div>"
func HtmlFormatterX(format string, data any) (s string) {
	f, err := HtmlFormatter(format, data)
	if err != nil {
		return format
	}
	return f
}

// TextFormatter formats a string using text/template with the provided data.
// It supports all template functions defined in templateFuncMap.
//
// Example:
//
//	data := struct {
//		Name string
//		Time time.Time
//	}{
//		Name: "John",
//		Time: time.Now(),
//	}
//	result, err := TextFormatter("Hello {{.Name}}! Current time: {{now}}", data)
//	// Output: "Hello John! Current time: 2024-03-21 10:00:00"
func TextFormatter(format string, data any) (s string, err error) {
	if format == "" {
		return "", errors.New(400, "FORMAT_IS_NULL", "format is null")
	}
	if validate.IsNil(data) {
		return "", errors.New(400, "DATA_IS_NIL", "data is nil")
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

// HtmlFormatter formats a string using html/template with the provided data.
// It supports all template functions defined in templateFuncMap and provides
// HTML escaping for security.
//
// Example:
//
//	data := struct {
//		Title   string
//		Content string
//	}{
//		Title:   "Welcome",
//		Content: "<script>alert('xss')</script>",
//	}
//	result, err := HtmlFormatter(`
//		<div class="title">{{.Title}}</div>
//		<div class="content">{{.Content}}</div>
//	`, data)
//	// Output: "<div class="title">Welcome</div><div class="content">&lt;script&gt;alert('xss')&lt;/script&gt;</div>"
func HtmlFormatter(format string, data any) (s string, err error) {
	if format == "" {
		return "", errors.New(400, "FORMAT_IS_NULL", "format is null")
	}
	if validate.IsNil(data) {
		return "", errors.New(400, "DATA_IS_NIL", "data is nil")
	}

	// Create a template object and define the template string
	t, err := html.New("html/template").Funcs(templateFuncMap).Parse(format)
	if err != nil {
		return "", nil
	}
	tmpl := html.Must(t, err)
	// Execute the template and fill in the data
	resultIoWriter := new(strings.Builder)

	if err = tmpl.Execute(resultIoWriter, data); err != nil {
		return "", err
	}
	return resultIoWriter.String(), nil
}

// templateFuncMap defines the available template functions that can be used in both
// text and HTML templates. These functions can be used in template strings.
//
// Available functions:
//   - now: Returns current time
//   - hasPrefix: Checks if string starts with prefix
//   - hasSuffix: Checks if string ends with suffix
//   - contains: Checks if string contains substring
//   - trimSpace: Removes leading and trailing whitespace
//   - trimPrefix: Removes leading prefix
//   - trimSuffix: Removes trailing suffix
//   - toUpper: Converts string to uppercase
//   - toLower: Converts string to lowercase
//   - replace: Replaces all occurrences of old with new
//   - split: Splits string into slice by separator
//
// Example usage in template:
//
//	{{now}}                    // Current time
//	{{hasPrefix .Name "Mr"}}   // Check if name starts with "Mr"
//	{{toUpper .Title}}         // Convert title to uppercase
//	{{split .Tags ","}}        // Split tags by comma
//	{{mask .Phone}}            // Mask phone number
//	{{maskEmail .Email}}       // Mask email
//	{{maskBankCard .Card}}     // Mask bank card number
//	{{title .Name}}            // Convert name to title case
//	{{jsonMarshal .Data}}      // Marshal data to JSON
//	{{yamlMarshal .Data}}      // Marshal data to YAML
var templateFuncMap = map[string]any{
	"now":          timex.Now,
	"hasPrefix":    strings.HasPrefix,
	"hasSuffix":    strings.HasSuffix,
	"contains":     strings.Contains,
	"trimSpace":    strings.TrimSpace,
	"trimPrefix":   strings.TrimPrefix,
	"trimSuffix":   strings.TrimSuffix,
	"toUpper":      strings.ToUpper,
	"toLower":      strings.ToLower,
	"replace":      strings.Replace,
	"split":        strings.Split,
	"mask":         strutil.MaskString,
	"maskEmail":    strutil.MaskEmail,
	"maskPhone":    strutil.MaskPhone,
	"maskBankCard": strutil.MaskBankCard,
	"title":        strutil.Title,
	"json":         jsonMarshal,
	"yaml":         yamlMarshal,
}

func jsonMarshal(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func yamlMarshal(v any) string {
	b, _ := yaml.Marshal(v)
	return string(b)
}
