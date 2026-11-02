package alertescalation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/lock"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	// ScheduleLockKey 是升级调度锁。15 秒 TTL 且不续期，正好对应一次 tick 的间隔。
	ScheduleLockKey    = "alioth:lock:schedule:escalation"
	scheduleLockTTL    = 15 * time.Second
	escalationLookback = 24 * time.Hour

	// maxNameRunes 与 proto 校验保持一致。
	maxNameRunes = 64
	// maxDescriptionRunes 是描述允许的字符数。
	maxDescriptionRunes = 1024
	// maxLabelsBytes 是标签匹配器 JSON 允许的字节数。
	maxLabelsBytes = 4096
	// maxSteps 是单条策略允许的步骤数，和 proto 的 repeated.max_items 一致。
	maxSteps = 20
	// maxDelayMinutes 是一步允许的最大累计分钟数（7 天）。
	maxDelayMinutes = 10080
	// maxStepGroups 是一步允许的通知组数量。
	maxStepGroups = 20
)

// Gateway 管理升级策略，并在锁内推进升级调度。
type Gateway struct {
	policies repository.AlertEscalationRepo
	access   access.Checker
	locks    *lock.Client
	log      *slog.Logger
}

func NewGateway(policies repository.AlertEscalationRepo, checker access.Checker, locks *lock.Client, logger *slog.Logger) *Gateway {
	return &Gateway{policies: policies, access: checker, locks: locks, log: logger}
}

// Create 校验并保存升级策略，步骤随策略一起写入。
func (gateway *Gateway) Create(ctx context.Context, in bo.AlertEscalationPolicy) (*bo.AlertEscalationPolicy, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertEscalationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.CreatedBy, _ = actor.UserID(ctx)
	if err := normalizePolicy(&in); err != nil {
		return nil, err
	}
	item, err := gateway.policies.CreatePolicy(ctx, in)
	return item, gateway.access.Wrap(err)
}

// Update 校验并整体替换策略的步骤。
func (gateway *Gateway) Update(ctx context.Context, in bo.AlertEscalationPolicy) (*bo.AlertEscalationPolicy, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertEscalationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizePolicy(&in); err != nil {
		return nil, err
	}
	item, err := gateway.policies.UpdatePolicy(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateStatus 启用或停用升级策略。停用后下一轮 tick 不再补齐新事件，
// 已存在的 PENDING 行由调度器取消。
func (gateway *Gateway) UpdateStatus(ctx context.Context, uid, status string) (*bo.AlertEscalationPolicy, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertEscalationWrite)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" && status != "DISABLED" {
		return nil, merr.Invalid("invalid status")
	}
	item, err := gateway.policies.UpdatePolicyStatus(ctx, ns, uid, status)
	return item, gateway.access.Wrap(err)
}

// Delete 软删除升级策略；历史升级记录保留，未发送的步骤由下一轮 tick 取消。
func (gateway *Gateway) Delete(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.AlertEscalationDelete)
	if err != nil {
		return err
	}
	return gateway.access.Wrap(gateway.policies.DeletePolicy(ctx, ns, uid))
}

// Get 读取升级策略详情。
func (gateway *Gateway) Get(ctx context.Context, uid string) (*bo.AlertEscalationPolicy, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertEscalationRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.policies.GetPolicy(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// List 分页列出升级策略。
func (gateway *Gateway) List(ctx context.Context, pageNo, pageSize int32, name, status string) ([]bo.AlertEscalationPolicy, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertEscalationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	status = strings.TrimSpace(status)
	if status != "" && status != "ACTIVE" && status != "DISABLED" {
		return nil, 0, 0, 0, merr.Invalid("invalid status")
	}
	items, total, err := gateway.policies.ListPolicies(ctx, ns, bo.AlertEscalationPolicyFilter{
		Name: strings.TrimSpace(name), Status: status,
	}, p, s)
	if err != nil {
		return nil, 0, 0, 0, gateway.access.Wrap(err)
	}
	return items, total, p, s, nil
}

// ListEventEscalations 返回一条事件的升级记录；实时告警读权限即可查看。
func (gateway *Gateway) ListEventEscalations(ctx context.Context, eventUID string) ([]bo.AlertEventEscalation, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertRealtimeRead)
	if err != nil {
		return nil, err
	}
	items, err := gateway.policies.ListEventEscalations(ctx, ns, eventUID)
	return items, gateway.access.Wrap(err)
}

// Schedule 在锁内推进升级：取消失效步骤、补齐命中事件、发送到期步骤。
// 拿不到锁说明其它副本正在推进，本轮直接跳过；锁不续期，最长占用 15 秒。
func (gateway *Gateway) Schedule(ctx context.Context) error {
	if gateway.locks == nil {
		return merr.Server("lock failed")
	}
	hold, err := gateway.locks.Acquire(ctx, ScheduleLockKey, scheduleLockTTL)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	now := time.Now().UTC()
	if _, err := gateway.policies.CancelStaleEscalations(ctx, now); err != nil {
		gateway.log.Error(fmt.Sprintf("cancel escalation: %v", err))
		return err
	}
	if _, err := gateway.policies.FillPendingEscalations(ctx, now.Add(-escalationLookback), now); err != nil {
		gateway.log.Error(fmt.Sprintf("fill escalation: %v", err))
		return err
	}
	if _, err := gateway.policies.SendDueEscalations(ctx, now); err != nil {
		gateway.log.Error(fmt.Sprintf("send escalation: %v", err))
		return err
	}
	return nil
}

// normalizePolicy 校验名称、描述、匹配条件与步骤，并去重匹配值和通知组。
func normalizePolicy(in *bo.AlertEscalationPolicy) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || utf8.RuneCountInString(in.Name) > maxNameRunes {
		return merr.Invalid("invalid escalation policy name")
	}
	in.Description = strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(in.Description) > maxDescriptionRunes {
		return merr.Invalid("description is too long")
	}
	labels, err := normalizeLabels(in.Labels)
	if err != nil {
		return err
	}
	in.Labels = labels
	if len(in.Labels) > maxLabelsBytes {
		return merr.Invalid("labels are too long")
	}
	in.LevelUIDs = uniqueUIDs(in.LevelUIDs)
	in.DatasourceUIDs = uniqueUIDs(in.DatasourceUIDs)
	in.RuleUIDs = uniqueUIDs(in.RuleUIDs)
	if len(in.Steps) == 0 || len(in.Steps) > maxSteps {
		return merr.Invalid("invalid escalation steps")
	}
	previousDelay := -1
	for index := range in.Steps {
		step := &in.Steps[index]
		if step.DelayMinutes < 0 || step.DelayMinutes > maxDelayMinutes {
			return merr.Invalid("invalid escalation delay")
		}
		if step.DelayMinutes <= previousDelay {
			return merr.Invalid("escalation delays must increase")
		}
		previousDelay = step.DelayMinutes
		step.NotificationGroupUIDs = uniqueUIDs(step.NotificationGroupUIDs)
		if len(step.NotificationGroupUIDs) == 0 || len(step.NotificationGroupUIDs) > maxStepGroups {
			return merr.Invalid("notification groups are required")
		}
	}
	if in.Steps[len(in.Steps)-1].DelayMinutes <= 0 {
		return merr.Invalid("the last escalation step must be delayed")
	}
	return nil
}

// normalizeLabels 把标签匹配器收紧成 JSON 对象：键非空、值能编译成正则。
func normalizeLabels(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}", nil
	}
	var labels map[string]string
	if err := json.Unmarshal([]byte(raw), &labels); err != nil {
		return "", merr.Invalid("invalid labels")
	}
	out := make(map[string]string, len(labels))
	for key, value := range labels {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return "", merr.Invalid("invalid labels")
		}
		if _, err := regexp.Compile("^(?:" + value + ")$"); err != nil {
			return "", merr.Invalid("invalid labels")
		}
		out[key] = value
	}
	payload, err := json.Marshal(out)
	if err != nil {
		return "", merr.Invalid("invalid labels")
	}
	return string(payload), nil
}

// uniqueUIDs 去掉空白项并去重，保持首次出现的顺序。
func uniqueUIDs(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
