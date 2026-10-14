package main

import "testing"

func TestLoadConfig(t *testing.T) {
	bc, err := loadConfig("../../configs/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if bc.GetServer().GetHttp().GetAddr() != "0.0.0.0:8022" {
		t.Fatalf("http addr = %q", bc.GetServer().GetHttp().GetAddr())
	}
	if bc.GetServer().GetGrpc().GetAddr() != "0.0.0.0:9022" {
		t.Fatalf("grpc addr = %q", bc.GetServer().GetGrpc().GetAddr())
	}
	if bc.GetAlioth().GetGrpc().GetAddr() != "alioth:9000" {
		t.Fatalf("upstream = %q", bc.GetAlioth().GetGrpc().GetAddr())
	}
}
