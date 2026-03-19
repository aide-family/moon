package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/data/impl/do"
)

func ToSubscriptionFilterBo(m *do.NotificationGroupSubscription) *bo.SubscriptionFilterBo {
	if m == nil {
		return nil
	}
	var strategyGroupUIDs, strategyUIDs []int64
	if m.StrategyGroupUIDs != nil {
		strategyGroupUIDs = m.StrategyGroupUIDs.List()
	}
	if m.StrategyUIDs != nil {
		strategyUIDs = m.StrategyUIDs.List()
	}
	levels := make([]bo.StrategyLevelPairBo, 0, len(m.StrategyLevels))
	for _, p := range m.StrategyLevels {
		levels = append(levels, bo.StrategyLevelPairBo{StrategyUID: p.StrategyUID, LevelUID: p.LevelUID})
	}
	var labels map[string]string
	if m.Labels != nil {
		labels = m.Labels.Map()
	}
	return &bo.SubscriptionFilterBo{
		StrategyGroupUIDs: strategyGroupUIDs,
		StrategyUIDs:      strategyUIDs,
		StrategyLevels:    levels,
		Labels:            labels,
	}
}

func ToNotificationGroupSubscriptionDO(ctx context.Context, notificationGroupUID snowflake.ID, req *bo.SubscriptionFilterBo) *do.NotificationGroupSubscription {
	if req == nil {
		req = &bo.SubscriptionFilterBo{}
	}
	levels := make(do.StrategyLevelPairsDO, 0, len(req.StrategyLevels))
	for _, p := range req.StrategyLevels {
		levels = append(levels, do.StrategyLevelPairDO{StrategyUID: p.StrategyUID, LevelUID: p.LevelUID})
	}
	m := &do.NotificationGroupSubscription{
		NotificationGroupUID: notificationGroupUID,
		StrategyGroupUIDs:    safety.NewSlice(req.StrategyGroupUIDs),
		StrategyUIDs:         safety.NewSlice(req.StrategyUIDs),
		StrategyLevels:       levels,
		Labels:               safety.NewMap(req.Labels),
	}
	m.Creator = contextx.GetUserUID(ctx)
	return m
}
