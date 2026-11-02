package data

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertescalationpolicy"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertescalationstep"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventescalation"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	labelsutil "github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/merr"
)

// escalationMatcher 是升级策略匹配条件的落库形态，语义与维护窗口一致：
// 四个条件之间是 AND，空条件不限制；labels 是键到完整匹配正则。
type escalationMatcher struct {
	Labels         map[string]string `json:"labels"`
	LevelUIDs      []string          `json:"level_uids"`
	DatasourceUIDs []string          `json:"datasource_uids"`
	RuleUIDs       []string          `json:"rule_uids"`
}

func parseEscalationLabels(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}
	var labels map[string]string
	if err := json.Unmarshal([]byte(raw), &labels); err != nil {
		return map[string]string{}
	}
	return labels
}

func encodeEscalationMatcher(policy bo.AlertEscalationPolicy) string {
	payload, err := json.Marshal(escalationMatcher{
		Labels:         parseEscalationLabels(policy.Labels),
		LevelUIDs:      policy.LevelUIDs,
		DatasourceUIDs: policy.DatasourceUIDs,
		RuleUIDs:       policy.RuleUIDs,
	})
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func decodeEscalationPolicy(row *ent.AlertEscalationPolicy, steps []*ent.AlertEscalationStep) bo.AlertEscalationPolicy {
	item := bo.AlertEscalationPolicy{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Description: row.Description,
		Status: string(row.Status), CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	var matcher escalationMatcher
	if err := json.Unmarshal([]byte(row.Matcher), &matcher); err == nil {
		if len(matcher.Labels) > 0 {
			payload, _ := json.Marshal(matcher.Labels)
			item.Labels = string(payload)
		}
		item.LevelUIDs = matcher.LevelUIDs
		item.DatasourceUIDs = matcher.DatasourceUIDs
		item.RuleUIDs = matcher.RuleUIDs
	}
	if item.Labels == "" {
		item.Labels = "{}"
	}
	item.Steps = make([]bo.AlertEscalationStep, 0, len(steps))
	for _, step := range steps {
		item.Steps = append(item.Steps, decodeEscalationStep(step))
	}
	return item
}

func decodeEscalationStep(row *ent.AlertEscalationStep) bo.AlertEscalationStep {
	return bo.AlertEscalationStep{
		UID: row.UID, NamespaceUID: row.NamespaceUID, PolicyUID: row.PolicyUID,
		StepOrder: row.StepOrder, DelayMinutes: row.DelayMinutes,
		NotificationGroupUIDs: row.NotificationGroupUids,
	}
}

// escalationMatches 判断事件是否命中策略。条件之间是 AND，空条件不限制。
func escalationMatches(policy bo.AlertEscalationPolicy, labels, levelUID, datasourceUID, ruleUID string) bool {
	if len(policy.LevelUIDs) > 0 && !containsValue(policy.LevelUIDs, levelUID) {
		return false
	}
	if len(policy.DatasourceUIDs) > 0 && !containsValue(policy.DatasourceUIDs, datasourceUID) {
		return false
	}
	if len(policy.RuleUIDs) > 0 && !containsValue(policy.RuleUIDs, ruleUID) {
		return false
	}
	matcher := parseEscalationLabels(policy.Labels)
	if len(matcher) == 0 {
		return true
	}
	actual, err := labelsutil.ParseObject(labels)
	if err != nil {
		return false
	}
	for key, pattern := range matcher {
		value, ok := actual[key]
		if !ok {
			return false
		}
		compiled, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil || !compiled.MatchString(value) {
			return false
		}
	}
	return true
}

// stepsOfPolicy 返回一条策略的步骤，按 step_order 升序。
func (r *alertEscalationRepo) stepsOfPolicy(ctx context.Context, policyUID string) ([]*ent.AlertEscalationStep, error) {
	return r.client.AlertEscalationStep.Query().Where(
		alertescalationstep.PolicyUIDEQ(policyUID),
	).Order(ent.Asc(alertescalationstep.FieldStepOrder)).All(ctx)
}

// validateStepGroups 校验步骤里的通知组都属于当前租户、启用且未删除。
func validateStepGroups(ctx context.Context, tx *ent.Tx, namespaceUID string, steps []bo.AlertEscalationStep) error {
	uids := make([]string, 0, len(steps))
	seen := make(map[string]struct{}, len(steps))
	for _, step := range steps {
		for _, uid := range step.NotificationGroupUIDs {
			if _, ok := seen[uid]; ok {
				continue
			}
			seen[uid] = struct{}{}
			uids = append(uids, uid)
		}
	}
	if len(uids) == 0 {
		return merr.Invalid("notification group is required")
	}
	n, err := tx.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.UIDIn(uids...),
		notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
		notificationgroup.DeletedAtIsNil(),
	).Count(ctx)
	if err != nil {
		return err
	}
	if n != len(uids) {
		return merr.Invalid("notification group is unavailable")
	}
	return nil
}

// createPolicySteps 按入参顺序写入步骤，step_order 由数组下标决定。
func createPolicySteps(ctx context.Context, tx *ent.Tx, namespaceUID, policyUID string, steps []bo.AlertEscalationStep) error {
	for order, step := range steps {
		uid := step.UID
		if uid == "" {
			uid = id.New()
		}
		if _, err := tx.AlertEscalationStep.Create().
			SetUID(uid).
			SetNamespaceUID(namespaceUID).
			SetPolicyUID(policyUID).
			SetStepOrder(order).
			SetDelayMinutes(step.DelayMinutes).
			SetNotificationGroupUids(step.NotificationGroupUIDs).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *alertEscalationRepo) CreatePolicy(ctx context.Context, in bo.AlertEscalationPolicy) (*bo.AlertEscalationPolicy, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := validateStepGroups(ctx, tx, in.NamespaceUID, in.Steps); err != nil {
		return nil, err
	}
	row, err := tx.AlertEscalationPolicy.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetDescription(in.Description).
		SetMatcher(encodeEscalationMatcher(in)).
		SetStatus(alertescalationpolicy.StatusACTIVE).
		SetCreatedBy(in.CreatedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := createPolicySteps(ctx, tx, in.NamespaceUID, row.UID, in.Steps); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetPolicy(ctx, in.NamespaceUID, row.UID)
}

func (r *alertEscalationRepo) UpdatePolicy(ctx context.Context, in bo.AlertEscalationPolicy) (*bo.AlertEscalationPolicy, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.UIDEQ(in.UID),
		alertescalationpolicy.NamespaceUIDEQ(in.NamespaceUID),
		alertescalationpolicy.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("escalation policy not found")
		}
		return nil, r.fail(err)
	}
	if err := validateStepGroups(ctx, tx, in.NamespaceUID, in.Steps); err != nil {
		return nil, err
	}
	// 先清空旧步骤再按新顺序重建：保留客户端带回的 uid 作为步骤身份，
	// 避免调换顺序时 (policy_uid, step_order) 唯一约束互相冲突。
	previous, err := tx.AlertEscalationStep.Query().Where(
		alertescalationstep.PolicyUIDEQ(row.UID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	known := make(map[string]struct{}, len(previous))
	for _, step := range previous {
		known[step.UID] = struct{}{}
	}
	for _, step := range in.Steps {
		if step.UID == "" {
			continue
		}
		if _, ok := known[step.UID]; !ok {
			return nil, merr.Invalid("invalid escalation step")
		}
	}
	if _, err := tx.AlertEscalationStep.Delete().Where(
		alertescalationstep.PolicyUIDEQ(row.UID),
	).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if _, err := tx.AlertEscalationPolicy.Update().Where(alertescalationpolicy.IDEQ(row.ID)).
		SetName(in.Name).
		SetDescription(in.Description).
		SetMatcher(encodeEscalationMatcher(in)).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := createPolicySteps(ctx, tx, in.NamespaceUID, row.UID, in.Steps); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetPolicy(ctx, in.NamespaceUID, in.UID)
}

func (r *alertEscalationRepo) UpdatePolicyStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertEscalationPolicy, error) {
	if _, err := r.client.AlertEscalationPolicy.Update().Where(
		alertescalationpolicy.UIDEQ(uid),
		alertescalationpolicy.NamespaceUIDEQ(namespaceUID),
		alertescalationpolicy.DeletedAtIsNil(),
	).SetStatus(alertescalationpolicy.Status(status)).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	return r.GetPolicy(ctx, namespaceUID, uid)
}

func (r *alertEscalationRepo) DeletePolicy(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.UIDEQ(uid),
		alertescalationpolicy.NamespaceUIDEQ(namespaceUID),
		alertescalationpolicy.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return merr.NotFound("escalation policy not found")
		}
		return r.fail(err)
	}
	if _, err := tx.AlertEscalationStep.Delete().Where(alertescalationstep.PolicyUIDEQ(row.UID)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.AlertEscalationPolicy.Update().Where(alertescalationpolicy.IDEQ(row.ID)).
		SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *alertEscalationRepo) GetPolicy(ctx context.Context, namespaceUID, uid string) (*bo.AlertEscalationPolicy, error) {
	row, err := r.client.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.UIDEQ(uid),
		alertescalationpolicy.NamespaceUIDEQ(namespaceUID),
		alertescalationpolicy.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("escalation policy not found")
		}
		return nil, r.fail(err)
	}
	steps, err := r.stepsOfPolicy(ctx, row.UID)
	if err != nil {
		return nil, r.fail(err)
	}
	item := decodeEscalationPolicy(row, steps)
	return &item, nil
}

func (r *alertEscalationRepo) ListPolicies(ctx context.Context, namespaceUID string, filter bo.AlertEscalationPolicyFilter, page, pageSize int) ([]bo.AlertEscalationPolicy, int, error) {
	query := r.client.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.NamespaceUIDEQ(namespaceUID),
		alertescalationpolicy.DeletedAtIsNil(),
	)
	if filter.Name != "" {
		query = query.Where(alertescalationpolicy.NameContainsFold(filter.Name))
	}
	if filter.Status != "" {
		query = query.Where(alertescalationpolicy.StatusEQ(alertescalationpolicy.Status(filter.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(alertescalationpolicy.FieldUpdatedAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.UID)
	}
	stepsByPolicy := make(map[string][]*ent.AlertEscalationStep, len(rows))
	if len(uids) > 0 {
		steps, err := r.client.AlertEscalationStep.Query().Where(
			alertescalationstep.PolicyUIDIn(uids...),
		).Order(ent.Asc(alertescalationstep.FieldStepOrder)).All(ctx)
		if err != nil {
			return nil, 0, r.fail(err)
		}
		for _, step := range steps {
			stepsByPolicy[step.PolicyUID] = append(stepsByPolicy[step.PolicyUID], step)
		}
	}
	items := make([]bo.AlertEscalationPolicy, 0, len(rows))
	for _, row := range rows {
		items = append(items, decodeEscalationPolicy(row, stepsByPolicy[row.UID]))
	}
	return items, total, nil
}

func (r *alertEscalationRepo) ListEventEscalations(ctx context.Context, namespaceUID, eventUID string) ([]bo.AlertEventEscalation, error) {
	rows, err := r.client.AlertEventEscalation.Query().Where(
		alerteventescalation.NamespaceUIDEQ(namespaceUID),
		alerteventescalation.EventUIDEQ(eventUID),
	).Order(ent.Asc(alerteventescalation.FieldDueAt)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(rows) == 0 {
		return []bo.AlertEventEscalation{}, nil
	}
	policyUIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		policyUIDs = append(policyUIDs, row.PolicyUID)
	}
	policies, err := r.client.AlertEscalationPolicy.Query().Where(
		alertescalationpolicy.UIDIn(policyUIDs...),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	policyNames := make(map[string]string, len(policies))
	for _, policy := range policies {
		policyNames[policy.UID] = policy.Name
	}
	steps, err := r.client.AlertEscalationStep.Query().Where(
		alertescalationstep.PolicyUIDIn(policyUIDs...),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	type stepKey struct {
		policyUID string
		order     int
	}
	stepByKey := make(map[stepKey]*ent.AlertEscalationStep, len(steps))
	for _, step := range steps {
		stepByKey[stepKey{policyUID: step.PolicyUID, order: step.StepOrder}] = step
	}
	items := make([]bo.AlertEventEscalation, 0, len(rows))
	for _, row := range rows {
		item := bo.AlertEventEscalation{
			UID: row.UID, NamespaceUID: row.NamespaceUID, EventUID: row.EventUID,
			PolicyUID: row.PolicyUID, PolicyName: policyNames[row.PolicyUID],
			StepOrder: row.StepOrder, DueAt: row.DueAt, Status: string(row.Status),
			SentAt: row.SentAt, CreatedAt: row.CreatedAt,
		}
		if step, ok := stepByKey[stepKey{policyUID: row.PolicyUID, order: row.StepOrder}]; ok {
			item.DelayMinutes = step.DelayMinutes
			item.NotificationGroupCount = len(step.NotificationGroupUids)
		}
		items = append(items, item)
	}
	return items, nil
}
