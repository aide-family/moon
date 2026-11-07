package data

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/conf"
)

func testPackageSource(localPath, downloadURL string) *agentPackageSource {
	return NewAgentPackageSource(&conf.Deploy{
		AlkaidVersion:     "v1.2.3",
		AlkaidPackagePath: localPath,
		AlkaidDownloadUrl: downloadURL,
	}, slog.New(slog.DiscardHandler)).(*agentPackageSource)
}

// TestAgentPackageSourceUsesBundledPath 确认镜像内置的 <version>/<arch> 模板路径会被展开。
func TestAgentPackageSourceUsesBundledPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alkaid_v1.2.3_linux_arm64.tar.gz")
	if err := os.WriteFile(path, []byte("bundled"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := testPackageSource(filepath.Join(dir, "alkaid_<version>_linux_<arch>.tar.gz"), "http://127.0.0.1:1/unused")

	name, content, err := source.Fetch(context.Background(), "arm64")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	defer content.Close()
	data, _ := io.ReadAll(content)
	if string(data) != "bundled" || name != "alkaid_v1.2.3_linux_arm64.tar.gz" {
		t.Fatalf("name = %q, body = %q", name, data)
	}
}

// TestAgentPackageSourceFallsBackToDownload 确认模板路径缺失时回退到远程下载。
func TestAgentPackageSourceFallsBackToDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("remote"))
	}))
	defer server.Close()
	source := testPackageSource(filepath.Join(t.TempDir(), "missing_<version>_<arch>.tar.gz"), server.URL+"/alkaid_<arch>.tar.gz")

	name, content, err := source.Fetch(context.Background(), "amd64")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	defer content.Close()
	data, _ := io.ReadAll(content)
	if string(data) != "remote" || name != "alkaid_amd64.tar.gz" {
		t.Fatalf("name = %q, body = %q", name, data)
	}
}

// TestAgentPackageSourceExplicitPathMissing 确认显式固定路径缺失时直接报错，不偷偷换远程包。
func TestAgentPackageSourceExplicitPathMissing(t *testing.T) {
	source := testPackageSource(filepath.Join(t.TempDir(), "missing.tar.gz"), "http://127.0.0.1:1/unused")
	if _, _, err := source.Fetch(context.Background(), "amd64"); err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("Fetch error = %v, want package unavailable", err)
	}
}

// TestAgentPackageSourceDownloadChecksVersion 确认没有版本时不会拼出一个不存在的下载地址。
func TestAgentPackageSourceDownloadChecksVersion(t *testing.T) {
	source := &agentPackageSource{downloadURL: "https://example.invalid/alkaid_<version>_<arch>.tar.gz"}
	if _, _, err := source.Fetch(context.Background(), "amd64"); err == nil || !strings.Contains(err.Error(), "version is unknown") {
		t.Fatalf("Fetch error = %v, want version unknown", err)
	}
}
