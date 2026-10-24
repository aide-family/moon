package alertevent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/lock"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	// EvalInterval 是调度周期，与调度锁的 TTL 相同。
	EvalInterval = 15 * time.Second
	// ScheduleLockKey 是到期扫描使用的锁。玉衡不持有评测锁。
	ScheduleLockKey = "alioth:lock:schedule:eval"
	// tabMineID 是实时告警页固定的「我的告警」Tab 标识。
	tabMineID = "mine"
	// SuppressDefaultSeconds 是调用方没有传抑制时长时使用的默认值，保持旧接口可用。
	SuppressDefaultSeconds int64 = 30 * 60
	// SuppressMinSeconds 和 SuppressMaxSeconds 是允许的抑制时长范围，30 天。
	SuppressMinSeconds int64 = 60
	SuppressMaxSeconds int64 = 30 * 24 * 60 * 60
)

var ProviderSet = wire.NewSet(NewCoordinator)

// Coordinator 调度规则、领取快照、回写事件，并执行人工处理。
type Coordinator struct {
	rules       repository.AlertRuleRepo
	events      repository.AlertEventRepo
	pages       repository.AlertPageRepo
	users       repository.UserRepo
	members     repository.MemberRepo
	datasources repository.DatasourceRepo
	locks       *lock.Client
	access      access.Checker
	log         *slog.Logger
}

func NewCoordinator(
	rules repository.AlertRuleRepo,
	events repository.AlertEventRepo,
	pages repository.AlertPageRepo,
	users repository.UserRepo,
	members repository.MemberRepo,
	datasources repository.DatasourceRepo,
	locks *lock.Client,
	checker access.Checker,
	logger *slog.Logger,
) *Coordinator {
	return &Coordinator{
		rules: rules, events: events, pages: pages, users: users, members: members,
		datasources: datasources, locks: locks, access: checker, log: logger,
	}
}

// Schedule 在锁内把到期规则的下一次评测时间推迟一个周期，并记下本轮批次。拿到锁后不续期。
func (coordinator *Coordinator) Schedule(ctx context.Context) error {
	if coordinator.locks == nil {
		return merr.Server("lock failed")
	}
	hold, err := coordinator.locks.Acquire(ctx, ScheduleLockKey, EvalInterval)
	if err != nil {
		if errors.Is(err, lock.ErrNotAcquired) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()
	now := time.Now().UTC().Truncate(time.Second)
	if err := coordinator.rules.ScheduleDue(ctx, now, now.Add(EvalInterval)); err != nil {
		coordinator.log.Error(fmt.Sprintf("schedule eval: %v", err))
		return err
	}
	return nil
}

// Claim 返回已调度且尚未回写的规则快照。按匹配数据源展开，只接受名为 megrez 的 service key。
func (coordinator *Coordinator) Claim(ctx context.Context) ([]bo.RuleSnapshot, error) {
	if err := coordinator.requireMegrez(ctx); err != nil {
		return nil, err
	}
	rules, err := coordinator.rules.ListScheduled(ctx)
	if err != nil {
		return nil, coordinator.access.Wrap(err)
	}
	snapshots := make([]bo.RuleSnapshot, 0, len(rules))
	for _, rule := range rules {
		if rule.ScheduledSlot == nil {
			continue
		}
		sources, err := coordinator.datasources.ListDatasourcesByMatchers(ctx, rule.NamespaceUID, rule.Driver, rule.DatasourceLabels)
		if err != nil {
			return nil, coordinator.access.Wrap(err)
		}
		if len(sources) == 0 {
			continue
		}
		for _, source := range sources {
			recorded, err := coordinator.events.HasBatch(ctx, rule.UID, source.UID, *rule.ScheduledSlot)
			if err != nil {
				return nil, err
			}
			if recorded {
				continue
			}
			copyRule := rule
			copyRule.DatasourceUID = source.UID
			copyRule.Driver = source.Driver
			fingerprints, err := coordinator.events.ListFiringFingerprints(ctx, rule.NamespaceUID, rule.UID, source.UID)
			if err != nil {
				return nil, err
			}
			snapshots = append(snapshots, bo.RuleSnapshot{Rule: copyRule, Datasource: source, FiringFingerprints: fingerprints})
		}
	}
	return snapshots, nil
}

// Report 按 fingerprint 幂等回写。重复批次直接返回。全部匹配数据源回写后清除 scheduled_slot。
func (coordinator *Coordinator) Report(ctx context.Context, report bo.EvalReport) error {
	if err := coordinator.requireMegrez(ctx); err != nil {
		return err
	}
	if report.RuleUID == "" || report.EvalSlot.IsZero() || report.DatasourceUID == "" {
		return merr.Invalid("invalid request")
	}
	rule, err := coordinator.rules.GetAlertRuleByUID(ctx, report.RuleUID)
	if err != nil {
		return coordinator.access.Wrap(err)
	}
	report.RuleUID = rule.UID
	rule.DatasourceUID = report.DatasourceUID
	if err := coordinator.events.ApplyReport(ctx, *rule, report); err != nil {
		return coordinator.access.Wrap(err)
	}
	sources, err := coordinator.datasources.ListDatasourcesByMatchers(ctx, rule.NamespaceUID, rule.Driver, rule.DatasourceLabels)
	if err != nil {
		return coordinator.access.Wrap(err)
	}
	uids := make([]string, 0, len(sources))
	for _, source := range sources {
		uids = append(uids, source.UID)
	}
	done, err := coordinator.events.AllBatchesRecorded(ctx, rule.UID, uids, report.EvalSlot)
	if err != nil {
		return coordinator.access.Wrap(err)
	}
	if done {
		if err := coordinator.rules.ClearScheduledSlot(ctx, rule.UID, report.EvalSlot); err != nil {
			return coordinator.access.Wrap(err)
		}
	}
	return nil
}

type ListQuery struct {
	View           string
	Name           string
	Mine           bool
	PageUID        string
	LevelUID       string
	RuleName       string
	DatasourceName string
	StartsAtFrom   *time.Time
	StartsAtTo     *time.Time
	EventStatus    string
	Suppression    string
}

func (coordinator *Coordinator) List(ctx context.Context, query ListQuery, pageNo, pageSize int32) ([]bo.AlertEvent, int, int, int, error) {
	ns, err := coordinator.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if query.View == "" {
		query.View = "REALTIME"
	}
	if query.View != "REALTIME" && query.View != "HISTORY" {
		return nil, 0, 0, 0, merr.Invalid("invalid request")
	}
	if query.Mine && query.PageUID != "" {
		return nil, 0, 0, 0, merr.Invalid("invalid request")
	}
	if (query.Mine || query.PageUID != "") && query.View != "REALTIME" {
		return nil, 0, 0, 0, merr.Invalid("invalid request")
	}
	if err := validateHistoryListQuery(query); err != nil {
		return nil, 0, 0, 0, err
	}

	filter := bo.AlertEventListFilter{
		View: query.View, Name: query.Name, LevelUID: query.LevelUID,
		RuleName: query.RuleName, DatasourceName: query.DatasourceName,
		StartsAtFrom: query.StartsAtFrom, StartsAtTo: query.StartsAtTo,
		EventStatus: query.EventStatus, Suppression: query.Suppression,
	}
	if query.Mine {
		eventUIDs, err := coordinator.eventUIDsForCurrentUser(ctx, ns)
		if err != nil {
			return nil, 0, 0, 0, err
		}
		filter.FilterByEvents = true
		filter.EventUIDs = eventUIDs
	}
	if query.PageUID != "" {
		userUID, ok := actor.UserID(ctx)
		if !ok {
			return nil, 0, 0, 0, merr.Unauthenticated("unauthorized")
		}
		pageItem, err := coordinator.pages.GetAlertPage(ctx, ns, query.PageUID)
		if err != nil {
			return nil, 0, 0, 0, coordinator.access.Wrap(err)
		}
		if pageItem.OwnerUID != userUID && !pageItem.Public {
			return nil, 0, 0, 0, merr.Forbidden("forbidden")
		}
		filter.LabelsJSON = pageItem.Labels
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := coordinator.events.ListEvents(ctx, ns, filter, p, s)
	if err != nil {
		return nil, 0, 0, 0, coordinator.access.Wrap(err)
	}
	return items, total, p, s, nil
}

func validateHistoryListQuery(query ListQuery) error {
	hasHistoryOnlyFilters := query.RuleName != "" || query.DatasourceName != "" ||
		query.StartsAtFrom != nil || query.StartsAtTo != nil || query.EventStatus != "" || query.Suppression != ""
	if query.View != "HISTORY" && hasHistoryOnlyFilters {
		return merr.Invalid("history filters require HISTORY view")
	}
	if query.StartsAtFrom != nil && query.StartsAtTo != nil && query.StartsAtFrom.After(*query.StartsAtTo) {
		return merr.Invalid("invalid alert event time range")
	}
	switch query.EventStatus {
	case "", "FIRING", "RECOVERED", "RECOVERED_BY_MANUAL":
	default:
		return merr.Invalid("invalid alert event status")
	}
	switch query.Suppression {
	case "", "SUPPRESSED", "UNSUPPRESSED":
	default:
		return merr.Invalid("invalid suppression filter")
	}
	return nil
}

// eventUIDsForCurrentUser 返回当前用户作为通知接收人的实时事件标识。
// 不能按规则放行：同一规则下只有命中了用户订阅的事件才会生成任务，
// 「我的告警」必须按实际任务的事件标识过滤，否则会把同规则下发给别人的事件也显示出来。
func (coordinator *Coordinator) eventUIDsForCurrentUser(ctx context.Context, namespaceUID string) ([]string, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("unauthorized")
	}
	tokens := []string{userUID}
	account, err := coordinator.users.UserByUID(ctx, userUID)
	if err != nil {
		return nil, coordinator.access.Wrap(err)
	}
	if account.Email != "" {
		tokens = append(tokens, account.Email)
	}
	if account.Phone != "" {
		tokens = append(tokens, account.Phone)
	}
	memberUID := ""
	member, err := coordinator.members.MemberByUser(ctx, namespaceUID, userUID)
	if err == nil {
		memberUID = member.UID
	} else if !merr.IsNotFound(err) {
		return nil, coordinator.access.Wrap(err)
	}
	uids, err := coordinator.events.ListEventUIDsForUser(ctx, namespaceUID, memberUID, tokens)
	if err != nil {
		return nil, coordinator.access.Wrap(err)
	}
	return uids, nil
}

// Chart 用事件对应的数据源和规则表达式查询事件开始至今的原始区间数据。
// 图表是告警详情的只读部分，按告警读权限放行，不要求额外的数据源权限。
func (coordinator *Coordinator) Chart(ctx context.Context, uid string) (*bo.AlertEventChart, error) {
	ns, err := coordinator.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	event, err := coordinator.events.GetEvent(ctx, ns, uid)
	if err != nil {
		return nil, coordinator.access.Wrap(err)
	}
	if event.DatasourceUID == "" || strings.TrimSpace(event.Expr) == "" {
		return nil, merr.Invalid("invalid request")
	}
	source, err := coordinator.datasources.GetDatasource(ctx, ns, event.DatasourceUID)
	if err != nil {
		return nil, coordinator.access.Wrap(err)
	}
	if source.Type != "METRICS" && source.Driver != "LOGS_LOKI" && source.Driver != "LOGS_VLOGS" {
		return nil, merr.Invalid("datasource does not support alert charts")
	}
	start, end, step := chartWindow(event.StartsAt, time.Now())
	body, err := coordinator.datasources.RangeQuery(ctx, *source, event.Expr, start, end, step)
	if err != nil {
		return nil, coordinator.access.Wrap(err)
	}
	return &bo.AlertEventChart{
		DatasourceName: source.Name, Expr: event.Expr,
		Start: start, End: end, Step: step, Body: body,
	}, nil
}

// chartWindow 把事件开始时间到当前时间切成最多 300 个点，步长向上取整到 15 秒，
// 既保持曲线细节，又避免长事件拉取过多样本。
func chartWindow(startsAt, now time.Time) (int64, int64, int64) {
	start := startsAt.UTC().Unix()
	end := now.UTC().Unix()
	if end <= start {
		end = start + 1
	}
	const (
		minStep   = int64(15)
		maxPoints = int64(300)
	)
	step := minStep
	if span := end - start; span > maxPoints*minStep {
		step = (span + maxPoints - 1) / maxPoints
		step = ((step + minStep - 1) / minStep) * minStep
	}
	return start, end, step
}

func (coordinator *Coordinator) Get(ctx context.Context, uid string) (*bo.AlertEvent, error) {
	ns, err := coordinator.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return nil, err
	}
	item, err := coordinator.events.GetEvent(ctx, ns, uid)
	return item, coordinator.access.Wrap(err)
}

func (coordinator *Coordinator) Stats(ctx context.Context) (bo.AlertStats, error) {
	ns, err := coordinator.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return bo.AlertStats{}, err
	}
	stats, err := coordinator.events.Stats(ctx, ns)
	return stats, coordinator.access.Wrap(err)
}

// RealtimeStats 返回实时告警页一次刷新所需的全部数量：已选 Tab 的数量和全局数量。
// 只统计当前用户已选择显示的页面，未选择的页面不参与查询；这里只做 count，不读列表数据。
func (coordinator *Coordinator) RealtimeStats(ctx context.Context) (bo.RealtimeAlertStats, error) {
	ns, err := coordinator.access.Allow(ctx, perm.AlertRead)
	if err != nil {
		return bo.RealtimeAlertStats{}, err
	}
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return bo.RealtimeAlertStats{}, merr.Unauthenticated("unauthorized")
	}
	pages, err := coordinator.pages.ListUserAlertPages(ctx, ns, userUID)
	if err != nil {
		return bo.RealtimeAlertStats{}, coordinator.access.Wrap(err)
	}
	stats, err := coordinator.events.Stats(ctx, ns)
	if err != nil {
		return bo.RealtimeAlertStats{}, coordinator.access.Wrap(err)
	}
	out := bo.RealtimeAlertStats{
		Tabs:       make([]bo.RealtimeAlertTabCount, 0, len(pages)+1),
		Firing:     stats.Firing,
		Suppressed: stats.Suppressed,
		Recovered:  stats.Recovered,
	}

	eventUIDs, err := coordinator.eventUIDsForCurrentUser(ctx, ns)
	if err != nil {
		return bo.RealtimeAlertStats{}, err
	}
	mineCount, err := coordinator.events.CountEvents(ctx, ns, bo.AlertEventListFilter{
		View:           "REALTIME",
		FilterByEvents: true,
		EventUIDs:      eventUIDs,
	})
	if err != nil {
		return bo.RealtimeAlertStats{}, coordinator.access.Wrap(err)
	}
	out.Tabs = append(out.Tabs, bo.RealtimeAlertTabCount{ID: tabMineID, Count: mineCount})
	for _, item := range pages {
		count, err := coordinator.events.CountEvents(ctx, ns, bo.AlertEventListFilter{
			View:       "REALTIME",
			LabelsJSON: item.Labels,
		})
		if err != nil {
			return bo.RealtimeAlertStats{}, coordinator.access.Wrap(err)
		}
		out.Tabs = append(out.Tabs, bo.RealtimeAlertTabCount{ID: item.UID, Count: count})
	}
	return out, nil
}

func (coordinator *Coordinator) Intervene(ctx context.Context, uids []string) error {
	ns, err := coordinator.access.Allow(ctx, perm.AlertIntervene)
	if err != nil {
		return err
	}
	userUID, _ := actor.UserID(ctx)
	return coordinator.access.Wrap(coordinator.events.Intervene(ctx, ns, uids, userUID))
}

// Suppress 抑制一条实时告警，并返回抑制到期时间。permanent 为真时不设到期时间。
// 时长为 0 时使用默认值，兼容还没传时长的旧客户端。
func (coordinator *Coordinator) Suppress(ctx context.Context, uid string, seconds int64, permanent bool) (time.Time, error) {
	ns, err := coordinator.access.Allow(ctx, perm.AlertIntervene)
	if err != nil {
		return time.Time{}, err
	}
	if permanent {
		return time.Time{}, coordinator.access.Wrap(coordinator.events.Suppress(ctx, ns, uid, nil))
	}
	if seconds == 0 {
		seconds = SuppressDefaultSeconds
	}
	if seconds < SuppressMinSeconds || seconds > SuppressMaxSeconds {
		return time.Time{}, merr.Invalid("invalid request")
	}
	until := time.Now().UTC().Add(time.Duration(seconds) * time.Second)
	return until, coordinator.access.Wrap(coordinator.events.Suppress(ctx, ns, uid, &until))
}

func (coordinator *Coordinator) Recover(ctx context.Context, uids []string) error {
	ns, err := coordinator.access.Allow(ctx, perm.AlertIntervene)
	if err != nil {
		return err
	}
	return coordinator.access.Wrap(coordinator.events.RecoverManual(ctx, ns, uids))
}

// BackfillHistory 把升级前留在主库的已恢复事件迁移到分片。调度器每小时做一小批，
// 写入失败时保留主库行，下一次继续重试。
func (coordinator *Coordinator) BackfillHistory(ctx context.Context) (int, error) {
	return coordinator.events.BackfillHistory(ctx, 200)
}

func (coordinator *Coordinator) requireMegrez(ctx context.Context) error {
	name, ok := actor.ServiceName(ctx)
	if !ok || name != "megrez" {
		return merr.Forbidden("service key is required")
	}
	return nil
}
