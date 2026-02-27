package strutil

import (
	htmlpl "html/template"
	"regexp"
	"strings"
	"testing"
	txtpl "text/template"
)

// TestNewTemplateName 测试模板名称生成函数
func TestNewTemplateName(t *testing.T) {
	// 测试用例1: 验证模板名称格式
	t.Run("ValidFormat", func(t *testing.T) {
		name := newTemplateName()

		// 验证格式: YYYYMMDDHHMMSSmmm.tmpl
		matched, err := regexp.MatchString(`^\d{17}\.tmpl$`, name)
		if err != nil {
			t.Errorf("Regex compilation failed: %v", err)
		}
		if !matched {
			t.Errorf("Template name format invalid, got: %s", name)
		}

		// 验证以.tmpl结尾
		if !strings.HasSuffix(name, ".tmpl") {
			t.Errorf("Template name should end with .tmpl, got: %s", name)
		}
	})
}

// TestExecuteTextTemplate 测试文本模板执行函数
func TestExecuteTextTemplate(t *testing.T) {
	// 测试用例1: 正常模板渲染
	t.Run("NormalTemplateRendering", func(t *testing.T) {
		tmpl := "Hello {{.Name}}!"
		data := struct {
			Name string
		}{"World"}

		result, err := ExecuteTextTemplate(tmpl, data)
		if err != nil {
			t.Fatalf("ExecuteTextTemplate failed: %v", err)
		}

		expected := "Hello World!"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例2: 带自定义函数的模板渲染
	t.Run("WithCustomFunction", func(t *testing.T) {
		tmpl := "Upper: {{upper .Name}}"
		data := struct {
			Name string
		}{"hello"}

		funcMap := txtpl.FuncMap{
			"upper": strings.ToUpper,
		}

		result, err := ExecuteTextTemplate(tmpl, data, funcMap)
		if err != nil {
			t.Fatalf("ExecuteTextTemplate failed: %v", err)
		}

		expected := "Upper: HELLO"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例3: 多个函数映射合并
	t.Run("MultipleFuncMaps", func(t *testing.T) {
		tmpl := "{{upper .Name}} - {{lower .Name}}"
		data := struct {
			Name string
		}{"Hello"}

		funcMap1 := txtpl.FuncMap{
			"upper": strings.ToUpper,
		}

		funcMap2 := txtpl.FuncMap{
			"lower": strings.ToLower,
		}

		result, err := ExecuteTextTemplate(tmpl, data, funcMap1, funcMap2)
		if err != nil {
			t.Fatalf("ExecuteTextTemplate failed: %v", err)
		}

		expected := "HELLO - hello"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例4: 模板解析错误
	t.Run("TemplateParseError", func(t *testing.T) {
		tmpl := "Hello {{.Name" // 语法错误，缺少闭合括号
		data := struct {
			Name string
		}{"World"}

		_, err := ExecuteTextTemplate(tmpl, data)
		if err == nil {
			t.Error("Expected error for invalid template syntax, but got nil")
		}
	})

	// 测试用例5: 空模板
	t.Run("EmptyTemplate", func(t *testing.T) {
		tmpl := ""
		data := struct {
			Name string
		}{"World"}

		result, err := ExecuteTextTemplate(tmpl, data)
		if err != nil {
			t.Fatalf("ExecuteTextTemplate failed: %v", err)
		}

		expected := ""
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例6: 复杂数据结构
	t.Run("ComplexDataStructure", func(t *testing.T) {
		tmpl := "{{range .Items}}{{.}}, {{end}}"
		data := struct {
			Items []string
		}{
			Items: []string{"apple", "banana", "cherry"},
		}

		result, err := ExecuteTextTemplate(tmpl, data)
		if err != nil {
			t.Fatalf("ExecuteTextTemplate failed: %v", err)
		}

		expected := "apple, banana, cherry, "
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// TestExecuteHTMLTemplateFile 测试HTML模板执行函数
func TestExecuteHTMLTemplateFile(t *testing.T) {
	// 测试用例1: 正常HTML模板渲染
	t.Run("NormalHTMLTemplateRendering", func(t *testing.T) {
		tmpl := "<h1>Hello {{.Name}}!</h1>"
		data := struct {
			Name string
		}{"World"}

		result, err := ExecuteHTMLTemplateFile(tmpl, data)
		if err != nil {
			t.Fatalf("ExecuteHTMLTemplateFile failed: %v", err)
		}

		expected := "<h1>Hello World!</h1>"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例2: HTML转义测试
	t.Run("HTML escaping", func(t *testing.T) {
		tmpl := "<p>{{.Content}}</p>"
		data := struct {
			Content string
		}{"<script>alert('xss')</script>"}

		result, err := ExecuteHTMLTemplateFile(tmpl, data)
		if err != nil {
			t.Fatalf("ExecuteHTMLTemplateFile failed: %v", err)
		}

		// HTML模板应该自动转义特殊字符
		expected := "<p>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</p>"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例3: 带自定义函数的HTML模板渲染
	t.Run("WithCustomFunction", func(t *testing.T) {
		tmpl := "<p>Upper: {{upper .Name}}</p>"
		data := struct {
			Name string
		}{"hello"}

		funcMap := htmlpl.FuncMap{
			"upper": strings.ToUpper,
		}

		result, err := ExecuteHTMLTemplateFile(tmpl, data, funcMap)
		if err != nil {
			t.Fatalf("ExecuteHTMLTemplateFile failed: %v", err)
		}

		expected := "<p>Upper: HELLO</p>"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例4: 多个函数映射合并
	t.Run("MultipleFuncMaps", func(t *testing.T) {
		tmpl := "<p>{{upper .Name}} - {{lower .Name}}</p>"
		data := struct {
			Name string
		}{"Hello"}

		funcMap1 := htmlpl.FuncMap{
			"upper": strings.ToUpper,
		}

		funcMap2 := htmlpl.FuncMap{
			"lower": strings.ToLower,
		}

		result, err := ExecuteHTMLTemplateFile(tmpl, data, funcMap1, funcMap2)
		if err != nil {
			t.Fatalf("ExecuteHTMLTemplateFile failed: %v", err)
		}

		expected := "<p>HELLO - hello</p>"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// 测试用例5: HTML模板解析错误
	t.Run("TemplateParseError", func(t *testing.T) {
		tmpl := "<h1>Hello {{.Name" // 语法错误，缺少闭合括号
		data := struct {
			Name string
		}{"World"}

		_, err := ExecuteHTMLTemplateFile(tmpl, data)
		if err == nil {
			t.Error("Expected error for invalid template syntax, but got nil")
		}
	})

	// 测试用例6: 空HTML模板
	t.Run("EmptyTemplate", func(t *testing.T) {
		tmpl := ""
		data := struct {
			Name string
		}{"World"}

		result, err := ExecuteHTMLTemplateFile(tmpl, data)
		if err != nil {
			t.Fatalf("ExecuteHTMLTemplateFile failed: %v", err)
		}

		expected := ""
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// TestExecuteTextTemplateWithNilData 测试文本模板处理nil数据
func TestExecuteTextTemplateWithNilData(t *testing.T) {
	t.Run("NilData", func(t *testing.T) {
		tmpl := "Static content"

		result, err := ExecuteTextTemplate(tmpl, nil)
		if err != nil {
			t.Fatalf("ExecuteTextTemplate failed with nil data: %v", err)
		}

		expected := "Static content"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// TestExecuteHTMLTemplateWithNilData 测试HTML模板处理nil数据
func TestExecuteHTMLTemplateWithNilData(t *testing.T) {
	t.Run("NilData", func(t *testing.T) {
		tmpl := "<p>Static content</p>"

		result, err := ExecuteHTMLTemplateFile(tmpl, nil)
		if err != nil {
			t.Fatalf("ExecuteHTMLTemplateFile failed with nil data: %v", err)
		}

		expected := "<p>Static content</p>"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}
