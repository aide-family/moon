package data

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/agent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertpage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrulegroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/inspectionplan"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncall"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetargetgroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/lock"
)

// labelCacheTTL 控制标签集合在 Redis 中的存活时间；标签写入会主动失效。
const labelCacheTTL = 10 * time.Minute

// labelCache 把租户内各配置对象的标签汇总成 key -> values，供前端标签匹配
// 下拉选择。价值不在缓存本身，而是让前端看到全租户的标签集合，而不是只看
// 当前已加载的数据源/目标组。
type labelCache struct {
	rdb    *redis.Client
	client *ent.Client
	log    *slog.Logger
}

func newLabelCache(c *conf.Data, client *ent.Client, logger *slog.Logger) (*labelCache, func()) {
	cache := &labelCache{client: client, log: logger}
	if c == nil || c.Redis == nil || c.Redis.Addr == "" {
		return cache, func() {}
	}
	cache.rdb = lock.NewRedisClient(c.Redis.Addr, c.Redis.Password)
	cleanup := func() {
		if err := cache.rdb.Close(); err != nil {
			logger.Error("close label cache redis: " + err.Error())
		}
	}
	return cache, cleanup
}

func labelCacheKey(namespaceUID string) string {
	return "alioth:labels:" + namespaceUID
}

// Invalidate 在标签写入后删除缓存；下一次读取会重新汇总全部标签来源。
func (cache *labelCache) Invalidate(ctx context.Context, namespaceUID string) {
	if cache == nil || cache.rdb == nil || namespaceUID == "" {
		return
	}
	if err := cache.rdb.Del(ctx, labelCacheKey(namespaceUID)).Err(); err != nil {
		cache.log.Warn("invalidate label cache: " + err.Error())
	}
}

// Options 返回租户标签集合。Redis 命中时直接用缓存，否则从各配置表重建。
func (cache *labelCache) Options(ctx context.Context, namespaceUID string) ([]bo.LabelOption, error) {
	if cache == nil {
		return []bo.LabelOption{}, nil
	}
	key := labelCacheKey(namespaceUID)
	if cache.rdb != nil && namespaceUID != "" {
		raw, err := cache.rdb.Get(ctx, key).Bytes()
		if err == nil {
			var cached []bo.LabelOption
			if json.Unmarshal(raw, &cached) == nil {
				return cached, nil
			}
		}
	}
	options, err := cache.build(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	if cache.rdb != nil && namespaceUID != "" {
		if raw, err := json.Marshal(options); err == nil {
			if err := cache.rdb.Set(ctx, key, raw, labelCacheTTL).Err(); err != nil {
				cache.log.Warn("cache label options: " + err.Error())
			}
		}
	}
	return options, nil
}

func (cache *labelCache) build(ctx context.Context, namespaceUID string) ([]bo.LabelOption, error) {
	index := make(map[string]map[string]struct{})
	add := func(raw string) {
		values, err := labels.ParseObject(raw)
		if err != nil {
			return
		}
		for key, value := range values {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			if index[key] == nil {
				index[key] = make(map[string]struct{})
			}
			index[key][value] = struct{}{}
		}
	}
	type source struct {
		rows []string
		err  error
	}
	sources := make([]source, 0, 9)
	collect := func(rows []string, err error) {
		sources = append(sources, source{rows: rows, err: err})
	}
	{
		rows, err := cache.client.Datasource.Query().
			Where(datasource.NamespaceUIDEQ(namespaceUID), datasource.DeletedAtIsNil()).
			Select(datasource.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.ProbeTargetGroup.Query().
			Where(probetargetgroup.NamespaceUIDEQ(namespaceUID), probetargetgroup.DeletedAtIsNil()).
			Select(probetargetgroup.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.AlertRule.Query().
			Where(alertrule.NamespaceUIDEQ(namespaceUID), alertrule.DeletedAtIsNil()).
			Select(alertrule.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.AlertRuleGroup.Query().
			Where(alertrulegroup.NamespaceUIDEQ(namespaceUID), alertrulegroup.DeletedAtIsNil()).
			Select(alertrulegroup.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.AlertPage.Query().
			Where(alertpage.NamespaceUIDEQ(namespaceUID), alertpage.DeletedAtIsNil()).
			Select(alertpage.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.NotificationSubscription.Query().
			Where(notificationsubscription.NamespaceUIDEQ(namespaceUID), notificationsubscription.DeletedAtIsNil()).
			Select(notificationsubscription.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.NotificationOnCall.Query().
			Where(notificationoncall.NamespaceUIDEQ(namespaceUID), notificationoncall.DeletedAtIsNil()).
			Select(notificationoncall.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.InspectionPlan.Query().
			Where(inspectionplan.NamespaceUIDEQ(namespaceUID), inspectionplan.DeletedAtIsNil()).
			Select(inspectionplan.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	{
		rows, err := cache.client.Agent.Query().
			Where(agent.NamespaceUIDEQ(namespaceUID), agent.DeletedAtIsNil()).
			Select(agent.FieldLabels).All(ctx)
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Labels)
		}
		collect(out, err)
	}
	for _, item := range sources {
		if item.err != nil {
			return nil, item.err
		}
		for _, raw := range item.rows {
			add(raw)
		}
	}
	keys := make([]string, 0, len(index))
	for key := range index {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	options := make([]bo.LabelOption, 0, len(keys))
	for _, key := range keys {
		values := make([]string, 0, len(index[key]))
		for value := range index[key] {
			values = append(values, value)
		}
		sort.Strings(values)
		options = append(options, bo.LabelOption{Key: key, Values: values})
	}
	return options, nil
}
