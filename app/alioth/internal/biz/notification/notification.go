package notification

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/alerttmpl"
	"github.com/aide-family/moon/pkg/lock"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewCenter)

// Center 管理通知渠道、模板、通知组和投递任务。
type Center struct {
	repo    repository.NotificationRepo
	members repository.MemberRepo
	access  access.Checker
	locks   *lock.Client
	log     *slog.Logger
}

func NewCenter(repo repository.NotificationRepo, members repository.MemberRepo, checker access.Checker, locks *lock.Client, logger *slog.Logger) *Center {
	return &Center{repo: repo, members: members, access: checker, locks: locks, log: logger}
}

const (
	reapLockKey = "alioth:lock:schedule:notify-reap"
)

// staleAfter 读取 SENDING 回收阈值，缺省 2 分钟，可通过环境变量调整。
func staleAfter() time.Duration {
	raw := strings.TrimSpace(os.Getenv("ALIOTH_NOTIFY_STALE_SECONDS"))
	if raw == "" {
		return 2 * time.Minute
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 2 * time.Minute
	}
	return time.Duration(seconds) * time.Second
}

// ReapStale 回收长时间停留在 SENDING 的任务。merak 崩溃后不会回写，只能由控制面兜底。
func (center *Center) ReapStale(ctx context.Context) error {
	if center.locks == nil {
		return merr.Server("lock failed")
	}
	hold, err := center.locks.Acquire(ctx, reapLockKey, 15*time.Second)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	reaped, err := center.repo.ReapStaleSending(ctx, time.Now().Add(-staleAfter()), 100)
	if err != nil {
		return center.access.Wrap(err)
	}
	if reaped > 0 && center.log != nil {
		center.log.Warn("reaped stale notification tasks", slog.Int("count", reaped))
	}
	return nil
}

func (center *Center) CreateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationChannelWrite)
	if err != nil {
		return nil, err
	}
	normalizeChannel(&in)
	if err := validateChannel(in); err != nil {
		return nil, err
	}
	if err := center.requireChannelCapability(ctx, in.Type, in.App); err != nil {
		return nil, err
	}
	if in.Type == channelTypeSMS && strings.TrimSpace(in.Password) == "" {
		return nil, merr.Invalid("sms access key is required")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, merr.Invalid("name is required")
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := center.repo.CreateChannel(ctx, in)
	return redactChannel(item), center.access.Wrap(err)
}

func (center *Center) UpdateChannel(ctx context.Context, in bo.NotificationChannel) (*bo.NotificationChannel, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationChannelWrite)
	if err != nil {
		return nil, err
	}
	normalizeChannel(&in)
	if err := validateChannel(in); err != nil {
		return nil, err
	}
	if err := center.requireChannelCapability(ctx, in.Type, in.App); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, merr.Invalid("name is required")
	}
	previous, err := center.repo.GetChannel(ctx, ns, in.UID)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	if err := center.requireChannelCapability(ctx, previous.Type, previous.App); err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := center.repo.UpdateChannel(ctx, in)
	return redactChannel(item), center.access.Wrap(err)
}

func (center *Center) UpdateChannelStatus(ctx context.Context, uid, status string) (*bo.NotificationChannel, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationChannelWrite)
	if err != nil {
		return nil, err
	}
	current, err := center.repo.GetChannel(ctx, ns, uid)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	if err := center.requireChannelCapability(ctx, current.Type, current.App); err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateChannelStatus(ctx, ns, uid, status)
	return redactChannel(item), center.access.Wrap(err)
}

func (center *Center) DeleteChannel(ctx context.Context, uid string) error {
	ns, err := center.access.Allow(ctx, perm.NotificationChannelDelete)
	if err != nil {
		return err
	}
	current, err := center.repo.GetChannel(ctx, ns, uid)
	if err != nil {
		return center.access.Wrap(err)
	}
	if err := center.requireChannelCapability(ctx, current.Type, current.App); err != nil {
		return err
	}
	return center.access.Wrap(center.repo.DeleteChannel(ctx, ns, uid))
}

func (center *Center) GetChannel(ctx context.Context, uid string) (*bo.NotificationChannel, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.GetChannel(ctx, ns, uid)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	if err := center.requireChannelCapability(ctx, item.Type, item.App); err != nil {
		return nil, err
	}
	return redactChannel(item), nil
}

func (center *Center) ListChannels(ctx context.Context, pageNo, pageSize int32, name, kind, app string) ([]bo.NotificationChannel, int, int, int, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	scopes, err := center.authorizedChannelScopes(ctx)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	// 请求的类型和应用只能在已授权组合里收窄，不能借筛选绕过能力检查。
	scopes = filterChannelScopes(scopes, strings.ToUpper(strings.TrimSpace(kind)), strings.ToUpper(strings.TrimSpace(app)))
	if len(scopes) == 0 {
		return []bo.NotificationChannel{}, 0, p, s, nil
	}
	items, total, err := center.repo.ListChannels(ctx, ns, p, s, strings.TrimSpace(name), scopes)
	out := make([]bo.NotificationChannel, 0, len(items))
	for i := range items {
		if item := redactChannel(&items[i]); item != nil {
			out = append(out, *item)
		}
	}
	return out, total, p, s, center.access.Wrap(err)
}

func (center *Center) SelectChannels(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, err
	}
	scopes, err := center.authorizedChannelScopes(ctx)
	if err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return []bo.SelectItem{}, nil
	}
	items, err := center.repo.SelectChannels(ctx, ns, page.Limit(limit), lastUID, scopes)
	return items, center.access.Wrap(err)
}

// TestChannel 只写入 PENDING 任务。短信类型直接拒绝，避免被报成 SENT。
func (center *Center) TestChannel(ctx context.Context, channelUID, templateUID, receivers, subject, body, params string) (string, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationChannelWrite)
	if err != nil {
		return "", err
	}
	channel, err := center.repo.GetChannel(ctx, ns, channelUID)
	if err != nil {
		return "", center.access.Wrap(err)
	}
	if err := center.requireChannelCapability(ctx, channel.Type, channel.App); err != nil {
		return "", err
	}
	if channel.Status != "ACTIVE" {
		return "", merr.Invalid("channel is disabled")
	}
	if channel.App == "SMS_ALICLOUD" {
		return "", merr.Invalid("sms delivery is not implemented")
	}
	if err := validateReceivers(channel.Type, receivers); err != nil {
		return "", err
	}
	if templateUID != "" {
		template, err := center.repo.GetTemplate(ctx, ns, templateUID)
		if err != nil {
			return "", center.access.Wrap(err)
		}
		if err := center.requireTemplateCapability(ctx, template.Type, template.App); err != nil {
			return "", err
		}
		if template.Status != "ACTIVE" {
			return "", merr.Invalid("template is disabled")
		}
		if err := validateGroupBinding(*channel, *template); err != nil {
			return "", err
		}
		if subject == "" {
			subject = template.Subject
		}
		if body == "" {
			body = template.Body
		}
	}
	if strings.TrimSpace(body) == "" {
		return "", merr.Invalid("body is required")
	}
	if channel.Type == "SMTP" && strings.TrimSpace(subject) == "" {
		return "", merr.Invalid("subject is required")
	}
	task, err := center.repo.CreateTestTask(ctx, bo.NotificationTask{
		NamespaceUID: ns,
		ChannelUID:   channelUID,
		TemplateUID:  templateUID,
		Receivers:    strings.TrimSpace(receivers),
		Subject:      subject,
		Body:         body,
		Params:       params,
	})
	if err != nil {
		return "", center.access.Wrap(err)
	}
	return task.UID, nil
}

func (center *Center) CreateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationTemplateWrite)
	if err != nil {
		return nil, err
	}
	normalizeTemplate(&in)
	if err := validateTemplate(in); err != nil {
		return nil, err
	}
	if err := center.requireTemplateCapability(ctx, in.Type, in.App); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, merr.Invalid("name is required")
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := center.repo.CreateTemplate(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateTemplate(ctx context.Context, in bo.NotificationTemplate) (*bo.NotificationTemplate, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationTemplateWrite)
	if err != nil {
		return nil, err
	}
	normalizeTemplate(&in)
	if err := validateTemplate(in); err != nil {
		return nil, err
	}
	if err := center.requireTemplateCapability(ctx, in.Type, in.App); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, merr.Invalid("name is required")
	}
	previous, err := center.repo.GetTemplate(ctx, ns, in.UID)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	if err := center.requireTemplateCapability(ctx, previous.Type, previous.App); err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	item, err := center.repo.UpdateTemplate(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateTemplateStatus(ctx context.Context, uid, status string) (*bo.NotificationTemplate, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationTemplateWrite)
	if err != nil {
		return nil, err
	}
	current, err := center.repo.GetTemplate(ctx, ns, uid)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	if err := center.requireTemplateCapability(ctx, current.Type, current.App); err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateTemplateStatus(ctx, ns, uid, status)
	return item, center.access.Wrap(err)
}

func (center *Center) DeleteTemplate(ctx context.Context, uid string) error {
	ns, err := center.access.Allow(ctx, perm.NotificationTemplateDelete)
	if err != nil {
		return err
	}
	current, err := center.repo.GetTemplate(ctx, ns, uid)
	if err != nil {
		return center.access.Wrap(err)
	}
	if err := center.requireTemplateCapability(ctx, current.Type, current.App); err != nil {
		return err
	}
	return center.access.Wrap(center.repo.DeleteTemplate(ctx, ns, uid))
}

func (center *Center) GetTemplate(ctx context.Context, uid string) (*bo.NotificationTemplate, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.GetTemplate(ctx, ns, uid)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	if err := center.requireTemplateCapability(ctx, item.Type, item.App); err != nil {
		return nil, err
	}
	return item, nil
}

func (center *Center) ListTemplates(ctx context.Context, pageNo, pageSize int32, name, kind, app string) ([]bo.NotificationTemplate, int, int, int, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	scopes, err := center.authorizedTemplateScopes(ctx)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	scopes = filterTemplateScopes(scopes, strings.ToUpper(strings.TrimSpace(kind)), strings.ToUpper(strings.TrimSpace(app)))
	if len(scopes) == 0 {
		return []bo.NotificationTemplate{}, 0, p, s, nil
	}
	items, total, err := center.repo.ListTemplates(ctx, ns, p, s, strings.TrimSpace(name), scopes)
	return items, total, p, s, center.access.Wrap(err)
}

func (center *Center) SelectTemplates(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, err
	}
	scopes, err := center.authorizedTemplateScopes(ctx)
	if err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return []bo.SelectItem{}, nil
	}
	items, err := center.repo.SelectTemplates(ctx, ns, page.Limit(limit), lastUID, scopes)
	return items, center.access.Wrap(err)
}

// authorizedChannelScopes 返回租户已开通的通知渠道「类型 + 应用」组合。
// 渠道、模板和通知组都按它过滤：未授权的应用连同它的模板一起隐藏。
func (center *Center) authorizedChannelScopes(ctx context.Context) ([]perm.ChannelScope, error) {
	capabilities, err := center.access.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	return perm.ChannelScopesOf(capabilities), nil
}

// authorizedTemplateScopes 把已开通渠道能力映射成可用的模板「类型 + 应用」组合。
func (center *Center) authorizedTemplateScopes(ctx context.Context) ([]perm.TemplateScope, error) {
	capabilities, err := center.access.Capabilities(ctx)
	if err != nil {
		return nil, err
	}
	return perm.TemplateScopesOf(capabilities), nil
}

// requireChannelCapability 要求租户已开通该渠道「类型 + 应用」对应能力。
func (center *Center) requireChannelCapability(ctx context.Context, channelType, app string) error {
	capability := perm.ChannelCapabilityOf(channelType, app)
	if capability == "" {
		return merr.Invalid("invalid request")
	}
	return center.access.RequireCapability(ctx, capability)
}

// requireTemplateCapability 要求模板对应的渠道能力已开通：模板是渠道的正文，
// 渠道能力不可用时模板也不能创建、修改、预览或测试。
func (center *Center) requireTemplateCapability(ctx context.Context, templateType, app string) error {
	capability := perm.TemplateCapabilityOf(templateType, app)
	if capability == "" {
		return merr.Invalid("invalid request")
	}
	return center.access.RequireCapability(ctx, capability)
}

// filterChannelScopes 按请求的类型和应用收窄授权范围；不在范围内时返回空，
// 列表不会退回全量。
func filterChannelScopes(scopes []perm.ChannelScope, channelType, app string) []perm.ChannelScope {
	wanted := make([]perm.ChannelScope, 0, len(scopes))
	for _, scope := range scopes {
		if channelType != "" && scope.Type != channelType {
			continue
		}
		if app != "" && scope.App != app {
			continue
		}
		wanted = append(wanted, scope)
	}
	return wanted
}

// filterTemplateScopes 按请求的模板类型和应用收窄授权范围。
func filterTemplateScopes(scopes []perm.TemplateScope, templateType, app string) []perm.TemplateScope {
	wanted := make([]perm.TemplateScope, 0, len(scopes))
	for _, scope := range scopes {
		if templateType != "" && scope.Type != templateType {
			continue
		}
		if app != "" && scope.App != app {
			continue
		}
		wanted = append(wanted, scope)
	}
	return wanted
}

// channelScopeSet 把渠道授权组合转成 "类型|应用" 集合，供通知组成对过滤时查询。
func channelScopeSet(scopes []perm.ChannelScope) map[string]struct{} {
	set := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		set[scope.Type+"|"+scope.App] = struct{}{}
	}
	return set
}

// templateScopeSet 把模板授权组合转成 "类型|应用" 集合。
func templateScopeSet(scopes []perm.TemplateScope) map[string]struct{} {
	set := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		set[scope.Type+"|"+scope.App] = struct{}{}
	}
	return set
}

// ListPresets 返回系统预设模板。预设是只读的，用户从预设创建自己的模板。
func (center *Center) ListPresets(ctx context.Context) ([]bo.NotificationPreset, error) {
	if _, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...); err != nil {
		return nil, err
	}
	items, err := center.repo.ListPresets(ctx)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	scopes, err := center.authorizedTemplateScopes(ctx)
	if err != nil {
		return nil, err
	}
	allowed := templateScopeSet(scopes)
	out := make([]bo.NotificationPreset, 0, len(items))
	for _, item := range items {
		// 预设模板也按「类型 + 应用」过滤：渠道能力没开通时不应从预设里创建对应模板。
		app := item.App
		if app == "" {
			app = "OTHER"
		}
		if _, ok := allowed[item.Type+"|"+app]; !ok {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func (center *Center) TestTemplate(ctx context.Context, templateUID, channelUID, receivers, params string) (string, error) {
	return center.TestChannel(ctx, channelUID, templateUID, receivers, "", "", params)
}

// PreviewTemplate 用固定的 mock 事件渲染模板，只用于编辑器和预设页预览，不创建任务。
// 渲染和投递共用 pkg/alerttmpl，因此预览里能用的字段、函数与真实发送一致。
func (center *Center) PreviewTemplate(ctx context.Context, in bo.NotificationTemplate) (string, string, error) {
	if _, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...); err != nil {
		return "", "", err
	}
	normalizeTemplate(&in)
	switch in.Type {
	case templateTypeEmail, templateTypeHook, templateTypeSMS:
	default:
		return "", "", merr.Invalid("invalid template type")
	}
	if err := center.requireTemplateCapability(ctx, in.Type, in.App); err != nil {
		return "", "", err
	}
	if err := alerttmpl.Validate(in.Subject, "subject"); err != nil {
		return "", "", merr.Invalid("invalid subject template")
	}
	if err := alerttmpl.Validate(in.Body, "body"); err != nil {
		return "", "", merr.Invalid("invalid body template")
	}
	labels := map[string]string{
		"app": "metallb", "namespace": "metallb-system", "pod": "controller-6bb4fd4547-fhvh4",
		"job": "kubernetes-pods", "instance": "192.168.182.10:7472",
		"alert_level": "P0", "datasource": "prometheus", "datasource_level": "低优",
	}
	labelsJSON, _ := json.Marshal(labels)
	externalURL := "http://prometheus.example.com/graph"
	params := map[string]any{
		"summary":         "测试告警：metallb controller 不可用",
		"description":     "这是 mock 数据，用于预览模板渲染效果。",
		"phase":           "FIRING",
		"eventUid":        "evt-preview-0001",
		"ruleUid":         "rule-preview-0001",
		"subscriptionUid": "sub-preview-0001",
		"groupUid":        "group-preview-0001",
		"fingerprint":     "fp-preview-0001",
		"externalURL":     externalURL,
	}
	data := alerttmpl.NewData(labels, string(labelsJSON), 1, externalURL, params)
	parsedURL, _ := url.Parse(externalURL)
	renderCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	renderedSubject, err := alerttmpl.Render(renderCtx, in.Subject, data, "subject", time.Now(), nil, parsedURL)
	if err != nil {
		return "", "", merr.Invalid("invalid subject template")
	}
	renderedBody, err := alerttmpl.Render(renderCtx, in.Body, data, "body", time.Now(), nil, parsedURL)
	if err != nil {
		return "", "", merr.Invalid("invalid body template")
	}
	return renderedSubject, renderedBody, nil
}

func normalizeChannel(in *bo.NotificationChannel) {
	in.Type = strings.ToUpper(strings.TrimSpace(in.Type))
	in.App = strings.ToUpper(strings.TrimSpace(in.App))
	if in.Type == channelTypeSMTP || in.Type == channelTypeSMS {
		in.App = appSMSAliCloud
		if in.Type == channelTypeSMTP {
			in.App = "OTHER"
		}
	}
}

func normalizeTemplate(in *bo.NotificationTemplate) {
	in.Type = strings.ToUpper(strings.TrimSpace(in.Type))
	in.App = strings.ToUpper(strings.TrimSpace(in.App))
	if in.Type == templateTypeEmail || in.Type == templateTypeSMS {
		in.App = "OTHER"
	}
	if in.App == "" {
		in.App = "OTHER"
	}
}

func (center *Center) CreateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationGroupWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := center.normalizeGroup(ctx, &in); err != nil {
		return nil, err
	}
	item, err := center.repo.CreateGroup(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateGroup(ctx context.Context, in bo.NotificationGroup) (*bo.NotificationGroup, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationGroupWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := center.normalizeGroup(ctx, &in); err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateGroup(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateGroupStatus(ctx context.Context, uid, status string) (*bo.NotificationGroup, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationGroupWrite)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateGroupStatus(ctx, ns, uid, status)
	return item, center.access.Wrap(err)
}

func (center *Center) DeleteGroup(ctx context.Context, uid string) error {
	ns, err := center.access.Allow(ctx, perm.NotificationGroupDelete)
	if err != nil {
		return err
	}
	return center.access.Wrap(center.repo.DeleteGroup(ctx, ns, uid))
}

func (center *Center) GetGroup(ctx context.Context, uid string) (*bo.NotificationGroup, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.GetGroup(ctx, ns, uid)
	if err != nil {
		return nil, center.access.Wrap(err)
	}
	groups := []bo.NotificationGroup{*item}
	if err := center.filterGroupAccess(ctx, ns, groups); err != nil {
		return nil, err
	}
	return &groups[0], nil
}

func (center *Center) ListGroups(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.NotificationGroup, int, int, int, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := center.repo.ListGroups(ctx, ns, p, s, strings.TrimSpace(name))
	if err != nil {
		return nil, 0, 0, 0, center.access.Wrap(err)
	}
	if err := center.filterGroupAccess(ctx, ns, items); err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, p, s, nil
}

// filterGroupAccess 过滤通知组里已不再授权的渠道与模板配对。
// 渠道类型没开通时，它绑定的模板信息也不返回给前端；历史配置保留在库里，
// 重新授权后会恢复展示，避免降级直接改坏已有配置。
func (center *Center) filterGroupAccess(ctx context.Context, namespaceUID string, groups []bo.NotificationGroup) error {
	if len(groups) == 0 {
		return nil
	}
	channelUIDs := make([]string, 0, len(groups))
	templateUIDs := make([]string, 0, len(groups))
	seenChannels := make(map[string]struct{})
	seenTemplates := make(map[string]struct{})
	for i := range groups {
		normalizeLegacyGroup(&groups[i])
		for _, uid := range groups[i].ChannelUIDs {
			if _, ok := seenChannels[uid]; !ok && uid != "" {
				seenChannels[uid] = struct{}{}
				channelUIDs = append(channelUIDs, uid)
			}
		}
		for _, uid := range groups[i].TemplateUIDs {
			if _, ok := seenTemplates[uid]; !ok && uid != "" {
				seenTemplates[uid] = struct{}{}
				templateUIDs = append(templateUIDs, uid)
			}
		}
	}
	channels, err := center.repo.ListChannelsByUIDs(ctx, namespaceUID, channelUIDs)
	if err != nil {
		return center.access.Wrap(err)
	}
	templates, err := center.repo.ListTemplatesByUIDs(ctx, namespaceUID, templateUIDs)
	if err != nil {
		return center.access.Wrap(err)
	}
	channelScopeByUID := make(map[string]perm.ChannelScope, len(channels))
	for _, item := range channels {
		channelScopeByUID[item.UID] = perm.ChannelScope{Type: item.Type, App: item.App}
	}
	templateScopeByUID := make(map[string]perm.TemplateScope, len(templates))
	for _, item := range templates {
		templateScopeByUID[item.UID] = perm.TemplateScope{Type: item.Type, App: item.App}
	}
	channelScopes, err := center.authorizedChannelScopes(ctx)
	if err != nil {
		return err
	}
	templateScopes, err := center.authorizedTemplateScopes(ctx)
	if err != nil {
		return err
	}
	allowedChannels := channelScopeSet(channelScopes)
	allowedTemplates := templateScopeSet(templateScopes)
	for i := range groups {
		filterGroupPairs(&groups[i], channelScopeByUID, templateScopeByUID, allowedChannels, allowedTemplates)
	}
	return nil
}

// filterGroupPairs 按「渠道类型 + 应用」和「模板类型 + 应用」成对过滤，
// 保持 channel_uids 与 template_uids 一一对应。
func filterGroupPairs(group *bo.NotificationGroup, channelScopeByUID map[string]perm.ChannelScope, templateScopeByUID map[string]perm.TemplateScope, allowedChannels, allowedTemplates map[string]struct{}) {
	channels := make([]string, 0, len(group.ChannelUIDs))
	templates := make([]string, 0, len(group.ChannelUIDs))
	filteredChannels := make(map[string]struct{})
	filteredTemplates := make(map[string]struct{})
	for index, uid := range group.ChannelUIDs {
		channelScope := channelScopeByUID[uid]
		if _, ok := allowedChannels[channelScope.Type+"|"+channelScope.App]; !ok {
			continue
		}
		templateUID := ""
		if index < len(group.TemplateUIDs) {
			templateUID = group.TemplateUIDs[index]
		}
		templateScope := templateScopeByUID[templateUID]
		templateApp := templateScope.App
		if templateApp == "" {
			templateApp = "OTHER"
		}
		if _, ok := allowedTemplates[templateScope.Type+"|"+templateApp]; !ok {
			continue
		}
		channels = append(channels, uid)
		templates = append(templates, templateUID)
		filteredChannels[uid] = struct{}{}
		filteredTemplates[templateUID] = struct{}{}
	}
	group.ChannelUIDs = channels
	group.TemplateUIDs = templates
	group.ChannelUID, group.TemplateUID = "", ""
	if len(channels) > 0 {
		group.ChannelUID, group.TemplateUID = channels[0], templates[0]
	}
	webhooks := make([]string, 0, len(group.WebhookChannelUIDs))
	for _, uid := range group.WebhookChannelUIDs {
		if _, ok := filteredChannels[uid]; ok {
			webhooks = append(webhooks, uid)
		}
	}
	group.WebhookChannelUIDs = webhooks
	if _, ok := filteredTemplates[group.WebhookTemplateUID]; !ok {
		group.WebhookTemplateUID = ""
	}
}

func (center *Center) SelectGroups(ctx context.Context, limit int32, lastUID string, activeOnly bool) ([]bo.SelectItem, error) {
	ns, err := center.access.AllowAny(ctx, perm.NotificationReadPermissions()...)
	if err != nil {
		return nil, err
	}
	items, err := center.repo.SelectGroups(ctx, ns, page.Limit(limit), lastUID, activeOnly)
	return items, center.access.Wrap(err)
}

func (center *Center) ListMessageLogs(ctx context.Context, pageNo, pageSize int32, filter bo.MessageLogListFilter) ([]bo.MessageLog, int, int, int, error) {
	ns, err := center.access.Allow(ctx, perm.MessageLogRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if filter.Status != "" && !validMessageLogStatus(filter.Status) {
		return nil, 0, 0, 0, merr.Invalid("invalid message log status")
	}
	if err := validListTimeRange(filter.StartAt, filter.EndAt); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := center.repo.ListMessageLogs(ctx, ns, filter, p, s)
	return items, total, p, s, center.access.Wrap(err)
}

func (center *Center) GetMessageLog(ctx context.Context, uid string) (*bo.MessageLog, error) {
	ns, err := center.access.Allow(ctx, perm.MessageLogRead)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.GetMessageLog(ctx, ns, uid)
	return item, center.access.Wrap(err)
}

func (center *Center) ListTasks(ctx context.Context, pageNo, pageSize int32, filter bo.NotificationTaskListFilter) ([]bo.NotificationTask, int, int, int, error) {
	ns, err := center.access.Allow(ctx, perm.MessageLogRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if filter.Status != "" && !validTaskStatus(filter.Status) {
		return nil, 0, 0, 0, merr.Invalid("invalid task status")
	}
	if err := validListTimeRange(filter.StartAt, filter.EndAt); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := center.repo.ListTasks(ctx, ns, filter, p, s)
	return items, total, p, s, center.access.Wrap(err)
}

func (center *Center) RetryTask(ctx context.Context, uid string) (*bo.NotificationTask, error) {
	ns, err := center.access.Allow(ctx, perm.MessageLogWrite)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.RetryTask(ctx, ns, uid)
	return item, center.access.Wrap(err)
}

func (center *Center) CancelTask(ctx context.Context, uid string) (*bo.NotificationTask, error) {
	ns, err := center.access.Allow(ctx, perm.MessageLogWrite)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.CancelTask(ctx, ns, uid)
	return item, center.access.Wrap(err)
}

// Claim 返回 PENDING 任务。只接受名为 merak 的 service key。
func (center *Center) Claim(ctx context.Context) ([]bo.NotifySnapshot, error) {
	if err := requireMerak(ctx); err != nil {
		return nil, err
	}
	return center.repo.ClaimPending(ctx, 20)
}

func (center *Center) MarkSending(ctx context.Context, taskUID string) error {
	if err := requireMerak(ctx); err != nil {
		return err
	}
	return center.repo.MarkSending(ctx, taskUID)
}

func (center *Center) Report(ctx context.Context, report bo.NotifyReport) error {
	if err := requireMerak(ctx); err != nil {
		return err
	}
	return center.repo.ApplyNotifyReport(ctx, report)
}

func validTaskStatus(status string) bool {
	switch status {
	case "PENDING", "SENDING", "SENT", "FAILED", "CANCELLED":
		return true
	default:
		return false
	}
}

func validMessageLogStatus(status string) bool {
	return status == "SENT" || status == "FAILED"
}

// MaxMessageQueryDays 是消息日志与通知任务单次查询允许的最大天数（含首尾）。
// 与前端日期选择器的上限一致，避免一次查询扫过所有归档周表。
const MaxMessageQueryDays = 30

// validListTimeRange 校验列表的时间范围：两端都可省略，EndAt 不包含。
func validListTimeRange(start, end time.Time) error {
	if !start.IsZero() && !end.IsZero() {
		if !start.Before(end) {
			return merr.Invalid("invalid time range")
		}
		if end.Sub(start) > MaxMessageQueryDays*24*time.Hour {
			return merr.Invalid("time range must be within 30 days")
		}
	}
	return nil
}

// BackfillMessageLogs 把主库遗留的消息日志归档到租户分片库。调度器每小时做一小批。
func (center *Center) BackfillMessageLogs(ctx context.Context) (int, error) {
	return center.repo.BackfillMessageLogs(ctx, 200)
}

func requireMerak(ctx context.Context) error {
	name, ok := actor.ServiceName(ctx)
	if !ok || name != "merak" {
		return merr.Forbidden("service key cannot call this api")
	}
	return nil
}

func redactChannel(item *bo.NotificationChannel) *bo.NotificationChannel {
	if item == nil {
		return nil
	}
	out := *item
	out.Password = ""
	out.Secret = ""
	return &out
}
