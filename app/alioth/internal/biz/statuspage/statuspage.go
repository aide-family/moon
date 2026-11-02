// Package statuspage 管理对外公开的探测状态页：配置、公开链接与展示数据。
package statuspage

import (
	"context"
	"log/slog"
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
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	// maxNameRunes 与 proto 校验保持一致。
	maxNameRunes = 64
	// maxTitleRunes 是公开标题允许的字符数。
	maxTitleRunes = 128
	// maxDescriptionRunes 是描述允许的字符数。
	maxDescriptionRunes = 1024
	// maxGroups 是状态页展示的目标组数量上限。
	maxGroups = 50
)

// Gateway 管理状态页与公开数据。
type Gateway struct {
	pages   repository.StatusPageRepo
	targets repository.ProbeTargetRepo
	access  access.Checker
	log     *slog.Logger
}

func NewGateway(pages repository.StatusPageRepo, targets repository.ProbeTargetRepo, checker access.Checker, logger *slog.Logger) *Gateway {
	return &Gateway{pages: pages, targets: targets, access: checker, log: logger}
}

// Create 校验并保存状态页。公开 token 由仓储生成。
func (gateway *Gateway) Create(ctx context.Context, in bo.StatusPage) (*bo.StatusPage, error) {
	ns, err := gateway.access.Allow(ctx, perm.StatusPageWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.CreatedBy, _ = actor.UserID(ctx)
	if err := gateway.normalize(ctx, &in); err != nil {
		return nil, err
	}
	item, err := gateway.pages.CreateStatusPage(ctx, in)
	return item, gateway.access.Wrap(err)
}

// Update 更新状态页配置，token 不变。
func (gateway *Gateway) Update(ctx context.Context, in bo.StatusPage) (*bo.StatusPage, error) {
	ns, err := gateway.access.Allow(ctx, perm.StatusPageWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := gateway.normalize(ctx, &in); err != nil {
		return nil, err
	}
	item, err := gateway.pages.UpdateStatusPage(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateStatus 启停状态页。停用后公开链接立即失效。
func (gateway *Gateway) UpdateStatus(ctx context.Context, uid, status string) (*bo.StatusPage, error) {
	ns, err := gateway.access.Allow(ctx, perm.StatusPageWrite)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" && status != "DISABLED" {
		return nil, merr.Invalid("invalid status")
	}
	item, err := gateway.pages.UpdateStatusPageStatus(ctx, ns, uid, status)
	return item, gateway.access.Wrap(err)
}

// RegenerateToken 重新生成公开链接凭证，旧链接立即失效。
func (gateway *Gateway) RegenerateToken(ctx context.Context, uid string) (*bo.StatusPage, error) {
	ns, err := gateway.access.Allow(ctx, perm.StatusPageWrite)
	if err != nil {
		return nil, err
	}
	item, err := gateway.pages.RegenerateStatusPageToken(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// Delete 软删除状态页。
func (gateway *Gateway) Delete(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.StatusPageDelete)
	if err != nil {
		return err
	}
	return gateway.access.Wrap(gateway.pages.DeleteStatusPage(ctx, ns, uid))
}

// Get 读取状态页详情，包含公开 token。
func (gateway *Gateway) Get(ctx context.Context, uid string) (*bo.StatusPage, error) {
	ns, err := gateway.access.Allow(ctx, perm.StatusPageRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.pages.GetStatusPage(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// List 分页列出状态页。列表项带 token，便于直接复制公开链接。
func (gateway *Gateway) List(ctx context.Context, pageNo, pageSize int32, name, status string) ([]bo.StatusPage, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.StatusPageRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	status = strings.TrimSpace(status)
	if status != "" && status != "ACTIVE" && status != "DISABLED" {
		return nil, 0, 0, 0, merr.Invalid("invalid status")
	}
	items, total, err := gateway.pages.ListStatusPages(ctx, ns, bo.StatusPageListFilter{
		Name: strings.TrimSpace(name), Status: status,
	}, p, s)
	if err != nil {
		return nil, 0, 0, 0, gateway.access.Wrap(err)
	}
	return items, total, p, s, nil
}

// Public 按 token 返回公开数据；不需要登录态与租户头。
func (gateway *Gateway) Public(ctx context.Context, token string) (*bo.PublicStatusPage, error) {
	return gateway.pages.PublicStatusPage(ctx, token, time.Now().UTC())
}

// normalize 校验名称、标题、描述与目标组归属，并去重目标组。
func (gateway *Gateway) normalize(ctx context.Context, in *bo.StatusPage) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || utf8.RuneCountInString(in.Name) > maxNameRunes {
		return merr.Invalid("invalid status page name")
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" || utf8.RuneCountInString(in.Title) > maxTitleRunes {
		return merr.Invalid("invalid status page title")
	}
	in.Description = strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(in.Description) > maxDescriptionRunes {
		return merr.Invalid("status page description is too long")
	}
	groups := dedupeStrings(in.TargetGroupUIDs)
	if len(groups) == 0 {
		return merr.Invalid("status page needs at least one target group")
	}
	if len(groups) > maxGroups {
		return merr.Invalid("too many target groups")
	}
	for _, uid := range groups {
		if _, err := gateway.targets.GetProbeTargetGroup(ctx, in.NamespaceUID, uid); err != nil {
			return err
		}
	}
	in.TargetGroupUIDs = groups
	return nil
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
