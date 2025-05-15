package bo

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

var _ UpdateTeamStrategyParams = (*SaveTeamStrategyParams)(nil)
var _ CreateTeamStrategyParams = (*SaveTeamStrategyParams)(nil)

type CreateTeamStrategyParams interface {
	GetStrategyGroup() do.StrategyGroup
	GetName() string
	GetRemark() string
	GetStrategyType() vobj.StrategyType
	GetReceiverRoutes() []do.NoticeGroup
	Validate() error
}

type UpdateTeamStrategyParams interface {
	GetStrategy() do.Strategy
	CreateTeamStrategyParams
}

type SaveTeamStrategyParams struct {
	StrategyGroupID uint32
	ID              uint32
	Name            string
	Remark          string
	StrategyType    vobj.StrategyType
	ReceiverRoutes  []uint32

	strategyDo     do.Strategy
	strategyGroup  do.StrategyGroup
	receiverRoutes []do.NoticeGroup
}

// GetStrategyGroup implements TeamStrategyParams.
func (s *SaveTeamStrategyParams) GetStrategyGroup() do.StrategyGroup {
	return s.strategyGroup
}

// GetID implements TeamStrategyParams.
func (s *SaveTeamStrategyParams) GetStrategy() do.Strategy {
	return s.strategyDo
}

// GetName implements TeamStrategyParams.
func (s *SaveTeamStrategyParams) GetName() string {
	return s.Name
}

// GetReceiverRoutes implements TeamStrategyParams.
func (s *SaveTeamStrategyParams) GetReceiverRoutes() []do.NoticeGroup {
	return s.receiverRoutes
}

// GetRemark implements TeamStrategyParams.
func (s *SaveTeamStrategyParams) GetRemark() string {
	return s.Remark
}

// GetStrategyType implements TeamStrategyParams.
func (s *SaveTeamStrategyParams) GetStrategyType() vobj.StrategyType {
	if validate.IsNil(s.strategyDo) {
		return s.StrategyType
	}
	if s.strategyDo.GetStatus().IsEnable() {
		return s.strategyDo.GetStrategyType()
	}
	return s.StrategyType
}

func (s *SaveTeamStrategyParams) Validate() error {
	if s.StrategyGroupID <= 0 {
		return merr.ErrorParamsError("strategy group id is required")
	}
	if validate.IsNil(s.strategyGroup) || s.strategyGroup.GetID() != s.StrategyGroupID {
		return merr.ErrorParamsError("strategy group is not found")
	}
	if strings.TrimSpace(s.Name) == "" {
		return merr.ErrorParamsError("name is required")
	}
	if !s.StrategyType.Exist() {
		return merr.ErrorParamsError("strategy type is invalid")
	}
	if utf8.RuneCountInString(s.Remark) > 255 {
		return merr.ErrorParamsError("remark is too long")
	}
	if s.ID > 0 && validate.IsNil(s.strategyDo) {
		return merr.ErrorParamsError("strategy is not found")
	}
	if s.ID > 0 && s.strategyDo.GetID() != s.ID {
		return merr.ErrorParamsError("strategy is not found")
	}
	if validate.IsNotNil(s.strategyDo) {
		if s.strategyDo.GetStatus().IsEnable() && s.strategyDo.GetStrategyGroupID() != s.StrategyGroupID {
			return merr.ErrorBadRequest("启用中的策略不能修改策略组")
		}
	}

	return nil
}

func (s *SaveTeamStrategyParams) ToUpdateTeamStrategyParams(
	strategyGroup do.StrategyGroup,
	strategyDo do.Strategy,
	receiverRoutes []do.NoticeGroup,
) UpdateTeamStrategyParams {
	s.strategyGroup = strategyGroup
	s.strategyDo = strategyDo
	s.receiverRoutes = receiverRoutes
	return s
}

func (s *SaveTeamStrategyParams) ToCreateTeamStrategyParams(
	strategyGroup do.StrategyGroup,
	receiverRoutes []do.NoticeGroup,
) CreateTeamStrategyParams {
	s.strategyGroup = strategyGroup
	s.receiverRoutes = receiverRoutes
	return s
}

var _ CreateTeamMetricStrategyParams = (*SaveTeamMetricStrategyParams)(nil)
var _ UpdateTeamMetricStrategyParams = (*SaveTeamMetricStrategyParams)(nil)

type CreateTeamMetricStrategyParams interface {
	GetStrategy() do.Strategy
	GetExpr() string
	GetLabels() kv.StringMap
	GetAnnotations() kv.StringMap
	GetDatasource() []do.DatasourceMetric
	Validate() error
}

type UpdateTeamMetricStrategyParams interface {
	CreateTeamMetricStrategyParams
	GetStrategyMetric() do.StrategyMetric
}

type SaveTeamMetricStrategyParams struct {
	StrategyID  uint32
	ID          uint32
	Expr        string
	Labels      kv.StringMap
	Annotations kv.StringMap
	Datasource  []uint32

	strategyDo       do.Strategy
	datasourceDos    []do.DatasourceMetric
	strategyMetricDo do.StrategyMetric
}

// GetAnnotations implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetAnnotations() kv.StringMap {
	return s.Annotations
}

// GetDatasource implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetDatasource() []do.DatasourceMetric {
	return s.datasourceDos
}

// GetExpr implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetExpr() string {
	return s.Expr
}

// GetStrategyMetric implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetStrategyMetric() do.StrategyMetric {
	return s.strategyMetricDo
}

// GetLabels implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetLabels() kv.StringMap {
	return s.Labels
}

// GetStrategy implements UpdateTeamMetricStrategyParams.
func (s *SaveTeamMetricStrategyParams) GetStrategy() do.Strategy {
	return s.strategyDo
}

func (s *SaveTeamMetricStrategyParams) ToCreateTeamMetricStrategyParams(strategyDo do.Strategy, datasourceDos []do.DatasourceMetric) CreateTeamMetricStrategyParams {
	s.strategyDo = strategyDo
	s.datasourceDos = datasourceDos
	return s
}

func (s *SaveTeamMetricStrategyParams) ToUpdateTeamMetricStrategyParams(
	strategyDo do.Strategy,
	datasourceDos []do.DatasourceMetric,
	strategyMetricDo do.StrategyMetric,
) UpdateTeamMetricStrategyParams {
	s.strategyDo = strategyDo
	s.strategyMetricDo = strategyMetricDo
	s.datasourceDos = datasourceDos
	return s
}

func (s *SaveTeamMetricStrategyParams) Validate() error {
	if s.StrategyID <= 0 {
		return merr.ErrorParamsError("strategy id is required")
	}
	if validate.IsNil(s.strategyDo) {
		return merr.ErrorParamsError("strategy is not found")
	}
	if strings.TrimSpace(s.Expr) == "" {
		return merr.ErrorParamsError("expr is required")
	}
	if len(s.Datasource) == 0 {
		return merr.ErrorParamsError("datasource is required")
	}
	if len(s.Annotations) == 0 {
		return merr.ErrorParamsError("annotations is required")
	}
	if s.ID > 0 && (validate.IsNil(s.strategyMetricDo) || s.strategyMetricDo.GetID() != s.ID) {
		return merr.ErrorParamsError("strategy metric is not found")
	}
	if len(s.Datasource) != len(s.datasourceDos) {
		return merr.ErrorParamsError("datasource is not found")
	}
	if s.strategyDo.GetStatus().IsEnable() {
		return merr.ErrorBadRequest("启用中的策略不能修改")
	}

	return nil
}

type LabelNotice interface {
	GetKey() string
	GetValue() string
	GetReceiverRoutes() []do.NoticeGroup
}

var _ LabelNotice = (*LabelNoticeParams)(nil)

type LabelNoticeParams struct {
	Key            string
	Value          string
	ReceiverRoutes []uint32

	noticeGroupDos []do.NoticeGroup
}

// GetKey implements LabelNotice.
func (l *LabelNoticeParams) GetKey() string {
	return l.Key
}

// GetReceiverRoutes implements LabelNotice.
func (l *LabelNoticeParams) GetReceiverRoutes() []do.NoticeGroup {
	return l.noticeGroupDos
}

// GetValue implements LabelNotice.
func (l *LabelNoticeParams) GetValue() string {
	return l.Value
}

type SaveTeamMetricStrategyLevel interface {
	GetID() uint32
	GetLevel() do.TeamDict
	GetAlarmPages() []do.TeamDict
	GetSampleMode() vobj.SampleMode
	GetTotal() int64
	GetCondition() vobj.ConditionMetric
	GetValues() []float64
	GetReceiverRoutes() []do.NoticeGroup
	GetLabelNotices() []LabelNotice
	GetDuration() time.Duration
	GetStatus() vobj.GlobalStatus
}

var _ SaveTeamMetricStrategyLevel = (*SaveTeamMetricStrategyLevelParams)(nil)

type SaveTeamMetricStrategyLevelParams struct {
	ID             uint32
	LevelId        uint32
	LevelName      string
	SampleMode     vobj.SampleMode
	Total          int64
	Condition      vobj.ConditionMetric
	Values         []float64
	ReceiverRoutes []uint32
	LabelNotices   []*LabelNoticeParams
	Duration       time.Duration
	Status         vobj.GlobalStatus
	AlarmPages     []uint32

	noticeGroupDos map[uint32]do.NoticeGroup
	dicts          map[uint32]do.TeamDict
}

func (s *SaveTeamMetricStrategyLevelParams) GetNoticeGroupIds() []uint32 {
	list := make([]uint32, 0, len(s.ReceiverRoutes)+len(s.LabelNotices))
	list = append(list, s.ReceiverRoutes...)
	for _, labelNotice := range s.LabelNotices {
		list = append(list, labelNotice.ReceiverRoutes...)
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
	list = append(list, s.LevelId)
	return slices.Unique(slices.MapFilter(list, func(id uint32) (uint32, bool) {
		if id > 0 {
			return id, true
		}
		return 0, false
	}))
}

// GetLevel implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetLevel() do.TeamDict {
	return s.dicts[s.LevelId]
}

func (s *SaveTeamMetricStrategyLevelParams) GetAlarmPages() []do.TeamDict {
	return slices.Map(s.AlarmPages, func(id uint32) do.TeamDict {
		return s.dicts[id]
	})
}

// GetCondition implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetCondition() vobj.ConditionMetric {
	return s.Condition
}

// GetTotal implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetTotal() int64 {
	return s.Total
}

// GetDuration implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetDuration() time.Duration {
	return s.Duration
}

// GetID implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetID() uint32 {
	return s.ID
}

// GetLabelNotices implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetLabelNotices() []LabelNotice {
	return slices.Map(s.LabelNotices, func(labelNotice *LabelNoticeParams) LabelNotice {
		labelNotice.noticeGroupDos = slices.Map(labelNotice.ReceiverRoutes, func(receiverRoute uint32) do.NoticeGroup {
			return s.noticeGroupDos[receiverRoute]
		})
		return labelNotice
	})
}

// GetLevelId implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetLevelId() uint32 {
	return s.LevelId
}

// GetLevelName implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetLevelName() string {
	return s.LevelName
}

// GetReceiverRoutes implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetReceiverRoutes() []do.NoticeGroup {
	return slices.Map(s.ReceiverRoutes, func(receiverRoute uint32) do.NoticeGroup {
		return s.noticeGroupDos[receiverRoute]
	})
}

// GetSampleMode implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetSampleMode() vobj.SampleMode {
	return s.SampleMode
}

// GetStatus implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetStatus() vobj.GlobalStatus {
	return s.Status
}

// GetValues implements SaveTeamMetricStrategyLevel.
func (s *SaveTeamMetricStrategyLevelParams) GetValues() []float64 {
	return s.Values
}

type SaveTeamMetricStrategyLevels interface {
	GetStrategyMetric() do.StrategyMetric
	GetLevels() []SaveTeamMetricStrategyLevel
}

var _ SaveTeamMetricStrategyLevels = (*SaveTeamMetricStrategyLevelsParams)(nil)

type SaveTeamMetricStrategyLevelsParams struct {
	StrategyMetricID uint32
	Levels           []*SaveTeamMetricStrategyLevelParams

	strategyMetricDo do.StrategyMetric
	noticeGroupDos   map[uint32]do.NoticeGroup
	dicts            map[uint32]do.TeamDict
}

func (s *SaveTeamMetricStrategyLevelsParams) Validate() error {
	if s.StrategyMetricID <= 0 {
		return merr.ErrorParamsError("strategy metric id is required")
	}
	if validate.IsNil(s.strategyMetricDo) || s.strategyMetricDo.GetID() != s.StrategyMetricID {
		return merr.ErrorParamsError("strategy metric is not found")
	}
	if len(s.Levels) > 5 {
		return merr.ErrorParamsError("levels is too many")
	}
	levelMap := slices.ToMap(s.Levels, func(level *SaveTeamMetricStrategyLevelParams) uint32 {
		return level.GetLevelId()
	})
	if len(levelMap) != len(s.Levels) {
		return merr.ErrorParamsError("level id is duplicate")
	}
	return nil
}

// GetLevels implements SaveTeamMetricStrategyLevels.
func (s *SaveTeamMetricStrategyLevelsParams) GetLevels() []SaveTeamMetricStrategyLevel {
	return slices.Map(s.Levels, func(level *SaveTeamMetricStrategyLevelParams) SaveTeamMetricStrategyLevel {
		level.noticeGroupDos = s.noticeGroupDos
		level.dicts = s.dicts
		return level
	})
}

// GetStrategyMetric implements SaveTeamMetricStrategyLevels.
func (s *SaveTeamMetricStrategyLevelsParams) GetStrategyMetric() do.StrategyMetric {
	return s.strategyMetricDo
}

func (s *SaveTeamMetricStrategyLevelsParams) ToSaveTeamMetricStrategyLevelsParams(
	strategyMetricDo do.StrategyMetric,
	noticeGroupDos []do.NoticeGroup,
	dicts []do.TeamDict,
) SaveTeamMetricStrategyLevels {
	s.strategyMetricDo = strategyMetricDo
	s.noticeGroupDos = slices.ToMap(noticeGroupDos, func(noticeGroup do.NoticeGroup) uint32 {
		return noticeGroup.GetID()
	})
	s.dicts = slices.ToMap(dicts, func(dict do.TeamDict) uint32 {
		return dict.GetID()
	})
	return s
}

type UpdateTeamStrategiesStatusParams struct {
	StrategyIds []uint32
	Status      vobj.GlobalStatus
}

func (s *UpdateTeamStrategiesStatusParams) Validate() error {
	if len(s.StrategyIds) == 0 {
		return merr.ErrorParamsError("strategy ids is required")
	}
	if !s.Status.Exist() {
		return merr.ErrorParamsError("status is invalid")
	}
	return nil
}

type OperateTeamStrategyParams struct {
	StrategyId uint32
}

func (s *OperateTeamStrategyParams) Validate() error {
	if s.StrategyId <= 0 {
		return merr.ErrorParamsError("strategy id is required")
	}
	return nil
}

type ListTeamStrategyParams struct {
	*PaginationRequest
	Keyword       string
	Status        vobj.GlobalStatus
	GroupIds      []uint32
	StrategyTypes []vobj.StrategyType
}

func (l *ListTeamStrategyParams) Validate() error {
	if l.Keyword != "" && utf8.RuneCountInString(l.Keyword) > 20 {
		return merr.ErrorParamsError("keyword is too long")
	}
	if !l.Status.Exist() {
		return merr.ErrorParamsError("status is invalid")
	}
	for _, strategyType := range l.StrategyTypes {
		if !strategyType.Exist() {
			return merr.ErrorParamsError("strategy type is invalid")
		}
	}
	return nil
}

func (l *ListTeamStrategyParams) ToListTeamStrategyReply(items []*team.Strategy) *ListTeamStrategyReply {
	return &ListTeamStrategyReply{
		PaginationReply: l.ToReply(),
		Items:           slices.Map(items, func(item *team.Strategy) do.Strategy { return item }),
	}
}

type ListTeamStrategyReply = ListReply[do.Strategy]

type SubscribeTeamStrategy interface {
	GetStrategy() do.Strategy
	GetNoticeType() vobj.NoticeType
}

type SubscribeTeamStrategyParams struct {
	StrategyId uint32
	NoticeType vobj.NoticeType

	strategyDo do.Strategy
}

func (s *SubscribeTeamStrategyParams) GetStrategy() do.Strategy {
	return s.strategyDo
}

func (s *SubscribeTeamStrategyParams) GetNoticeType() vobj.NoticeType {
	return s.NoticeType
}

func (s *SubscribeTeamStrategyParams) Validate() error {
	if s.StrategyId <= 0 {
		return merr.ErrorParamsError("strategy id is required")
	}
	if validate.IsNil(s.strategyDo) {
		return merr.ErrorParamsError("strategy is not found")
	}
	if !s.NoticeType.Exist() {
		return merr.ErrorParamsError("notice type is invalid")
	}
	return nil
}

func (s *SubscribeTeamStrategyParams) ToSubscribeTeamStrategyParams(strategyDo do.Strategy) SubscribeTeamStrategy {
	s.strategyDo = strategyDo
	return s
}

type SubscribeTeamStrategiesParams struct {
	*PaginationRequest
	StrategyId  uint32
	Subscribers []uint32
	NoticeType  vobj.NoticeType
}

func (s *SubscribeTeamStrategiesParams) Validate() error {
	if s.StrategyId <= 0 {
		return merr.ErrorParamsError("strategy id is required")
	}
	if !s.NoticeType.Exist() {
		return merr.ErrorParamsError("notice type is invalid")
	}
	return nil
}

func (s *SubscribeTeamStrategiesParams) ToSubscribeTeamStrategiesReply(items []*team.StrategySubscriber) *SubscribeTeamStrategiesReply {
	return &SubscribeTeamStrategiesReply{
		PaginationReply: s.ToReply(),
		Items:           slices.Map(items, func(item *team.StrategySubscriber) do.TeamStrategySubscriber { return item }),
	}
}

type SubscribeTeamStrategiesReply = ListReply[do.TeamStrategySubscriber]

type DeleteUnUsedLevelsParams struct {
	StrategyMetricID uint32
	RuleIds          []uint32
}

type FindTeamMetricStrategyLevelsParams struct {
	StrategyMetricID uint32
	RuleIds          []uint32
}
