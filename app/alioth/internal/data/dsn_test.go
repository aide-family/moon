package data

import (
	"net/url"
	"os"
	"strings"
	"testing"
)

// testDSN 读取集成测试库。未设置时跳过。开发库 alioth 上跑这些测试会留下没有租户的规则，评测会去查测试里写死的地址。
func testDSN(t *testing.T) string {
	t.Helper()
	source := strings.TrimSpace(os.Getenv("ALIOTH_TEST_DSN"))
	if source == "" {
		t.Skip("set ALIOTH_TEST_DSN to a database other than the dev database")
	}
	if dsnDatabase(source) == "alioth" {
		t.Fatal("ALIOTH_TEST_DSN must not be the dev database alioth")
	}
	return source
}

func dsnDatabase(source string) string {
	parsed, err := url.Parse(source)
	if err == nil && parsed.Host != "" && parsed.Path != "" {
		return strings.Trim(parsed.Path, "/")
	}
	for _, part := range strings.Fields(source) {
		key, value, ok := strings.Cut(part, "=")
		if ok && key == "dbname" {
			return value
		}
	}
	return ""
}

func TestDSNDatabase(t *testing.T) {
	cases := []struct {
		source string
		want   string
	}{
		{source: "postgres://alioth:alioth@127.0.0.1:5432/alioth?sslmode=disable", want: "alioth"},
		{source: "postgres://alioth:alioth@127.0.0.1:5432/alioth_test?sslmode=disable", want: "alioth_test"},
		{source: "host=127.0.0.1 dbname=alioth_test user=alioth", want: "alioth_test"},
	}
	for _, item := range cases {
		if got := dsnDatabase(item.source); got != item.want {
			t.Fatalf("dsnDatabase(%q) = %q, want %q", item.source, got, item.want)
		}
	}
}
