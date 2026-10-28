package main

import (
	"log/slog"
	"testing"

	"github.com/aide-family/moon/app/alkaid/internal/conf"
	"github.com/aide-family/moon/pkg/bootstrap"
)

func TestLoadConfig(t *testing.T) {
	bc, err := bootstrap.Load[conf.Bootstrap]("../../configs/config.yaml", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	if bc.GetServer().GetHttp().GetAddr() != "0.0.0.0:8022" {
		t.Fatalf("http addr = %q", bc.GetServer().GetHttp().GetAddr())
	}
	if bc.GetServer().GetGrpc().GetAddr() != "0.0.0.0:9022" {
		t.Fatalf("grpc addr = %q", bc.GetServer().GetGrpc().GetAddr())
	}
	if bc.GetAlioth().GetGrpc().GetAddr() != "127.0.0.1:9000" {
		t.Fatalf("upstream = %q", bc.GetAlioth().GetGrpc().GetAddr())
	}
}
