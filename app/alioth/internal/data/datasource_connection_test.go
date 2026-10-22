package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
)

// TestDatasourceConnectionMigration 保证已有部署里同地址的多个启用数据源会被停用，
// 然后唯一索引才能成功创建。
func TestDatasourceConnectionMigration(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	defer cleanup()
	if err := store.db.PingContext(ctx); err != nil {
		t.Skip(err)
	}
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	ns := "ns-ds-conn-" + time.Now().Format("150405.000")
	t.Cleanup(func() {
		_, _ = store.client.Datasource.Delete().Where(datasource.NamespaceUIDEQ(ns)).Exec(ctx)
	})

	// 模拟旧版本留下的重复连接：先去掉唯一索引，再写入两个 ACTIVE。
	if _, err := store.db.ExecContext(ctx, `DROP INDEX IF EXISTS datasource_namespace_uid_type_endpoint_index_pattern`); err != nil {
		t.Fatal(err)
	}
	for index, item := range []struct {
		name     string
		endpoint string
	}{
		{name: "prom", endpoint: "http://127.0.0.1:9090"},
		{name: "prom-old", endpoint: "http://127.0.0.1:9090/"},
	} {
		if _, err := store.client.Datasource.Create().
			SetUID(id.New()).
			SetNamespaceUID(ns).
			SetName(item.name).
			SetType(datasource.TypeMETRICS).
			SetDriver(datasource.DriverMETRICS_PROMETHEUS).
			SetEndpoint(item.endpoint).
			SetStatus(datasource.StatusACTIVE).
			Save(ctx); err != nil {
			t.Fatalf("create duplicate %d: %v", index, err)
		}
	}

	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	rows, err := store.client.Datasource.Query().Where(datasource.NamespaceUIDEQ(ns)).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	active, disabled := 0, 0
	for _, row := range rows {
		if row.Endpoint != "http://127.0.0.1:9090" {
			t.Fatalf("endpoint was not normalized: %q", row.Endpoint)
		}
		switch row.Status {
		case datasource.StatusACTIVE:
			active++
		case datasource.StatusDISABLED:
			disabled++
		}
	}
	if len(rows) != 2 || active != 1 || disabled != 1 {
		t.Fatalf("rows=%d active=%d disabled=%d", len(rows), active, disabled)
	}
	var exists bool
	if err := store.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'datasource_namespace_uid_type_endpoint_index_pattern')`,
	).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("unique connection index was not created")
	}
}
