// Package alertmaintenance 管理告警维护窗口：计划维护期间按标签、级别、
// 数据源与规则静默匹配的告警，不生成通知任务；窗口结束后抑制自动到期。
package alertmaintenance

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	// maxNameRunes 与 proto 校验保持一致。
	maxNameRunes = 64
	// maxDescriptionRunes 是描述允许的字符数。
	maxDescriptionRunes = 1024
	// maxLabelsBytes 是标签匹配器 JSON 允许的字节数。
	maxLabelsBytes = 4096
)

// Gateway 管理维护窗口。
type Gateway struct {
	windows repository.AlertMaintenanceRepo
	access  access.Checker
	log     *slog.Logger
}

func NewGateway(windows repository.AlertMaintenanceRepo, checker access.Checker, logger *slog.Logger) *Gateway {
	return &Gateway{windows: windows, access: checker, log: logger}
}

// Create 校验并保存维护窗口；窗口当前生效时会抑制已触发且匹配的事件。
func (gateway *Gateway) Create(ctx context.Context, in bo.AlertMaintenanceWindow) (*bo.AlertMaintenanceWindow, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertMaintenanceWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.CreatedBy, _ = actor.UserID(ctx)
	if err := normalizeWindow(&in); err != nil {
		return nil, err
	}
	item, err := gateway.windows.CreateWindow(ctx, in)
	return item, gateway.access.Wrap(err)
}

// Update 更新窗口；旧配置设置的抑制先释放，再按新配置重新抑制。
func (gateway *Gateway) Update(ctx context.Context, in bo.AlertMaintenanceWindow) (*bo.AlertMaintenanceWindow, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertMaintenanceWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeWindow(&in); err != nil {
		return nil, err
	}
	item, err := gateway.windows.UpdateWindow(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateStatus 启停窗口；停用立即解除它设置的抑制。
func (gateway *Gateway) UpdateStatus(ctx context.Context, uid, status string) (*bo.AlertMaintenanceWindow, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertMaintenanceWrite)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" && status != "DISABLED" {
		return nil, merr.Invalid("invalid status")
	}
	item, err := gateway.windows.UpdateWindowStatus(ctx, ns, uid, status)
	return item, gateway.access.Wrap(err)
}

// Delete 软删除窗口并解除它设置的抑制。
func (gateway *Gateway) Delete(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.AlertMaintenanceDelete)
	if err != nil {
		return err
	}
	return gateway.access.Wrap(gateway.windows.DeleteWindow(ctx, ns, uid))
}

// Get 读取维护窗口详情。
func (gateway *Gateway) Get(ctx context.Context, uid string) (*bo.AlertMaintenanceWindow, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertMaintenanceRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.windows.GetWindow(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// List 分页列出维护窗口。
func (gateway *Gateway) List(ctx context.Context, pageNo, pageSize int32, name, status string) ([]bo.AlertMaintenanceWindow, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.AlertMaintenanceRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	status = strings.TrimSpace(status)
	if status != "" && status != "ACTIVE" && status != "DISABLED" {
		return nil, 0, 0, 0, merr.Invalid("invalid status")
	}
	items, total, err := gateway.windows.ListWindows(ctx, ns, bo.AlertMaintenanceWindowFilter{
		Name: strings.TrimSpace(name), Status: status,
	}, p, s)
	if err != nil {
		return nil, 0, 0, 0, gateway.access.Wrap(err)
	}
	return items, total, p, s, nil
}

// normalizeWindow 校验名称、标签匹配器与生效区间，并去重匹配值。
func normalizeWindow(in *bo.AlertMaintenanceWindow) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || utf8.RuneCountInString(in.Name) > maxNameRunes {
		return merr.Invalid("invalid maintenance window name")
	}
	in.Description = strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(in.Description) > maxDescriptionRunes {
		return merr.Invalid("description is too long")
	}
	labels, err := normalizeLabelMatcher(in.Labels)
	if err != nil {
		return err
	}
	in.Labels = labels
	in.LevelUIDs = dedupeStrings(in.LevelUIDs)
	in.DatasourceUIDs = dedupeStrings(in.DatasourceUIDs)
	in.RuleUIDs = dedupeStrings(in.RuleUIDs)
	if in.StartsAt.IsZero() || in.EndsAt.IsZero() || !in.EndsAt.After(in.StartsAt) {
		return merr.Invalid("invalid maintenance window time range")
	}
	return nil
}

// normalizeLabelMatcher 把标签匹配器规范成排序后的紧凑 JSON 对象；
// 空串保存为空对象，非对象或非字符串值拒绝。
func normalizeLabelMatcher(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}", nil
	}
	if len(raw) > maxLabelsBytes {
		return "", merr.Invalid("label matcher is too long")
	}
	var matcher map[string]string
	if err := json.Unmarshal([]byte(raw), &matcher); err != nil {
		return "", merr.Invalid("invalid label matcher")
	}
	if len(matcher) == 0 {
		return "{}", nil
	}
	encoded, err := json.Marshal(matcher)
	if err != nil {
		return "", merr.Invalid("invalid label matcher")
	}
	return string(encoded), nil
}

func dedupeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
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
