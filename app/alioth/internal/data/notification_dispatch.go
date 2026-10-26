package data

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationchannel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationoncall"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationoncallgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationsubscriptiongroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtemplate"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/app/alioth/pkg/oncall"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	pkgnotify "github.com/aide-family/Alioth/pkg/notify"
)

// notifyEvent 是分发订阅时需要的告警事件字段。
type notifyEvent struct {
	NamespaceUID  string
	EventUID      string
	RuleUID       string
	DatasourceUID string
	// DatasourceLevelUID 是数据源自身的级别标识，用于按数据源级别过滤订阅和 OnCall。
	DatasourceLevelUID string
	LevelUID           string
	Phase              string
	Summary            string
	Description        string
	Labels             string
	Value              string
	ExternalURL        string
	Fingerprint        string
}

type notifyTarget struct {
	ChannelUID  string
	ChannelType notificationchannel.Type
	TemplateUID string
	Sender      string
	Receivers   string
	Subject     string
	Body        string
}

// notificationSender 返回可供审计展示的实际发送身份，不包含渠道凭据或连接地址。
func notificationSender(channel *ent.NotificationChannel) string {
	if channel.FromAddress != "" {
		return channel.FromAddress
	}
	if channel.Type == notificationchannel.TypeSMTP && channel.Username != "" {
		return channel.Username
	}
	return channel.Name
}

// notificationReceiver 返回可审计的 Webhook 接收端；去除 URL 用户信息、查询参数和 fragment，避免记录密钥。
func notificationReceiver(channel *ent.NotificationChannel) string {
	if channel.Type != notificationchannel.TypeWEBHOOK {
		return ""
	}
	parsed, err := url.Parse(channel.Endpoint)
	if err != nil {
		return channel.Name
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	return parsed.String()
}

// groupRoute 是本次事件要投递的通知组。subscription_uid 非空表示该组由订阅路由产生；
// 两套路由同时命中同一通知组时订阅优先，任务归属保持订阅。
type groupRoute struct {
	GroupUID        string
	SubscriptionUID string
}

// matchedGroupSource 是命中事件的订阅或 OnCall，以及它绑定且在本次分发中生效的通知组。
type matchedGroupSource struct {
	UID       string
	Interval  int
	GroupUIDs []string
}

// dispatchMatch 是合并后的通知组路由和本次事件适用的最长重复通知间隔。
type dispatchMatch struct {
	routes      []groupRoute
	maxInterval time.Duration
}

// dispatchEventNotifications 把命中的订阅和 OnCall 合并成一组通知组后生成任务。
// 两套路由共用一次接收方去重；同一指纹在重复通知间隔内不再产生 FIRING 任务。
// 调用方必须在写入事件的事务内调用，保证事件和任务同时生效。
func dispatchEventNotifications(ctx context.Context, tx *ent.Tx, event notifyEvent) error {
	matched, err := collectNotificationRoutes(ctx, tx, event)
	if err != nil {
		return err
	}
	switch notificationtask.Phase(event.Phase) {
	case notificationtask.PhaseFIRING:
		// 同一指纹在最长间隔内已经发过 FIRING 通知时，整条事件不再生成任务。
		suppressed, err := firingSuppressed(ctx, tx, event, matched.maxInterval)
		if err != nil || suppressed {
			return err
		}
	case notificationtask.PhaseRECOVERED:
		// 被间隔抑制或从未发出的 FIRING 不应在恢复时补发通知。
		fired, err := firingDelivered(ctx, tx, event.NamespaceUID, event.EventUID, event.Fingerprint)
		if err != nil || !fired {
			return err
		}
	}
	if len(matched.routes) == 0 {
		return nil
	}
	tracker, err := newDispatchTracker(ctx, tx, event)
	if err != nil {
		return err
	}
	return createGroupEventTasks(ctx, tx, event, matched.routes, tracker, paramsJSON(event))
}

// collectNotificationRoutes 收集命中的订阅和 OnCall，合并成同一组通知组路由。
// 订阅条件按标签、告警级别、数据源和数据源级别过滤；OnCall 还要求绑定在同一个
// time.Now() 换算到各自时区后处于接收时段。贡献了至少一个生效通知组的来源
// 参与重复通知间隔计算，取其中最大的间隔。
func collectNotificationRoutes(ctx context.Context, tx *ent.Tx, event notifyEvent) (dispatchMatch, error) {
	subscriptions, err := matchedSubscriptions(ctx, tx, event)
	if err != nil {
		return dispatchMatch{}, err
	}
	onCalls, err := matchedOnCalls(ctx, tx, event)
	if err != nil {
		return dispatchMatch{}, err
	}
	merged := make(map[string]groupRoute, len(subscriptions)+len(onCalls))
	order := make([]string, 0, len(subscriptions)+len(onCalls))
	for _, subscription := range subscriptions {
		for _, groupUID := range subscription.GroupUIDs {
			if _, ok := merged[groupUID]; ok {
				continue
			}
			merged[groupUID] = groupRoute{GroupUID: groupUID, SubscriptionUID: subscription.UID}
			order = append(order, groupUID)
		}
	}
	for _, onCall := range onCalls {
		for _, groupUID := range onCall.GroupUIDs {
			if _, ok := merged[groupUID]; ok {
				continue
			}
			merged[groupUID] = groupRoute{GroupUID: groupUID}
			order = append(order, groupUID)
		}
	}
	if len(order) == 0 {
		return dispatchMatch{}, nil
	}
	active, err := activeNotificationGroups(ctx, tx, event.NamespaceUID, order)
	if err != nil {
		return dispatchMatch{}, err
	}
	routes := make([]groupRoute, 0, len(order))
	for _, groupUID := range order {
		if _, ok := active[groupUID]; !ok {
			continue
		}
		routes = append(routes, merged[groupUID])
	}
	sources := make([]matchedGroupSource, 0, len(subscriptions)+len(onCalls))
	sources = append(sources, subscriptions...)
	sources = append(sources, onCalls...)
	maxInterval := 0
	for _, source := range sources {
		if !sourceContributes(source, active) {
			continue
		}
		if source.Interval > maxInterval {
			maxInterval = source.Interval
		}
	}
	return dispatchMatch{routes: routes, maxInterval: time.Duration(maxInterval) * time.Second}, nil
}

// matchedSubscriptions 返回条件命中的 ACTIVE 订阅及其通知组绑定。
func matchedSubscriptions(ctx context.Context, tx *ent.Tx, event notifyEvent) ([]matchedGroupSource, error) {
	rows, err := tx.NotificationSubscription.Query().Where(
		notificationsubscription.NamespaceUIDEQ(event.NamespaceUID),
		notificationsubscription.StatusEQ(notificationsubscription.StatusACTIVE),
		notificationsubscription.DeletedAtIsNil(),
	).Order(ent.Asc(notificationsubscription.FieldUID)).All(ctx)
	if err != nil {
		return nil, err
	}
	matched := make([]matchedGroupSource, 0, len(rows))
	for _, row := range rows {
		if !matchSubscription(row, event) {
			continue
		}
		matched = append(matched, matchedGroupSource{UID: row.UID, Interval: row.RepeatIntervalSeconds})
	}
	if len(matched) == 0 {
		return nil, nil
	}
	uids := make([]string, len(matched))
	for i := range matched {
		uids[i] = matched[i].UID
	}
	links, err := tx.NotificationSubscriptionGroup.Query().Where(
		notificationsubscriptiongroup.NamespaceUIDEQ(event.NamespaceUID),
		notificationsubscriptiongroup.SubscriptionUIDIn(uids...),
		notificationsubscriptiongroup.DeletedAtIsNil(),
	).Order(ent.Asc(notificationsubscriptiongroup.FieldUID)).All(ctx)
	if err != nil {
		return nil, err
	}
	groupsBySubscription := make(map[string][]string, len(matched))
	for _, link := range links {
		if strings.TrimSpace(link.GroupUID) == "" {
			continue
		}
		groupsBySubscription[link.SubscriptionUID] = append(groupsBySubscription[link.SubscriptionUID], link.GroupUID)
	}
	for i := range matched {
		matched[i].GroupUIDs = groupsBySubscription[matched[i].UID]
	}
	return matched, nil
}

// matchedOnCalls 返回条件命中的 ACTIVE OnCall，以及当前时刻处于接收时段的绑定。
// 同一次分发里所有 OnCall 用同一个时刻判断，避免跨分钟边界时前后不一致。
func matchedOnCalls(ctx context.Context, tx *ent.Tx, event notifyEvent) ([]matchedGroupSource, error) {
	rows, err := tx.NotificationOnCall.Query().Where(
		notificationoncall.NamespaceUIDEQ(event.NamespaceUID),
		notificationoncall.StatusEQ(notificationoncall.StatusACTIVE),
		notificationoncall.DeletedAtIsNil(),
	).Order(ent.Asc(notificationoncall.FieldUID)).All(ctx)
	if err != nil {
		return nil, err
	}
	matched := make([]matchedGroupSource, 0, len(rows))
	matchedRows := make([]*ent.NotificationOnCall, 0, len(rows))
	for _, row := range rows {
		if !matchOnCall(row, event) {
			continue
		}
		matched = append(matched, matchedGroupSource{UID: row.UID, Interval: row.RepeatIntervalSeconds})
		matchedRows = append(matchedRows, row)
	}
	if len(matched) == 0 {
		return nil, nil
	}
	uids := make([]string, len(matched))
	for i := range matched {
		uids[i] = matched[i].UID
	}
	links, err := tx.NotificationOnCallGroup.Query().Where(
		notificationoncallgroup.NamespaceUIDEQ(event.NamespaceUID),
		notificationoncallgroup.OncallUIDIn(uids...),
		notificationoncallgroup.DeletedAtIsNil(),
	).Order(ent.Asc(notificationoncallgroup.FieldUID)).All(ctx)
	if err != nil {
		return nil, err
	}
	linksByOnCall := make(map[string][]*ent.NotificationOnCallGroup, len(matched))
	for _, link := range links {
		linksByOnCall[link.OncallUID] = append(linksByOnCall[link.OncallUID], link)
	}
	now := time.Now()
	for i, row := range matchedRows {
		location, err := oncall.LoadLocation(row.Timezone)
		if err != nil {
			// 保存时已经校验时区；历史脏数据只跳过该 OnCall，不能让事件写入失败。
			continue
		}
		at := now.In(location)
		seen := make(map[string]struct{}, len(linksByOnCall[row.UID]))
		for _, link := range linksByOnCall[row.UID] {
			if !onCallWindowActive(link, at) {
				continue
			}
			if strings.TrimSpace(link.GroupUID) == "" {
				continue
			}
			if _, ok := seen[link.GroupUID]; ok {
				continue
			}
			seen[link.GroupUID] = struct{}{}
			matched[i].GroupUIDs = append(matched[i].GroupUIDs, link.GroupUID)
		}
	}
	return matched, nil
}

// activeNotificationGroups 过滤出启用且未删除的通知组。停用的组不投递，也不参与间隔计算。
func activeNotificationGroups(ctx context.Context, tx *ent.Tx, namespaceUID string, groupUIDs []string) (map[string]struct{}, error) {
	rows, err := tx.NotificationGroup.Query().Where(
		notificationgroup.UIDIn(groupUIDs...),
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
		notificationgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, err
	}
	active := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		active[row.UID] = struct{}{}
	}
	return active, nil
}

// sourceContributes 判断一个来源是否真的贡献了生效的通知组。
func sourceContributes(source matchedGroupSource, active map[string]struct{}) bool {
	for _, groupUID := range source.GroupUIDs {
		if _, ok := active[groupUID]; ok {
			return true
		}
	}
	return false
}

// firingSuppressed 判断同一指纹是否在最长重复通知间隔内已经生成过 FIRING 任务。
// 只统计仍在投递链路上的任务；恢复通知不受这个间隔限制。
func firingSuppressed(ctx context.Context, tx *ent.Tx, event notifyEvent, interval time.Duration) (bool, error) {
	if interval <= 0 || strings.TrimSpace(event.Fingerprint) == "" {
		return false, nil
	}
	return tx.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(event.NamespaceUID),
		notificationtask.FingerprintEQ(event.Fingerprint),
		notificationtask.PhaseEQ(notificationtask.PhaseFIRING),
		notificationtask.StatusIn(notificationtask.StatusPENDING, notificationtask.StatusSENDING, notificationtask.StatusSENT),
		notificationtask.CreatedAtGT(time.Now().Add(-interval)),
	).Exist(ctx)
}

// firingDelivered 判断同一指纹是否存在仍在投递链路上的 FIRING 任务。恢复通知只针对
// 真正发出过（或仍在队列里）的 FIRING；被间隔抑制或彻底失败的告警恢复时不补发通知。
// 升级前写入的任务没有 fingerprint，再按事件标识兜底，避免老事件恢复时漏发。
func firingDelivered(ctx context.Context, tx *ent.Tx, namespaceUID, eventUID, fingerprint string) (bool, error) {
	if strings.TrimSpace(fingerprint) != "" {
		found, err := tx.NotificationTask.Query().Where(
			notificationtask.NamespaceUIDEQ(namespaceUID),
			notificationtask.FingerprintEQ(fingerprint),
			notificationtask.PhaseEQ(notificationtask.PhaseFIRING),
			notificationtask.StatusIn(notificationtask.StatusPENDING, notificationtask.StatusSENDING, notificationtask.StatusSENT),
		).Exist(ctx)
		if err != nil || found {
			return found, err
		}
	}
	if strings.TrimSpace(eventUID) == "" {
		return false, nil
	}
	return tx.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.EventUIDEQ(eventUID),
		notificationtask.PhaseEQ(notificationtask.PhaseFIRING),
		notificationtask.StatusIn(notificationtask.StatusPENDING, notificationtask.StatusSENDING, notificationtask.StatusSENT),
	).Exist(ctx)
}

// dispatchTracker 记录本次事件已经安排过的接收方。同一事件同时命中订阅和 OnCall、
// 或命中多个通知组时，同一个联系人只投递一次，同一个 Webhook 渠道只安排一次。
type dispatchTracker struct {
	contacts          map[string]struct{}
	scheduledChannels map[string]struct{}
}

// newDispatchTracker 用同一事件、同一阶段已有的任务初始化去重状态，
// 覆盖事务内重试或上一轮只写入部分任务的场景。
func newDispatchTracker(ctx context.Context, tx *ent.Tx, event notifyEvent) (*dispatchTracker, error) {
	tracker := &dispatchTracker{
		contacts:          map[string]struct{}{},
		scheduledChannels: map[string]struct{}{},
	}
	if event.EventUID == "" {
		return tracker, nil
	}
	rows, err := tx.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(event.EventUID),
		notificationtask.PhaseEQ(notificationtask.Phase(event.Phase)),
	).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		tracker.scheduledChannels[row.ChannelUID] = struct{}{}
		for _, contact := range pkgnotify.SplitReceivers(row.Receivers) {
			tracker.contacts[normalizeContact(contact)] = struct{}{}
		}
	}
	return tracker, nil
}

// pendingTarget 返回本次仍要投递的目标。SMTP/SMS 去掉本次事件已经投递过的联系人，
// 没有剩余接收方时整个目标跳过；Webhook 按渠道去重。
func (tracker *dispatchTracker) pendingTarget(target notifyTarget) (string, bool) {
	if target.ChannelType != notificationchannel.TypeSMTP && target.ChannelType != notificationchannel.TypeSMS {
		if _, ok := tracker.scheduledChannels[target.ChannelUID]; ok {
			return "", false
		}
		return target.Receivers, true
	}
	contacts := pkgnotify.SplitReceivers(target.Receivers)
	remaining := make([]string, 0, len(contacts))
	for _, contact := range contacts {
		if tracker.contactDelivered(contact) {
			continue
		}
		remaining = append(remaining, contact)
	}
	if len(remaining) == 0 {
		return "", false
	}
	return strings.Join(remaining, ","), true
}

// markScheduled 在任务写入后登记接收方，供后续通知组去重。
func (tracker *dispatchTracker) markScheduled(target notifyTarget, receivers string) {
	tracker.scheduledChannels[target.ChannelUID] = struct{}{}
	if target.ChannelType != notificationchannel.TypeSMTP && target.ChannelType != notificationchannel.TypeSMS {
		return
	}
	for _, contact := range pkgnotify.SplitReceivers(receivers) {
		tracker.contacts[normalizeContact(contact)] = struct{}{}
	}
}

// normalizeContact 归一化一个联系人：邮箱统一小写，手机号去掉空格、横线和括号等
// 格式字符只保留数字，避免同一号码的不同写法被当成两个接收方。
func normalizeContact(contact string) string {
	value := strings.ToLower(strings.TrimSpace(contact))
	if strings.Contains(value, "@") {
		return value
	}
	var digits strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	if digits.Len() == 0 {
		return value
	}
	return digits.String()
}

// contactDelivered 判断联系人是否已经投递过。手机号可能带或不带国家码，
// 归一化后相同，或较短一方是较长一方的后缀（至少 8 位）时视为同一联系人。
func (tracker *dispatchTracker) contactDelivered(contact string) bool {
	key := normalizeContact(contact)
	if _, ok := tracker.contacts[key]; ok {
		return true
	}
	if strings.Contains(key, "@") || len(key) < 8 {
		return false
	}
	for known := range tracker.contacts {
		if strings.Contains(known, "@") || len(known) < 8 {
			continue
		}
		if strings.HasSuffix(known, key) || strings.HasSuffix(key, known) {
			return true
		}
	}
	return false
}

// createGroupEventTasks 把命中的通知组展开成渠道目标并写入 PENDING 任务。
// 路由里 subscription_uid 为空表示该组只由 OnCall 命中。
func createGroupEventTasks(ctx context.Context, tx *ent.Tx, event notifyEvent, routes []groupRoute, tracker *dispatchTracker, params string) error {
	groupUIDs := make([]string, len(routes))
	subscriptionUIDs := make(map[string]string, len(routes))
	for i, route := range routes {
		groupUIDs[i] = route.GroupUID
		if route.SubscriptionUID != "" {
			subscriptionUIDs[route.GroupUID] = route.SubscriptionUID
		}
	}
	groups, err := tx.NotificationGroup.Query().Where(
		notificationgroup.UIDIn(groupUIDs...),
		notificationgroup.NamespaceUIDEQ(event.NamespaceUID),
		notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
		notificationgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return err
	}
	groupsByUID := make(map[string]*ent.NotificationGroup, len(groups))
	for _, group := range groups {
		groupsByUID[group.UID] = group
	}
	// 按路由顺序处理：订阅命中的通知组排在 OnCall 之前，接收方重叠时订阅的
	// 渠道和模板先占用联系人，与 subscription_uid 优先的归属保持一致。
	for _, route := range routes {
		group, ok := groupsByUID[route.GroupUID]
		if !ok {
			// 停用或已删除的通知组不投递。
			continue
		}
		targets, err := resolveGroupTargets(ctx, tx, group)
		if err != nil {
			return err
		}
		for _, target := range targets {
			// 同一通知组可能被两套路由命中：先查本事务里已有的任务再写入。
			// PostgreSQL 里唯一约束冲突会让整个事件事务失效，不能靠捕获错误继续。
			exists, err := eventTaskExists(ctx, tx, event, group.UID, target.ChannelUID)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
			receivers, ok := tracker.pendingTarget(target)
			if !ok {
				continue
			}
			err = tx.NotificationTask.Create().
				SetUID(id.New()).
				SetNamespaceUID(event.NamespaceUID).
				SetEventUID(event.EventUID).
				SetPhase(notificationtask.Phase(event.Phase)).
				SetStatus(notificationtask.StatusPENDING).
				SetRuleUID(event.RuleUID).
				SetFingerprint(event.Fingerprint).
				SetSubscriptionUID(subscriptionUIDs[group.UID]).
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
				Exec(ctx)
			if err != nil {
				// (event_uid, phase, group_uid, channel_uid) 唯一约束是最后一道兜底。
				if ent.IsConstraintError(err) {
					continue
				}
				return err
			}
			tracker.markScheduled(target, receivers)
		}
	}
	return nil
}

// eventTaskExists 判断同一事件、阶段、通知组和渠道的任务是否已经存在。
// 试发事件的 event_uid 为空，不占用唯一约束。
func eventTaskExists(ctx context.Context, tx *ent.Tx, event notifyEvent, groupUID, channelUID string) (bool, error) {
	if event.EventUID == "" {
		return false, nil
	}
	return tx.NotificationTask.Query().Where(
		notificationtask.EventUIDEQ(event.EventUID),
		notificationtask.PhaseEQ(notificationtask.Phase(event.Phase)),
		notificationtask.GroupUIDEQ(groupUID),
		notificationtask.ChannelUIDEQ(channelUID),
	).Exist(ctx)
}

// matchOnCall 判断事件是否满足 OnCall 的标签、告警级别、数据源和数据源级别条件。
// 四个条件之间是 AND，空条件表示不限制。
func matchOnCall(onCall *ent.NotificationOnCall, event notifyEvent) bool {
	if len(onCall.LevelUids) > 0 && !containsString(onCall.LevelUids, event.LevelUID) {
		return false
	}
	if len(onCall.DatasourceUids) > 0 && !containsString(onCall.DatasourceUids, event.DatasourceUID) {
		return false
	}
	if len(onCall.DatasourceLevelUids) > 0 && !containsString(onCall.DatasourceLevelUids, event.DatasourceLevelUID) {
		return false
	}
	return labelsMatch(onCall.Labels, event.Labels)
}

// onCallWindowActive 判断绑定在 at 时刻是否处于接收时段。空时段表示全天候接收。
func onCallWindowActive(link *ent.NotificationOnCallGroup, at time.Time) bool {
	specs, err := oncall.Unmarshal(link.TimeWindows)
	if err != nil {
		return false
	}
	windows, err := oncall.Compile(specs)
	if err != nil {
		return false
	}
	return oncall.Active(windows, at)
}

// matchSubscription 判断事件是否满足订阅的标签、告警级别、数据源和数据源级别条件。
// 四个条件之间是 AND，空条件表示不限制。
func matchSubscription(subscription *ent.NotificationSubscription, event notifyEvent) bool {
	if len(subscription.LevelUids) > 0 && !containsString(subscription.LevelUids, event.LevelUID) {
		return false
	}
	if len(subscription.DatasourceUids) > 0 && !containsString(subscription.DatasourceUids, event.DatasourceUID) {
		return false
	}
	if len(subscription.DatasourceLevelUids) > 0 && !containsString(subscription.DatasourceLevelUids, event.DatasourceLevelUID) {
		return false
	}
	return labelsMatch(subscription.Labels, event.Labels)
}

func labelsMatch(selector, eventLabels string) bool {
	matched, err := labelsutil.MatchersJSONContained(selector, eventLabels)
	if err != nil {
		return false
	}
	return matched
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

// resolveGroupTargets 把一个通知组展开成实际投递目标：主渠道一次，每个选中的 Webhook 各一次。
func resolveGroupTargets(ctx context.Context, tx *ent.Tx, group *ent.NotificationGroup) ([]notifyTarget, error) {
	channelUIDs, templateUIDs := groupChannelTargets(group)
	targets := make([]notifyTarget, 0, len(channelUIDs))
	seen := map[string]struct{}{}
	for index, channelUID := range channelUIDs {
		if _, ok := seen[channelUID]; ok {
			continue
		}
		if strings.TrimSpace(channelUID) == "" {
			continue
		}
		templateUID := ""
		if index < len(templateUIDs) {
			templateUID = strings.TrimSpace(templateUIDs[index])
		}
		if templateUID == "" {
			continue
		}
		channel, err := tx.NotificationChannel.Query().Where(
			notificationchannel.UIDEQ(channelUID),
			notificationchannel.NamespaceUIDEQ(group.NamespaceUID),
			notificationchannel.DeletedAtIsNil(),
			notificationchannel.StatusEQ(notificationchannel.StatusACTIVE),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		template, err := tx.NotificationTemplate.Query().Where(
			notificationtemplate.UIDEQ(templateUID),
			notificationtemplate.NamespaceUIDEQ(group.NamespaceUID),
			notificationtemplate.DeletedAtIsNil(),
			notificationtemplate.StatusEQ(notificationtemplate.StatusACTIVE),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		if !templateCoversChannel(channel, template) {
			continue
		}
		receivers := ""
		switch channel.Type {
		case notificationchannel.TypeSMTP:
			receivers, err = resolveGroupReceivers(ctx, tx, group, channel.Type)
			if err != nil {
				return nil, err
			}
			if receivers == "" {
				continue
			}
		case notificationchannel.TypeSMS:
			receivers, err = resolveGroupReceivers(ctx, tx, group, channel.Type)
			if err != nil {
				return nil, err
			}
			if receivers == "" {
				continue
			}
		default:
			receivers = notificationReceiver(channel)
		}
		seen[channelUID] = struct{}{}
		targets = append(targets, notifyTarget{
			ChannelUID: channel.UID, ChannelType: channel.Type, TemplateUID: template.UID,
			Sender: notificationSender(channel), Receivers: receivers,
			Subject: template.Subject, Body: template.Body,
		})
	}
	return targets, nil
}

// groupChannelTargets 返回渠道和模板的一一对应列表。历史通知组只有旧版单渠道字段，
// 这里按旧字段展开，升级后无需手工重建通知组。
func groupChannelTargets(group *ent.NotificationGroup) ([]string, []string) {
	if len(group.ChannelUids) > 0 {
		return group.ChannelUids, group.TemplateUids
	}
	if strings.TrimSpace(group.ChannelUID) == "" {
		return nil, nil
	}
	channelUIDs := []string{group.ChannelUID}
	templateUIDs := []string{group.TemplateUID}
	for range group.WebhookChannelUids {
		channelUIDs = append(channelUIDs, "")
		templateUIDs = append(templateUIDs, group.WebhookTemplateUID)
	}
	copy(channelUIDs[1:], group.WebhookChannelUids)
	return channelUIDs, templateUIDs
}

// templateCoversChannel 判断模板是否能覆盖渠道。OTHER Webhook 模板可以覆盖任意应用。
func templateCoversChannel(channel *ent.NotificationChannel, template *ent.NotificationTemplate) bool {
	switch channel.Type {
	case notificationchannel.TypeSMTP:
		return template.Type == notificationtemplate.TypeEMAIL
	case notificationchannel.TypeSMS:
		return template.Type == notificationtemplate.TypeSMS_ALICLOUD
	case notificationchannel.TypeWEBHOOK:
		return template.Type == notificationtemplate.TypeWEBHOOK &&
			(template.App == notificationtemplate.AppOTHER || string(template.App) == string(channel.App))
	default:
		return false
	}
}

// resolveGroupReceivers 按渠道类型合并手填接收方和成员资料。
// 邮件取成员邮箱，短信取成员手机号；同一个人只保留一次。
func resolveGroupReceivers(ctx context.Context, tx *ent.Tx, group *ent.NotificationGroup, kind notificationchannel.Type) (string, error) {
	base := group.EmailReceivers
	memberUIDs := group.EmailMemberUids
	field := "email"
	if kind == notificationchannel.TypeSMS {
		base = group.SmsReceivers
		memberUIDs = group.SmsMemberUids
		field = "phone"
	}
	// 历史数据没有新字段时回退到旧版邮件接收方和成员。
	if kind == notificationchannel.TypeSMTP {
		if base == "" {
			base = group.Receivers
		}
		if len(memberUIDs) == 0 {
			memberUIDs = group.MemberUids
		}
	}
	values := pkgnotify.SplitReceivers(base)
	if len(memberUIDs) > 0 {
		members, err := tx.Member.Query().Where(
			member.UIDIn(memberUIDs...),
			member.NamespaceUIDEQ(group.NamespaceUID),
			member.StatusEQ(member.StatusJOINED),
			member.DeletedAtIsNil(),
		).All(ctx)
		if err != nil {
			return "", err
		}
		userUIDs := make([]string, 0, len(members))
		for _, row := range members {
			userUIDs = append(userUIDs, row.UserUID)
		}
		if len(userUIDs) > 0 {
			users, err := tx.User.Query().Where(
				user.UIDIn(userUIDs...),
				user.StatusEQ(user.StatusACTIVE),
			).All(ctx)
			if err != nil {
				return "", err
			}
			for _, account := range users {
				value := account.Email
				if field == "phone" {
					value = account.Phone
				}
				if strings.TrimSpace(value) != "" {
					values = append(values, value)
				}
			}
		}
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return strings.Join(out, ","), nil
}
