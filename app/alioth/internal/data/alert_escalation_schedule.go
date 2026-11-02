package data

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertescalationpolicy"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertescalationstep"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventescalation"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

// escalationFillLimit 是一轮补行最多处理的策略与事件数量，避免一次 tick 拖太久。
const escalationFillLimit = 500

// escalationNotifyEvent 把事件行转成分发任务需要的快照。升级任务没有订阅来源，
// ExternalURL 留空；模板里仍然可以用标签、摘要和描述。
func escalationNotifyEvent(event *ent.AlertEvent) notifyEvent {
	return notifyEvent{
		NamespaceUID: event.NamespaceUID, EventUID: event.UID, RuleUID: event.RuleUID,
		DatasourceUID: event.DatasourceUID, LevelUID: event.LevelUID, Phase: "FIRING",
		Summary: event.Summary, Description: event.Description, Labels: event.Labels,
		Value: event.SampleValue, Fingerprint: event.Fingerprint,
	}
}

// FillPendingEscalations 给 since 之后开始、仍未认领且未被抑制的 FIRING 事件补齐
// 命中策略的 PENDING 行。每条策略单独事务，某条策略失败不影响其它策略。
func (r *alertEscalationRepo) FillPendingEscalations(ctx context.Context, since, now time.Time) (int, error) {
	policies, err := r.client.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.StatusEQ(alertescalationpolicy.StatusACTIVE),
		alertescalationpolicy.DeletedAtIsNil(),
	).Limit(escalationFillLimit).All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	created := 0
	for _, row := range policies {
		steps, err := r.stepsOfPolicy(ctx, row.UID)
		if err != nil {
			return created, r.fail(err)
		}
		if len(steps) == 0 {
			continue
		}
		decoded := decodeEscalationPolicy(row, steps)
		// 策略创建之前已经触发的事件不升级：新建策略不应立即补发历史告警的
		// 已到期步骤。since 与策略创建时间取较晚者。
		eventSince := since
		if row.CreatedAt.After(eventSince) {
			eventSince = row.CreatedAt
		}
		events, err := r.client.AlertEvent.Query().Where(
			alertevent.NamespaceUIDEQ(row.NamespaceUID),
			alertevent.StatusEQ(alertevent.StatusFIRING),
			alertevent.SuppressedEQ(false),
			alertevent.AckedByEQ(""),
			alertevent.StartsAtGTE(eventSince),
			alertevent.StartsAtLTE(now),
		).Limit(escalationFillLimit).All(ctx)
		if err != nil {
			return created, r.fail(err)
		}
		matched := make([]*ent.AlertEvent, 0, len(events))
		for _, event := range events {
			if escalationMatches(decoded, event.Labels, event.LevelUID, event.DatasourceUID, event.RuleUID) {
				matched = append(matched, event)
			}
		}
		if len(matched) == 0 {
			continue
		}
		eventUIDs := make([]string, 0, len(matched))
		for _, event := range matched {
			eventUIDs = append(eventUIDs, event.UID)
		}
		existing, err := r.client.AlertEventEscalation.Query().Where(
			alerteventescalation.PolicyUIDEQ(row.UID),
			alerteventescalation.EventUIDIn(eventUIDs...),
		).All(ctx)
		if err != nil {
			return created, r.fail(err)
		}
		// 按「事件 + 步骤」判断缺失，避免某一步已存在时把整条事件跳过；
		// 步骤列表变动后新增的步骤也能补到已有事件上。
		seen := make(map[escalationRowKey]struct{}, len(existing))
		for _, item := range existing {
			seen[escalationRowKey{eventUID: item.EventUID, stepOrder: item.StepOrder}] = struct{}{}
		}
		n, err := r.fillPolicyEscalations(ctx, row.UID, row.NamespaceUID, matched, steps, seen)
		created += n
		if err != nil {
			return created, r.fail(err)
		}
	}
	return created, nil
}

// escalationRowKey 定位「事件 + 策略步骤」的一行升级记录。
type escalationRowKey struct {
	eventUID  string
	stepOrder int
}

// fillPolicyEscalations 在一条策略的事务里补齐命中事件缺失的 PENDING 行。
func (r *alertEscalationRepo) fillPolicyEscalations(ctx context.Context, policyUID, namespaceUID string, events []*ent.AlertEvent, steps []*ent.AlertEscalationStep, seen map[escalationRowKey]struct{}) (int, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	created := 0
	for _, event := range events {
		for _, step := range steps {
			if _, ok := seen[escalationRowKey{eventUID: event.UID, stepOrder: step.StepOrder}]; ok {
				continue
			}
			if _, err := tx.AlertEventEscalation.Create().
				SetUID(id.New()).
				SetNamespaceUID(namespaceUID).
				SetEventUID(event.UID).
				SetPolicyUID(policyUID).
				SetStepOrder(step.StepOrder).
				SetDueAt(event.StartsAt.Add(time.Duration(step.DelayMinutes) * time.Minute)).
				SetStatus(alerteventescalation.StatusPENDING).
				Save(ctx); err != nil {
				if ent.IsConstraintError(err) {
					continue
				}
				return created, err
			}
			created++
		}
	}
	if err := tx.Commit(); err != nil {
		return created, err
	}
	return created, nil
}

// CancelStaleEscalations 把事件已认领/恢复/抑制，或策略已停用/删除的 PENDING 行
// 置为 CANCELLED。事件恢复的常见路径也会在写事务里立即取消，这里兜住其它路径。
func (r *alertEscalationRepo) CancelStaleEscalations(ctx context.Context, now time.Time) (int, error) {
	rows, err := r.client.AlertEventEscalation.Query().Where(
		alerteventescalation.StatusEQ(alerteventescalation.StatusPENDING),
	).Order(ent.Asc(alerteventescalation.FieldDueAt)).Limit(1000).All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	if len(rows) == 0 {
		return 0, nil
	}
	eventUIDs := make([]string, 0, len(rows))
	policyUIDs := make([]string, 0, len(rows))
	seenEvent := make(map[string]struct{}, len(rows))
	seenPolicy := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if _, ok := seenEvent[row.EventUID]; !ok {
			seenEvent[row.EventUID] = struct{}{}
			eventUIDs = append(eventUIDs, row.EventUID)
		}
		if _, ok := seenPolicy[row.PolicyUID]; !ok {
			seenPolicy[row.PolicyUID] = struct{}{}
			policyUIDs = append(policyUIDs, row.PolicyUID)
		}
	}
	events, err := r.client.AlertEvent.Query().Where(alertevent.UIDIn(eventUIDs...)).All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	eventByUID := make(map[string]*ent.AlertEvent, len(events))
	for _, event := range events {
		eventByUID[event.UID] = event
	}
	policies, err := r.client.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.UIDIn(policyUIDs...),
	).All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	policyByUID := make(map[string]*ent.AlertEscalationPolicy, len(policies))
	for _, policy := range policies {
		policyByUID[policy.UID] = policy
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	cancelled := 0
	for _, row := range rows {
		if !escalationRowStale(eventByUID[row.EventUID], policyByUID[row.PolicyUID], now) {
			continue
		}
		if _, err := tx.AlertEventEscalation.Update().
			Where(alerteventescalation.IDEQ(row.ID), alerteventescalation.StatusEQ(alerteventescalation.StatusPENDING)).
			SetStatus(alerteventescalation.StatusCANCELLED).
			Save(ctx); err != nil {
			return cancelled, r.fail(err)
		}
		cancelled++
	}
	if err := tx.Commit(); err != nil {
		return cancelled, r.fail(err)
	}
	return cancelled, nil
}

// escalationRowStale 判断一条 PENDING 行是否已不再需要发送。
func escalationRowStale(event *ent.AlertEvent, policy *ent.AlertEscalationPolicy, now time.Time) bool {
	if event == nil {
		return true
	}
	if event.Status != alertevent.StatusFIRING || event.AckedBy != "" || event.Suppressed {
		return true
	}
	if event.SuppressedUntil != nil && !event.SuppressedUntil.After(now) {
		return true
	}
	if policy == nil || policy.DeletedAt != nil || policy.Status != alertescalationpolicy.StatusACTIVE {
		return true
	}
	return false
}

// cancelEventEscalations 在事件被认领或恢复的同一事务里取消未发送的升级步骤。
func cancelEventEscalations(ctx context.Context, tx *ent.Tx, namespaceUID string, eventUIDs []string) error {
	if len(eventUIDs) == 0 {
		return nil
	}
	_, err := tx.AlertEventEscalation.Update().
		Where(
			alerteventescalation.NamespaceUIDEQ(namespaceUID),
			alerteventescalation.EventUIDIn(eventUIDs...),
			alerteventescalation.StatusEQ(alerteventescalation.StatusPENDING),
		).
		SetStatus(alerteventescalation.StatusCANCELLED).
		Save(ctx)
	return err
}

// SendDueEscalations 把到期的 PENDING 行生成通知任务。同一事件、同一通知组、同一渠道
// 已有 FIRING 任务时跳过该渠道；整步没有生成任何新任务时记 SKIPPED，否则记 SENT
// 并追加 ESCALATED 时间线。一个 tick 的行数有上限，失败整体回滚下一轮重试。
func (r *alertEscalationRepo) SendDueEscalations(ctx context.Context, now time.Time) (int, error) {
	rows, err := r.client.AlertEventEscalation.Query().Where(
		alerteventescalation.StatusEQ(alerteventescalation.StatusPENDING),
		alerteventescalation.DueAtLTE(now),
	).Order(ent.Asc(alerteventescalation.FieldDueAt)).Limit(200).All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	if len(rows) == 0 {
		return 0, nil
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	sent := 0
	for _, row := range rows {
		policy, event, step, err := r.escalationRowContext(ctx, tx, row)
		if err != nil {
			return sent, r.fail(err)
		}
		if policy == nil || step == nil || escalationRowStale(event, policy, now) {
			if _, err := tx.AlertEventEscalation.Update().
				Where(alerteventescalation.IDEQ(row.ID), alerteventescalation.StatusEQ(alerteventescalation.StatusPENDING)).
				SetStatus(alerteventescalation.StatusCANCELLED).Save(ctx); err != nil {
				return sent, r.fail(err)
			}
			continue
		}
		created, _, err := r.dispatchEscalationStep(ctx, tx, event, step)
		if err != nil {
			return sent, r.fail(err)
		}
		if created > 0 {
			if _, err := tx.AlertEventEscalation.Update().
				Where(alerteventescalation.IDEQ(row.ID)).
				SetStatus(alerteventescalation.StatusSENT).
				SetSentAt(now).
				Save(ctx); err != nil {
				return sent, r.fail(err)
			}
			if err := appendEventAction(ctx, tx, row.NamespaceUID, event.UID, actionEscalated, "",
				fmt.Sprintf("policy %s step %d notified %d channel(s)", policy.Name, row.StepOrder, created)); err != nil {
				return sent, r.fail(err)
			}
			sent++
			continue
		}
		// 没有新增任务：已有同组同渠道任务时记 SKIPPED，通知组全部不可用时也记 SKIPPED。
		if _, err := tx.AlertEventEscalation.Update().
			Where(alerteventescalation.IDEQ(row.ID)).
			SetStatus(alerteventescalation.StatusSKIPPED).
			SetSentAt(now).
			Save(ctx); err != nil {
			return sent, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return sent, r.fail(err)
	}
	return sent, nil
}

// escalationRowContext 加载一条升级行对应的策略、事件与步骤。
func (r *alertEscalationRepo) escalationRowContext(ctx context.Context, tx *ent.Tx, row *ent.AlertEventEscalation) (*ent.AlertEscalationPolicy, *ent.AlertEvent, *ent.AlertEscalationStep, error) {
	policy, err := tx.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.UIDEQ(row.PolicyUID),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}
	event, err := tx.AlertEvent.Query().Where(alertevent.UIDEQ(row.EventUID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return policy, nil, nil, nil
		}
		return nil, nil, nil, err
	}
	step, err := tx.AlertEscalationStep.Query().Where(
		alertescalationstep.PolicyUIDEQ(row.PolicyUID),
		alertescalationstep.StepOrderEQ(row.StepOrder),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return policy, event, nil, nil
		}
		return nil, nil, nil, err
	}
	return policy, event, step, nil
}

// dispatchEscalationStep 把一步的通知组展开成渠道目标并生成通知任务。
// 返回新增任务数和已存在的同组同渠道任务数。
func (r *alertEscalationRepo) dispatchEscalationStep(ctx context.Context, tx *ent.Tx, event *ent.AlertEvent, step *ent.AlertEscalationStep) (int, int, error) {
	notify := escalationNotifyEvent(event)
	tracker, err := newDispatchTracker(ctx, tx, notify)
	if err != nil {
		return 0, 0, err
	}
	params := paramsJSON(notify)
	created, existing := 0, 0
	for _, groupUID := range step.NotificationGroupUids {
		group, err := tx.NotificationGroup.Query().Where(
			notificationgroup.UIDEQ(groupUID),
			notificationgroup.NamespaceUIDEQ(event.NamespaceUID),
			notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
			notificationgroup.DeletedAtIsNil(),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return created, existing, err
		}
		// 升级步骤只直接指定通知组，不经过 OnCall 绑定，因此没有值班轮换。
		targets, err := resolveGroupTargets(ctx, tx, group, "")
		if err != nil {
			return created, existing, err
		}
		for _, target := range targets {
			exists, err := eventTaskExists(ctx, tx, notify, group.UID, target.ChannelUID)
			if err != nil {
				return created, existing, err
			}
			if exists {
				existing++
				continue
			}
			receivers, ok := tracker.pendingTarget(target)
			if !ok {
				existing++
				continue
			}
			if err := tx.NotificationTask.Create().
				SetUID(id.New()).
				SetNamespaceUID(event.NamespaceUID).
				SetEventUID(event.UID).
				SetPhase(notificationtask.PhaseFIRING).
				SetStatus(notificationtask.StatusPENDING).
				SetRuleUID(event.RuleUID).
				SetFingerprint(event.Fingerprint).
				SetGroupUID(group.UID).
				SetChannelUID(target.ChannelUID).
				SetTemplateUID(target.TemplateUID).
				SetSender(target.Sender).
				SetReceivers(receivers).
				SetSubject(target.Subject).
				SetBody(target.Body).
				SetContent(target.Body).
				SetParams(params).
				SetRetriesLeft(3).
				Exec(ctx); err != nil {
				if ent.IsConstraintError(err) {
					existing++
					continue
				}
				return created, existing, err
			}
			tracker.markScheduled(target, receivers)
			created++
		}
	}
	return created, existing, nil
}
