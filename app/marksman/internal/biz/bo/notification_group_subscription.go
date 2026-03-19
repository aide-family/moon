package bo

import (
	"slices"

	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type StrategyLevelPairBo struct {
	StrategyUID int64
	LevelUID    int64
}

type SubscriptionFilterBo struct {
	StrategyGroupUIDs []int64
	StrategyUIDs      []int64
	StrategyLevels    []StrategyLevelPairBo
	Labels            map[string]string
}

// NewSubscriptionFilterBo builds SubscriptionFilterBo from proto SubscriptionFilter.
func NewSubscriptionFilterBo(req *apiv1.SubscriptionFilter) *SubscriptionFilterBo {
	if req == nil {
		return &SubscriptionFilterBo{}
	}
	levels := make([]StrategyLevelPairBo, 0, len(req.GetStrategyLevels()))
	for _, p := range req.GetStrategyLevels() {
		if p != nil {
			levels = append(levels, StrategyLevelPairBo{StrategyUID: p.GetStrategyUid(), LevelUID: p.GetLevelUid()})
		}
	}
	return &SubscriptionFilterBo{
		StrategyGroupUIDs: req.GetStrategyGroupUids(),
		StrategyUIDs:      req.GetStrategyUids(),
		StrategyLevels:    levels,
		Labels:            req.GetLabels(),
	}
}

// ToAPIV1SubscriptionFilter converts BO to proto SubscriptionFilter.
func ToAPIV1SubscriptionFilter(b *SubscriptionFilterBo) *apiv1.SubscriptionFilter {
	if b == nil {
		return nil
	}
	levels := make([]*apiv1.StrategyLevelPair, 0, len(b.StrategyLevels))
	for _, p := range b.StrategyLevels {
		levels = append(levels, &apiv1.StrategyLevelPair{
			StrategyUid: p.StrategyUID,
			LevelUid:    p.LevelUID,
		})
	}
	return &apiv1.SubscriptionFilter{
		StrategyGroupUids: b.StrategyGroupUIDs,
		StrategyUids:      b.StrategyUIDs,
		StrategyLevels:    levels,
		Labels:            b.Labels,
	}
}

// MatchesAlert returns true if the alert satisfies all non-empty dimensions of the filter (AND semantics).
// Used when deciding which notification groups receive an alert.
func (f *SubscriptionFilterBo) MatchesAlert(strategyGroupUID, strategyUID, levelUID snowflake.ID, labels map[string]string) bool {
	if f == nil {
		return false
	}
	// If all dimensions are empty, treat as no subscription (do not match).
	hasGroup := len(f.StrategyGroupUIDs) > 0
	hasStrategy := len(f.StrategyUIDs) > 0
	hasLevel := len(f.StrategyLevels) > 0
	hasLabels := len(f.Labels) > 0
	if !hasGroup && !hasStrategy && !hasLevel && !hasLabels {
		return false
	}
	if hasGroup {
		if !slices.Contains(f.StrategyGroupUIDs, strategyGroupUID.Int64()) {
			return false
		}
	}
	if hasStrategy {
		if !slices.Contains(f.StrategyUIDs, strategyUID.Int64()) {
			return false
		}
	}
	if hasLevel {
		s, l := strategyUID.Int64(), levelUID.Int64()
		if !slices.ContainsFunc(f.StrategyLevels, func(p StrategyLevelPairBo) bool { return p.StrategyUID == s && p.LevelUID == l }) {
			return false
		}
	}
	if hasLabels {
		if labels == nil {
			return false
		}
		for k, v := range f.Labels {
			if labels[k] != v {
				return false
			}
		}
	}
	return true
}
