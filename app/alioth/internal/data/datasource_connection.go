package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/moon/app/alioth/pkg/endpoint"
	"github.com/aide-family/moon/pkg/merr"
)

// prepareDatasourceConnections 在创建唯一索引前清理历史数据：
//  1. 规范化地址，避免末尾斜杠或大小写让同一连接看起来不同；
//  2. 同租户、同类型、同地址、同索引模式只保留最早创建的启用数据源，
//     其余改为 DISABLED。这样旧的部署也能成功创建部分唯一索引。
//
// 返回被停用的数据源 UID，供迁移在事件表可用后把仍在触发的告警收口。
func prepareDatasourceConnections(ctx context.Context, db *sql.DB) ([]string, error) {
	kind, err := relationKind(ctx, db, "datasources")
	if err != nil || kind == "" {
		return nil, err
	}
	if err := normalizeStoredEndpoints(ctx, db); err != nil {
		return nil, err
	}
	return disableDuplicateConnections(ctx, db)
}

func normalizeStoredEndpoints(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `SELECT id, endpoint FROM datasources WHERE deleted_at IS NULL`)
	if err != nil {
		slog.Default().Error(fmt.Sprintf("datasource endpoint: %v", err))
		return merr.Server("request failed")
	}
	type update struct {
		id       int64
		endpoint string
	}
	updates := make([]update, 0)
	for rows.Next() {
		var id int64
		var value string
		if err := rows.Scan(&id, &value); err != nil {
			_ = rows.Close()
			return merr.Server("request failed")
		}
		normalized := endpoint.Normalize(value)
		if normalized != value {
			updates = append(updates, update{id: id, endpoint: normalized})
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		slog.Default().Error(fmt.Sprintf("datasource endpoint: %v", err))
		return merr.Server("request failed")
	}
	_ = rows.Close()
	for _, item := range updates {
		if _, err := db.ExecContext(ctx, `UPDATE datasources SET endpoint = $1 WHERE id = $2`, item.endpoint, item.id); err != nil {
			slog.Default().Error(fmt.Sprintf("datasource endpoint: %v", err))
			return merr.Server("request failed")
		}
	}
	return nil
}

func disableDuplicateConnections(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, namespace_uid, type, endpoint, index_pattern FROM datasources
		WHERE deleted_at IS NULL AND status = 'ACTIVE'
		ORDER BY namespace_uid, type, endpoint, index_pattern, created_at, id`)
	if err != nil {
		slog.Default().Error(fmt.Sprintf("datasource connection: %v", err))
		return nil, merr.Server("request failed")
	}
	seen := map[string]struct{}{}
	duplicates := make([]int64, 0)
	for rows.Next() {
		var id int64
		var namespaceUID, kind, value, indexPattern string
		if err := rows.Scan(&id, &namespaceUID, &kind, &value, &indexPattern); err != nil {
			_ = rows.Close()
			return nil, merr.Server("request failed")
		}
		key := strings.Join([]string{namespaceUID, kind, value, indexPattern}, "\x00")
		if _, ok := seen[key]; ok {
			duplicates = append(duplicates, id)
			continue
		}
		seen[key] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		slog.Default().Error(fmt.Sprintf("datasource connection: %v", err))
		return nil, merr.Server("request failed")
	}
	_ = rows.Close()
	disabled := make([]string, 0, len(duplicates))
	for _, id := range duplicates {
		var uid string
		if err := db.QueryRowContext(ctx, `UPDATE datasources SET status = 'DISABLED', updated_at = $1 WHERE id = $2 RETURNING uid`, time.Now().UTC(), id).Scan(&uid); err != nil {
			slog.Default().Error(fmt.Sprintf("datasource connection: %v", err))
			return nil, merr.Server("request failed")
		}
		disabled = append(disabled, uid)
	}
	return disabled, nil
}

// recoverDisabledDatasourceFiring 把迁移停用的数据源上仍在触发的事件改为恢复。
// 迁移阶段不补发恢复通知：这些事件来自从未被允许的重复连接，用户没有收到触发通知。
func recoverDisabledDatasourceFiring(ctx context.Context, client *ent.Client, datasourceUIDs []string) error {
	if len(datasourceUIDs) == 0 {
		return nil
	}
	rows, err := client.AlertEvent.Query().Where(
		alertevent.DatasourceUIDIn(datasourceUIDs...),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	).All(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, row := range rows {
		if _, err := client.AlertEvent.Update().Where(
			alertevent.UIDEQ(row.UID),
			alertevent.StatusEQ(alertevent.StatusFIRING),
		).SetStatus(alertevent.StatusRECOVERED).SetEndsAt(now).Save(ctx); err != nil {
			return err
		}
		if _, err := client.AlertEventFiring.Delete().Where(
			alerteventfiring.EventUIDEQ(row.UID),
		).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
