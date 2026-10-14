package main

import (
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	bc, err := loadConfig("../../configs/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if bc.GetServer().GetHttp().GetAddr() != "0.0.0.0:8002" {
		t.Fatalf("http addr = %q", bc.GetServer().GetHttp().GetAddr())
	}
	body, err := os.ReadFile("../../configs/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "postgres") || strings.Contains(strings.ToLower(string(body)), "database") {
		t.Fatal("merak config contains a business database")
	}
}
