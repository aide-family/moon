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
	ExcludeLabels     map[string]string
	DatasourceUIDs    []int64
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
		ExcludeLabels:     req.GetExcludeLabels(),
		DatasourceUIDs:    req.GetDatasourceUids(),
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
		ExcludeLabels:     b.ExcludeLabels,
		DatasourceUids:    b.DatasourceUIDs,
	}
}

type MatchesAlertParams struct {
	StrategyGroupUID snowflake.ID
	StrategyUID      snowflake.ID
	LevelUID         snowflake.ID
	DatasourceUID    snowflake.ID
	Labels           map[string]string
}

// MatchesAlert returns true if the alert matches at least one non-empty dimension of the filter (OR semantics).
//
// Notes:
// - `labels` in SubscriptionFilterBo matches against the provided `labels` argument.
// - When matching `labels`, all key-value pairs in the filter must be present in the alert with the same values.
// Used when deciding which notification groups receive an alert.
func (f *SubscriptionFilterBo) MatchesAlert(params *MatchesAlertParams) bool {
	if f == nil || params == nil {
		return false
	}

	labels := params.Labels
	hasExcludeLabels := len(f.ExcludeLabels) > 0
	// ExcludeLabels is a hard constraint:
	// if the alert labels contain ALL excluded key-value pairs with the same values,
	// the whole subscription rule is treated as non-matching, regardless of other dimensions.
	if hasExcludeLabels && len(labels) > 0 {
		excludeAllMatched := true
		for k, v := range f.ExcludeLabels {
			val, ok := labels[k]
			if !ok || val != v {
				excludeAllMatched = false
				break
			}
		}
		if excludeAllMatched {
			return false
		}
	}

	strategyGroupUID, strategyUID, levelUID, datasourceUID := params.StrategyGroupUID, params.StrategyUID, params.LevelUID, params.DatasourceUID
	// If all dimensions are empty, treat as no subscription (do not match).
	hasGroup := len(f.StrategyGroupUIDs) > 0
	hasStrategy := len(f.StrategyUIDs) > 0
	hasLevel := len(f.StrategyLevels) > 0
	hasLabels := len(f.Labels) > 0
	hasDatasource := len(f.DatasourceUIDs) > 0

	groupMatched := hasGroup && slices.Contains(f.StrategyGroupUIDs, strategyGroupUID.Int64())
	strategyMatched := hasStrategy && slices.Contains(f.StrategyUIDs, strategyUID.Int64())
	levelMatched := false
	if hasLevel {
		s, l := strategyUID.Int64(), levelUID.Int64()
		levelMatched = slices.ContainsFunc(f.StrategyLevels, func(p StrategyLevelPairBo) bool { return p.StrategyUID == s && p.LevelUID == l })
	}

	labelsMatched := false
	if hasLabels {
		if labels != nil {
			labelsMatched = true
			for k, v := range f.Labels {
				val, ok := labels[k]
				if !ok || val != v {
					labelsMatched = false
					break
				}
			}
		}
	}

	datasourceMatched := hasDatasource && slices.Contains(f.DatasourceUIDs, datasourceUID.Int64())

	return groupMatched || strategyMatched || levelMatched || datasourceMatched || labelsMatched
}
