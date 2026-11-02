package data

import (
	"context"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

// 事件时间线的动作编码。与 M28 文档一致，字符串保留扩展空间。
const (
	actionFiring          = "FIRING"
	actionNotified        = "NOTIFIED"
	actionAcked           = "ACKED"
	actionUnacked         = "UNACKED"
	actionIntervened      = "INTERVENED"
	actionSuppressed      = "SUPPRESSED"
	actionMaintenance     = "MAINTENANCE"
	actionEscalated       = "ESCALATED"
	actionRecovered       = "RECOVERED"
	actionRecoveredByHand = "RECOVERED_BY_MANUAL"
)

// appendEventAction 在同一事务里追加一条事件时间线。只追加、不修改；
// 时间线写入失败要让主操作一起回滚，避免「认领了但没记录」。
func appendEventAction(ctx context.Context, tx *ent.Tx, namespaceUID, eventUID, action, actorUID, detail string) error {
	if namespaceUID == "" || eventUID == "" {
		return nil
	}
	_, err := tx.AlertEventAction.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetEventUID(eventUID).
		SetAction(action).
		SetActorUID(actorUID).
		SetDetail(detail).
		Save(ctx)
	return err
}

// appendEventActions 批量追加同一动作，用于批量介入与批量恢复。
func appendEventActions(ctx context.Context, tx *ent.Tx, namespaceUID string, eventUIDs []string, action, actorUID, detail string) error {
	for _, eventUID := range eventUIDs {
		if err := appendEventAction(ctx, tx, namespaceUID, eventUID, action, actorUID, detail); err != nil {
			return err
		}
	}
	return nil
}
