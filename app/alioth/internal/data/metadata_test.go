package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

func metadataValue(metadata *bo.DatasourceMetadata, key string) string {
	for _, item := range metadata.Items {
		if item.Key == key {
			return item.Value
		}
	}
	return ""
}

func TestLokiMetadataReadsBuildInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != lokiBuildInfoPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"version":"3.2.0","revision":"abc","branch":"main","buildDate":"2026-01-01T00:00:00Z","buildUser":"builder","goVersion":"go1.24"}`))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	metadata, err := repo.DatasourceMetadata(context.Background(), bo.Datasource{
		UID: "ds-loki", Driver: driverLogsLoki, Endpoint: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Driver != driverLogsLoki || metadataValue(metadata, "version") != "3.2.0" || metadataValue(metadata, "goVersion") != "go1.24" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestElasticsearchMetadataReadsHealthAndAllocation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case esClusterHealthPath:
			_, _ = w.Write([]byte(`{"cluster_name":"alioth","status":"green","number_of_nodes":2,"number_of_data_nodes":2,"active_primary_shards":10,"active_shards":20,"relocating_shards":0,"initializing_shards":0,"unassigned_shards":0,"delayed_unassigned_shards":0,"number_of_pending_tasks":0,"active_shards_percent_as_number":100}`))
		case "/_cat/allocation":
			if got := r.URL.Query().Get("format"); got != "json" {
				t.Fatalf("format = %q", got)
			}
			_, _ = w.Write([]byte(`[{"shards":"10","disk.used":"1gb","disk.avail":"9gb","disk.percent":"10","host":"es-1","ip":"10.0.0.1","node":"es-1","node.role":"dim"}]`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	metadata, err := repo.DatasourceMetadata(context.Background(), bo.Datasource{
		UID: "ds-es", Driver: driverLogsElasticsearch, Endpoint: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if metadataValue(metadata, "clusterName") != "alioth" || metadataValue(metadata, "activeShards") != "20" {
		t.Fatalf("items = %+v", metadata.Items)
	}
	if len(metadata.Nodes) != 1 {
		t.Fatalf("nodes = %+v", metadata.Nodes)
	}
	node := metadata.Nodes[0]
	if node.Name != "es-1" || node.Address != "10.0.0.1" || node.Shards != "10" || node.DiskPercent != "10" {
		t.Fatalf("node = %+v", node)
	}
}

func TestVlogsMetadataReadsHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != vlogsHealthPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte("OK\n"))
	}))
	defer srv.Close()

	repo := newLogTestRepo()
	metadata, err := repo.DatasourceMetadata(context.Background(), bo.Datasource{
		UID: "ds-vlogs", Driver: driverLogsVlogs, Endpoint: srv.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if metadataValue(metadata, "status") != "up" || metadataValue(metadata, "health") != "OK" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestDatasourceMetadataRejectsMetricsDriver(t *testing.T) {
	repo := newLogTestRepo()
	if _, err := repo.DatasourceMetadata(context.Background(), bo.Datasource{
		UID: "ds-prom", Driver: driverMetricsPrometheus, Endpoint: "http://127.0.0.1:1",
	}); !merr.IsInvalid(err) {
		t.Fatalf("metrics metadata error = %v", err)
	}
}
