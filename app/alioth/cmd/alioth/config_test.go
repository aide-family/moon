package main

import (
	"log/slog"
	"testing"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/pkg/bootstrap"
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
}
