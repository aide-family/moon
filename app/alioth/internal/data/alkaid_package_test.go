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
	"time"

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

// 本地开发时版本号是 dev（go run 没注入 ldflags）：本机已打好的包必须还能命中，
// 不能因为拼不出版本号就卡在 package 步骤。
func TestAgentPackageSourceMatchesLocalPackageWithoutVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alkaid_dev_linux_amd64.tar.gz")
	if err := os.WriteFile(path, []byte("local"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := NewAgentPackageSource(&conf.Deploy{
		AlkaidPackagePath: filepath.Join(dir, "alkaid_<version>_linux_<arch>.tar.gz"),
		AlkaidDownloadUrl: "http://127.0.0.1:1/unused",
	}, slog.New(slog.DiscardHandler)).(*agentPackageSource)

	name, content, err := source.Fetch(context.Background(), "amd64")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	defer content.Close()
	data, _ := io.ReadAll(content)
	if string(data) != "local" || name != "alkaid_dev_linux_amd64.tar.gz" {
		t.Fatalf("name = %q, body = %q", name, data)
	}
}

// 通配符命中多个包时取最近修改的那个，避免堆了旧版本后随机装回旧包。
func TestOpenLocalPackagePicksNewestMatch(t *testing.T) {
	dir := t.TempDir()
	older := filepath.Join(dir, "alkaid_v0.0.1_linux_amd64.tar.gz")
	newer := filepath.Join(dir, "alkaid_v0.0.2_linux_amd64.tar.gz")
	for path, body := range map[string]string{older: "old", newer: "new"} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(older, old, old); err != nil {
		t.Fatal(err)
	}
	file, name, err := openLocalPackage(filepath.Join(dir, "alkaid_*_linux_amd64.tar.gz"))
	if err != nil {
		t.Fatalf("openLocalPackage: %v", err)
	}
	defer file.Close()
	if name != "alkaid_v0.0.2_linux_amd64.tar.gz" {
		t.Fatalf("name = %q, want the newest package", name)
	}
}

// 本地和下载都不成立时，报错要写清看过的路径和该跑哪条命令。
func TestAgentPackageSourceUnknownVersionExplainsFix(t *testing.T) {
	dir := t.TempDir()
	source := NewAgentPackageSource(&conf.Deploy{
		AlkaidPackagePath: filepath.Join(dir, "alkaid_<version>_linux_<arch>.tar.gz"),
		AlkaidDownloadUrl: "https://example.invalid/alkaid_<version>_<arch>.tar.gz",
	}, slog.New(slog.DiscardHandler)).(*agentPackageSource)

	_, _, err := source.Fetch(context.Background(), "amd64")
	if err == nil {
		t.Fatal("Fetch succeeded, want an actionable error")
	}
	for _, want := range []string{
		"version is unknown",
		"make package-alkaid ARCHES=amd64",
		"ALIOTH_DEPLOY_ALKAID_PACKAGE_PATH",
		"alkaid_*_linux_amd64.tar.gz",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not contain %q", err, want)
		}
	}
}
