package main

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/aide-family/Alioth/app/megrez/internal/conf"
	"github.com/aide-family/Alioth/pkg/bootstrap"
)

func TestLoadConfig(t *testing.T) {
	bc, err := bootstrap.Load[conf.Bootstrap]("../../configs/config.yaml", slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	if bc.GetServer().GetHttp().GetAddr() != "0.0.0.0:8001" {
		t.Fatalf("http addr = %q", bc.GetServer().GetHttp().GetAddr())
	}
	body, err := os.ReadFile("../../configs/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "postgres") || strings.Contains(strings.ToLower(string(body)), "database") {
		t.Fatal("megrez config contains a business database")
	}
}
