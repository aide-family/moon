// Package board 管理 Grafana 大盘：校验名称、面板链接与布局参数，
// 并按当前租户的权限读写大盘。浏览器直接嵌入 Grafana，控制面只保存展示元数据。
package board

import (
	"context"
	"log/slog"
	"net/url"
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
	// maxPanelTitleRunes 是面板标题允许的字符数。
	maxPanelTitleRunes = 128
	// maxURLBytes 是 Grafana 地址允许的最大长度。
	maxURLBytes = 2048
	// maxDescriptionRunes 是描述允许的字符数。
	maxDescriptionRunes = 1024
	// maxPanels 是单块大盘允许的面板数量，和 proto 的 repeated.max_items 一致。
	maxPanels = 50
	// minPanelHeight 与 maxPanelHeight 与 proto 校验一致，单位像素。
	minPanelHeight = 120
	maxPanelHeight = 1600
)

// Gateway 管理大盘与面板。
type Gateway struct {
	boards repository.BoardRepo
	access access.Checker
	log    *slog.Logger
}

func NewGateway(boards repository.BoardRepo, checker access.Checker, logger *slog.Logger) *Gateway {
	return &Gateway{boards: boards, access: checker, log: logger}
}

// CreateBoard 校验并保存大盘；创建人来自当前登录用户。
func (gateway *Gateway) CreateBoard(ctx context.Context, in bo.Board) (*bo.Board, error) {
	ns, err := gateway.access.Allow(ctx, perm.BoardWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.CreatedBy, _ = actor.UserID(ctx)
	if err := normalizeBoard(&in); err != nil {
		return nil, err
	}
	item, err := gateway.boards.CreateBoard(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateBoard 校验并整体替换大盘的面板列表。
func (gateway *Gateway) UpdateBoard(ctx context.Context, in bo.Board) (*bo.Board, error) {
	ns, err := gateway.access.Allow(ctx, perm.BoardWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := normalizeBoard(&in); err != nil {
		return nil, err
	}
	item, err := gateway.boards.UpdateBoard(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateBoardStatus 启停大盘。
func (gateway *Gateway) UpdateBoardStatus(ctx context.Context, uid, status string) (*bo.Board, error) {
	ns, err := gateway.access.Allow(ctx, perm.BoardWrite)
	if err != nil {
		return nil, err
	}
	if status != "ACTIVE" && status != "DISABLED" {
		return nil, merr.Invalid("invalid status")
	}
	item, err := gateway.boards.UpdateBoardStatus(ctx, ns, uid, status)
	return item, gateway.access.Wrap(err)
}

// DeleteBoard 软删除大盘及其面板。
func (gateway *Gateway) DeleteBoard(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.BoardDelete)
	if err != nil {
		return err
	}
	return gateway.access.Wrap(gateway.boards.DeleteBoard(ctx, ns, uid))
}

// GetBoard 读取大盘详情，包含面板。
func (gateway *Gateway) GetBoard(ctx context.Context, uid string) (*bo.Board, error) {
	ns, err := gateway.access.Allow(ctx, perm.BoardRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.boards.GetBoard(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// ListBoards 分页列出大盘，不返回面板。
func (gateway *Gateway) ListBoards(ctx context.Context, pageNo, pageSize int32, name, status string) ([]bo.Board, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.BoardRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	status = strings.TrimSpace(status)
	if status != "" && status != "ACTIVE" && status != "DISABLED" {
		return nil, 0, 0, 0, merr.Invalid("invalid status")
	}
	items, total, err := gateway.boards.ListBoards(ctx, ns, bo.BoardListFilter{
		Name: strings.TrimSpace(name), Status: status,
	}, p, s)
	if err != nil {
		return nil, 0, 0, 0, gateway.access.Wrap(err)
	}
	return items, total, p, s, nil
}

// normalizeBoard 校验名称与描述，并归一化每个面板。面板顺序即展示顺序。
func normalizeBoard(in *bo.Board) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || utf8.RuneCountInString(in.Name) > maxNameRunes {
		return merr.Invalid("invalid board name")
	}
	in.Description = strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(in.Description) > maxDescriptionRunes {
		return merr.Invalid("description is too long")
	}
	if len(in.Panels) > maxPanels {
		return merr.Invalid("too many panels")
	}
	for index := range in.Panels {
		panel := &in.Panels[index]
		panel.Title = strings.TrimSpace(panel.Title)
		if panel.Title == "" || utf8.RuneCountInString(panel.Title) > maxPanelTitleRunes {
			return merr.Invalid("invalid panel title")
		}
		panel.Description = strings.TrimSpace(panel.Description)
		if utf8.RuneCountInString(panel.Description) > maxDescriptionRunes {
			return merr.Invalid("panel description is too long")
		}
		panel.GrafanaURL = strings.TrimSpace(panel.GrafanaURL)
		if len(panel.GrafanaURL) > maxURLBytes {
			return merr.Invalid("grafana url is too long")
		}
		if err := validateGrafanaURL(panel.GrafanaURL); err != nil {
			return err
		}
		if panel.Width < 1 || panel.Width > 12 {
			return merr.Invalid("invalid panel width")
		}
		if panel.Height < minPanelHeight || panel.Height > maxPanelHeight {
			return merr.Invalid("invalid panel height")
		}
		panel.Sort = index
	}
	return nil
}

// validateGrafanaURL 校验 Grafana 面板或仪表盘的分享 / 嵌入地址。
// 地址由浏览器直接加载，控制面不代理；这里拦掉可识别的危险形式：
// 非 http/https、带 userinfo、带 fragment、路径不是 Grafana 的 dashboard /
// panel / 公开面板路径，以及 query 里带凭据类参数名。
func validateGrafanaURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return merr.Invalid("invalid grafana url")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return merr.Invalid("invalid grafana url")
	}
	if parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return merr.Invalid("invalid grafana url")
	}
	path := parsed.Path
	if !strings.Contains(path, "/d/") && !strings.Contains(path, "/d-solo/") &&
		!strings.Contains(path, "/public-dashboards/") {
		return merr.Invalid("invalid grafana url")
	}
	for name := range parsed.Query() {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") ||
			strings.Contains(lower, "password") || strings.Contains(lower, "api_key") ||
			strings.Contains(lower, "apikey") {
			return merr.Invalid("grafana url must not contain credentials")
		}
	}
	return nil
}
