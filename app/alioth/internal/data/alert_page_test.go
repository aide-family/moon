package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestAlertColumnsStore(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	store, cleanup, err := NewData(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	defer cleanup()
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}
	repo := NewAlertPageRepo(store, slog.Default())
	ns := "ns-columns-" + time.Now().Format("150405.000")
	t.Cleanup(func() {
		_, _ = store.client.AlertEventColumn.Delete().Where().Exec(ctx)
	})
	want := []bo.AlertColumn{
		{Key: "summary", Type: "FIELD"},
		{Key: "node", Type: "LABEL", Label: "节点"},
	}
	if err := repo.SetAlertColumns(ctx, ns, "user-1", "mine", want); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetAlertColumns(ctx, ns, "user-1", "mine")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1].Key != "node" || got[1].Label != "节点" {
		t.Fatalf("columns = %+v", got)
	}
	// 同一用户同一视图覆盖写入，不新增第二条。
	if err := repo.SetAlertColumns(ctx, ns, "user-1", "mine", want[:1]); err != nil {
		t.Fatal(err)
	}
	got, err = repo.GetAlertColumns(ctx, ns, "user-1", "mine")
	if err != nil || len(got) != 1 {
		t.Fatalf("overwritten columns = %+v, err = %v", got, err)
	}
	other, err := repo.GetAlertColumns(ctx, ns, "user-2", "mine")
	if err != nil || len(other) != 0 {
		t.Fatalf("other user columns = %+v, err = %v", other, err)
	}
}
