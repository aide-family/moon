package data

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
)

// TestPartitionRepoEnsuresUpcomingWeeks 验证分区维护任务会提前建出未来周的分区。
// 写路径不再执行 DDL，这条路径是告警事件与消息日志能持续写入的唯一保证。
func TestPartitionRepoEnsuresUpcomingWeeks(t *testing.T) {
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
	if err := store.db.PingContext(ctx); err != nil {
		t.Skip(err)
	}
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}

	repo := NewPartitionRepo(store, slog.Default())
	// 三周后的时间点必须落在已创建的分区里（维护任务预创建 4 周）。
	at := time.Now().UTC().AddDate(0, 0, 21)
	if err := repo.EnsurePartitions(ctx, at); err != nil {
		t.Fatal(err)
	}
	start, _ := weekStart(at)
	for _, table := range partitionedTables {
		name := fmt.Sprintf("%s_y%04dw%02d", table, start.Year(), isoWeek(start))
		var exists *string
		if err := store.db.QueryRowContext(ctx, `SELECT to_regclass($1)::text`, name).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists == nil || *exists == "" {
			t.Fatalf("partition %s was not created", name)
		}
	}
}
