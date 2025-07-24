package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type LabelNoticeParams struct {
	Key               string
	Value             string
	ReceiverRoutesIds []uint32

	noticeGroupDos []do.NoticeGroup
}

func (l *LabelNoticeParams) GetNoticeGroupDos() []do.NoticeGroup {
	return l.noticeGroupDos
}

func (l *LabelNoticeParams) Validate() error {
	if len(l.ReceiverRoutesIds) == 0 {
		return merr.ErrorParams("receiver routes is required")
	}
	if len(l.ReceiverRoutesIds) != len(l.noticeGroupDos) {
		return merr.ErrorParams("receiver routes is not found")
	}
	return nil
}

type CreateTeamMetricStrategyLevelParams interface {
	Validate() error
	GetStrategyMetric() do.StrategyMetric
	GetLevel() do.TeamDict
	GetAlarmPages() []do.TeamDict
	GetNoticeGroupDos() []do.NoticeGroup
	GetReceiverRoutesIds() []uint32
	GetLabelNotices() []*LabelNoticeParams
	GetDuration() int64
	GetTotal() int64
	GetCondition() vobj.ConditionMetric
	GetValues() []float64
	GetSampleMode() vobj.SampleMode
	WithStrategyMetric(strategyMetric do.StrategyMetric)
	WithDicts(dicts []do.TeamDict)
	WithNoticeGroupDos(noticeGroupDos []do.NoticeGroup)
}

type UpdateTeamMetricStrategyLevelParams interface {
	CreateTeamMetricStrategyLevelParams
	GetStrategyMetricLevel() do.StrategyMetricRule
	WithStrategyMetricLevel(strategyMetricLevel do.StrategyMetricRule)
}

type SaveTeamMetricStrategyLevelParams struct {
	StrategyMetricLevelID uint32
	LevelID               uint32
	SampleMode            vobj.SampleMode
	Total                 int64
	Condition             vobj.ConditionMetric
	Values                []float64
	ReceiverRoutesIds     []uint32
	LabelNotices          []*LabelNoticeParams
	Duration              int64
	AlarmPages            []uint32
	StrategyMetricID      uint32

	strategyMetricDo      do.StrategyMetric
	strategyMetricLevelDo do.StrategyMetricRule
	noticeGroupDos        map[uint32]do.NoticeGroup
	dicts                 map[uint32]do.TeamDict
}

// GetCondition implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetCondition() vobj.ConditionMetric {
	return s.Condition
}

// GetDuration implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetDuration() int64 {
	return s.Duration
}

// GetNoticeGroupDos implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetNoticeGroupDos() []do.NoticeGroup {
	return slices.MapFilter(s.ReceiverRoutesIds, func(id uint32) (do.NoticeGroup, bool) {
		noticeGroup, ok := s.noticeGroupDos[id]
		return noticeGroup, ok
	})
}

// GetReceiverRoutesIds implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetReceiverRoutesIds() []uint32 {
	return s.ReceiverRoutesIds
}

// GetSampleMode implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetSampleMode() vobj.SampleMode {
	return s.SampleMode
}

// GetStrategyMetric implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetStrategyMetric() do.StrategyMetric {
	return s.strategyMetricDo
}

// GetStrategyMetricLevel implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetStrategyMetricLevel() do.StrategyMetricRule {
	return s.strategyMetricLevelDo
}

// GetTotal implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetTotal() int64 {
	return s.Total
}

// GetValues implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) GetValues() []float64 {
	return s.Values
}

// WithDicts implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) WithDicts(dicts []do.TeamDict) {
	s.dicts = slices.ToMap(dicts, func(dict do.TeamDict) uint32 {
		return dict.GetID()
	})
}

// WithNoticeGroupDos implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) WithNoticeGroupDos(noticeGroupDos []do.NoticeGroup) {
	s.noticeGroupDos = slices.ToMap(noticeGroupDos, func(noticeGroup do.NoticeGroup) uint32 {
		return noticeGroup.GetID()
	})
}

// WithStrategyMetric implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) WithStrategyMetric(strategyMetric do.StrategyMetric) {
	s.strategyMetricDo = strategyMetric
}

// WithStrategyMetricLevel implements UpdateTeamMetricStrategyLevelParams.
func (s *SaveTeamMetricStrategyLevelParams) WithStrategyMetricLevel(strategyMetricLevel do.StrategyMetricRule) {
	s.strategyMetricLevelDo = strategyMetricLevel
}

func (s *SaveTeamMetricStrategyLevelParams) Validate() error {
	if s.StrategyMetricID <= 0 {
		return merr.ErrorParams("strategy metric id is required")
	}
	if validate.IsNil(s.strategyMetricDo) {
		return merr.ErrorParams("strategy metric is not found")
	}
	if s.LevelID <= 0 {
		return merr.ErrorParams("level id is required")
	}
	if validate.IsNil(s.GetLevel()) {
		return merr.ErrorParams("level is not found")
	}
	dictIds := s.GetDictIds()
	if len(s.dicts) != len(dictIds) {
		return merr.ErrorParams("dicts is not found")
	}
	return nil
}

func (s *SaveTeamMetricStrategyLevelParams) GetNoticeGroupIds() []uint32 {
	list := make([]uint32, 0, len(s.ReceiverRoutesIds)+len(s.LabelNotices))
	list = append(list, s.ReceiverRoutesIds...)
	for _, labelNotice := range s.LabelNotices {
		list = append(list, labelNotice.ReceiverRoutesIds...)
	}
	return slices.Unique(slices.MapFilter(list, func(id uint32) (uint32, bool) {
		if id > 0 {
			return id, true
		}
		return 0, false
	}))
}

func (s *SaveTeamMetricStrategyLevelParams) GetDictIds() []uint32 {
	list := make([]uint32, 0, len(s.AlarmPages)+1)
	list = append(list, s.AlarmPages...)
	list = append(list, s.LevelID)
	return slices.Unique(slices.MapFilter(list, func(id uint32) (uint32, bool) {
		if id > 0 {
			return id, true
		}
		return 0, false
	}))
}

// GetLevel implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetLevel() do.TeamDict {
	return s.dicts[s.LevelID]
}

func (s *SaveTeamMetricStrategyLevelParams) GetAlarmPages() []do.TeamDict {
	return slices.MapFilter(s.AlarmPages, func(id uint32) (do.TeamDict, bool) {
		dict, ok := s.dicts[id]
		return dict, ok
	})
}

func (s *SaveTeamMetricStrategyLevelParams) GetLabelNotices() []*LabelNoticeParams {
	return slices.MapFilter(s.LabelNotices, func(labelNotice *LabelNoticeParams) (*LabelNoticeParams, bool) {
		labelNotice.noticeGroupDos = slices.Map(labelNotice.ReceiverRoutesIds, func(receiverRoute uint32) do.NoticeGroup {
			return s.noticeGroupDos[receiverRoute]
		})
		return labelNotice, true
	})
}

type ListTeamMetricStrategyLevelsParams struct {
	*PaginationRequest
	StrategyMetricID uint32
	LevelID          uint32
	Status           vobj.GlobalStatus
	Keyword          string
}

func (l *ListTeamMetricStrategyLevelsParams) ToListReply(items []do.StrategyMetricRule) *ListTeamMetricStrategyLevelsReply {
	return &ListTeamMetricStrategyLevelsReply{
		PaginationReply: l.ToReply(),
		Items:           items,
	}
}

type ListTeamMetricStrategyLevelsReply = ListReply[do.StrategyMetricRule]

type UpdateTeamMetricStrategyLevelStatusParams struct {
	StrategyMetricLevelIds []uint32
	Status                 vobj.GlobalStatus
}
