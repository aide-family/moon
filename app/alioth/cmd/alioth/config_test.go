package main

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/pkg/bootstrap"
)

func TestLoadConfig(t *testing.T) {
	bc, err := bootstrap.Load[conf.Bootstrap]("../../configs/config.yaml", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	if bc.GetServer().GetHttp().GetAddr() != "0.0.0.0:8000" {
		t.Fatalf("http addr = %q", bc.GetServer().GetHttp().GetAddr())
	}
	if bc.GetServer().GetGrpc().GetAddr() != "0.0.0.0:9000" {
		t.Fatalf("grpc addr = %q", bc.GetServer().GetGrpc().GetAddr())
	}
	if bc.GetData().GetDatabase().GetSource() == "" {
		t.Fatal("database source is empty")
	}
	if bc.GetData().GetRedis().GetAddr() != "127.0.0.1:6379" {
		t.Fatalf("redis addr = %q", bc.GetData().GetRedis().GetAddr())
	}
	// 内置安装包路径的占位符不能带花括号，否则会被 ${VAR:default} 展开提前截断。
	if got := bc.GetDeploy().GetAlkaidPackagePath(); got != "/app/alkaid/alkaid_<version>_linux_<arch>.tar.gz" {
		t.Fatalf("alkaid package path = %q", got)
	}
	if got := bc.GetDeploy().GetAlkaidDownloadUrl(); !strings.Contains(got, "/alkaid_<version>_linux_<arch>.tar.gz") {
		t.Fatalf("alkaid download url = %q", got)
	}
}
