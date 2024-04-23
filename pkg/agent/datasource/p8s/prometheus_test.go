package p8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

var (
	endpoint = "https://prom-server.aide-cloud.cn/"
)

func TestPrometheusDatasource_Metadata(t *testing.T) {
	datasource := NewPrometheusDatasource(WithEndpoint(endpoint))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	metadata, err := datasource.Metadata(ctx)
	if err != nil {
		t.Fatal(err)
	}
	metadataBs, _ := json.Marshal(metadata)
	// 写到当前目录 metadata.json文件
	ioWrite, err := os.OpenFile("metadata.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer ioWrite.Close()
	fmt.Fprintln(ioWrite, string(metadataBs))
}

func TestPrometheusDatasource_series(t *testing.T) {
	datasource := NewPrometheusDatasource(WithEndpoint(endpoint)).(*PrometheusDatasource)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	series, err := datasource.series(ctx, time.Now(), "up", "aggregator_openapi_v2_regeneration_count")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("series = %v", series)
}

func TestPrometheusDatasource_metadata(t *testing.T) {
	datasource := NewPrometheusDatasource(WithEndpoint(endpoint)).(*PrometheusDatasource)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	metadata, err := datasource.metadata(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("metadata = %v", metadata)
}
