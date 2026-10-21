package alertpage

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

const (
	tabMine  = "mine"
	kindMine = "mine"
	kindPage = "page"
	// MaxRealtimeTabs 是实时告警页同时显示的告警页面数量上限，不含固定的「我的告警」。
	MaxRealtimeTabs = 10
)

// Usecase 管理告警页面定义与实时 Tab 订阅。
type Usecase struct {
	pages  repository.AlertPageRepo
	access access.Checker
}

func NewUsecase(pages repository.AlertPageRepo, checker access.Checker) *Usecase {
	return &Usecase{pages: pages, access: checker}
}

func (uc *Usecase) Create(ctx context.Context, in bo.AlertPage) (*bo.AlertPage, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertWrite)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	labels, err := labelsutil.NormalizeMatchers(in.Labels)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.OwnerUID = userUID
	in.Name = strings.TrimSpace(in.Name)
	in.Labels = labels
	in.Remark = strings.TrimSpace(in.Remark)
	item, err := uc.pages.CreateAlertPage(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Update(ctx context.Context, in bo.AlertPage) (*bo.AlertPage, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertWrite)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	existing, err := uc.pages.GetAlertPage(ctx, ns, in.UID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if existing.OwnerUID != userUID {
		return nil, merr.Forbidden("forbidden")
	}
	labels, err := labelsutil.NormalizeMatchers(in.Labels)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	in.Labels = labels
	in.Remark = strings.TrimSpace(in.Remark)
	item, err := uc.pages.UpdateAlertPage(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Delete(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.AlertWrite)
	if err != nil {
		return err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return merr.Unauthenticated("unauthorized")
	}
	existing, err := uc.pages.GetAlertPage(ctx, ns, uid)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if existing.OwnerUID != userUID {
		return merr.Forbidden("forbidden")
	}
	return uc.access.Wrap(uc.pages.DeleteAlertPage(ctx, ns, uid))
}

func (uc *Usecase) Get(ctx context.Context, uid string) (*bo.AlertPage, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	item, err := uc.pages.GetAlertPage(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if item.OwnerUID != userUID && !item.Public {
		return nil, merr.Forbidden("forbidden")
	}
	return item, nil
}

func (uc *Usecase) List(ctx context.Context, pageNo, pageSize int32, name, scope string) ([]bo.AlertPage, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, 0, 0, 0, merr.Unauthenticated("unauthorized")
	}
	if scope == "" {
		scope = "manage"
	}
	if scope != "manage" && scope != "selectable" {
		return nil, 0, 0, 0, merr.Invalid("invalid request")
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.pages.ListAlertPages(ctx, ns, userUID, scope, strings.TrimSpace(name), p, s)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) ListMyTabs(ctx context.Context) ([]bo.RealtimeAlertTab, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	pages, err := uc.pages.ListUserAlertPages(ctx, ns, userUID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	return buildTabs(pages), nil
}

func (uc *Usecase) SetMyTabs(ctx context.Context, pageUIDs []string) ([]bo.RealtimeAlertTab, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	unique := uniqueNonEmpty(pageUIDs)
	if len(unique) > MaxRealtimeTabs {
		return nil, merr.Invalid("too many alert pages")
	}
	if len(unique) > 0 {
		allowed, err := uc.pages.ListSelectableByUIDs(ctx, ns, userUID, unique)
		if err != nil {
			return nil, uc.access.Wrap(err)
		}
		if len(allowed) != len(unique) {
			return nil, merr.Forbidden("forbidden")
		}
	}
	if err := uc.pages.SetUserAlertPages(ctx, ns, userUID, unique); err != nil {
		return nil, uc.access.Wrap(err)
	}
	pages, err := uc.pages.ListUserAlertPages(ctx, ns, userUID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	return buildTabs(pages), nil
}

func buildTabs(pages []bo.AlertPage) []bo.RealtimeAlertTab {
	tabs := make([]bo.RealtimeAlertTab, 0, len(pages)+1)
	tabs = append(tabs, bo.RealtimeAlertTab{ID: tabMine, Name: tabMine, Kind: kindMine})
	for _, page := range pages {
		tabs = append(tabs, bo.RealtimeAlertTab{ID: page.UID, Name: page.Name, Kind: kindPage})
	}
	return tabs
}

func uniqueNonEmpty(uids []string) []string {
	seen := make(map[string]struct{}, len(uids))
	out := make([]string, 0, len(uids))
	for _, uid := range uids {
		uid = strings.TrimSpace(uid)
		if uid == "" || uid == tabMine {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		out = append(out, uid)
	}
	return out
}

const (
	// ColumnKindField 是内置字段列，例如 summary、datasource。
	ColumnKindField = "FIELD"
	// ColumnKindLabel 是标签列，key 为事件标签键。
	ColumnKindLabel = "LABEL"
	// MaxAlertColumns 是单个视图允许配置的列数。
	MaxAlertColumns = 20
	// maxColumnAlias 是列名的长度上限。
	maxColumnAlias = 64
	// maxLabelKey 是标签键的长度上限。
	maxLabelKey = 128
)

// allowedFieldColumns 是实时表格支持的内置字段列。字段名固定，前端只负责渲染。
var allowedFieldColumns = map[string]struct{}{
	"summary": {}, "datasource": {}, "level": {}, "sample_value": {}, "starts_at": {},
	"duration": {}, "intervened": {}, "status": {}, "description": {}, "fingerprint": {},
	"expr": {}, "rule": {},
}

// GetColumns 读取当前用户在 mine 或某个告警页面上的列配置。
func (uc *Usecase) GetColumns(ctx context.Context, scope string) ([]bo.AlertColumn, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	if err := uc.requireScope(ctx, ns, userUID, scope); err != nil {
		return nil, err
	}
	return uc.pages.GetAlertColumns(ctx, ns, userUID, scope)
}

// SetColumns 覆盖保存当前用户在 mine 或某个告警页面上的列配置。
func (uc *Usecase) SetColumns(ctx context.Context, scope string, columns []bo.AlertColumn) ([]bo.AlertColumn, error) {
	ns, err := uc.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	normalized, err := normalizeColumns(columns)
	if err != nil {
		return nil, err
	}
	if err := uc.requireScope(ctx, ns, userUID, scope); err != nil {
		return nil, err
	}
	if err := uc.pages.SetAlertColumns(ctx, ns, userUID, scope, normalized); err != nil {
		return nil, uc.access.Wrap(err)
	}
	return normalized, nil
}

// requireScope 校验视图标识。mine 固定可用；告警页面必须是当前用户自己的或公开的。
func (uc *Usecase) requireScope(ctx context.Context, namespaceUID, userUID, scope string) error {
	scope = strings.TrimSpace(scope)
	if scope == tabMine {
		return nil
	}
	if scope == "" {
		return merr.Invalid("invalid request")
	}
	item, err := uc.pages.GetAlertPage(ctx, namespaceUID, scope)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if item.OwnerUID != userUID && !item.Public {
		return merr.Forbidden("forbidden")
	}
	return nil
}

// normalizeColumns 清理并校验列配置。同一个字段或标签只能出现一次，避免表头重复。
func normalizeColumns(columns []bo.AlertColumn) ([]bo.AlertColumn, error) {
	if len(columns) > MaxAlertColumns {
		return nil, merr.Invalid("too many columns")
	}
	seen := make(map[string]struct{}, len(columns))
	out := make([]bo.AlertColumn, 0, len(columns))
	for _, column := range columns {
		key := strings.TrimSpace(column.Key)
		kind := strings.ToUpper(strings.TrimSpace(column.Type))
		label := strings.TrimSpace(column.Label)
		if key == "" || utf8.RuneCountInString(label) > maxColumnAlias {
			return nil, merr.Invalid("invalid request")
		}
		switch kind {
		case ColumnKindField:
			if _, ok := allowedFieldColumns[key]; !ok {
				return nil, merr.Invalid("invalid request")
			}
		case ColumnKindLabel:
			if len(key) > maxLabelKey || !validLabelKey(key) {
				return nil, merr.Invalid("invalid request")
			}
		default:
			return nil, merr.Invalid("invalid request")
		}
		identity := kind + "\x00" + key
		if _, ok := seen[identity]; ok {
			return nil, merr.Invalid("invalid request")
		}
		seen[identity] = struct{}{}
		out = append(out, bo.AlertColumn{Key: key, Type: kind, Label: label})
	}
	return out, nil
}

// validLabelKey 与 Prometheus 标签名保持一致，避免把任意文本当成列键。
func validLabelKey(key string) bool {
	for index, r := range key {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
		case r >= '0' && r <= '9':
			if index == 0 {
				return false
			}
		default:
			return false
		}
	}
	return key != ""
}
