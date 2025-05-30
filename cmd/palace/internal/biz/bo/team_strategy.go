package bo

import (
	"strings"
	"unicode/utf8"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

type CreateTeamStrategyParams interface {
	Validate() error
	GetStrategyGroup() do.StrategyGroup
	GetName() string
	GetRemark() string
	GetStrategyType() vobj.StrategyType
	GetReceiverRoutes() []do.NoticeGroup
	WithStrategyGroup(strategyGroup do.StrategyGroup)
	WithReceiverRoutes(receiverRoutes []do.NoticeGroup)
}

type UpdateTeamStrategyParams interface {
	CreateTeamStrategyParams
	GetStrategy() do.Strategy
	WithStrategy(strategy do.Strategy)
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

// GetName implements UpdateTeamStrategyParams.
func (s *SaveTeamStrategyParams) GetName() string {
	return s.Name
}

// GetReceiverRoutes implements UpdateTeamStrategyParams.
func (s *SaveTeamStrategyParams) GetReceiverRoutes() []do.NoticeGroup {
	return s.receiverRoutes
}

// GetRemark implements UpdateTeamStrategyParams.
func (s *SaveTeamStrategyParams) GetRemark() string {
	return s.Remark
}

// GetStrategy implements UpdateTeamStrategyParams.
func (s *SaveTeamStrategyParams) GetStrategy() do.Strategy {
	return s.strategyDo
}

// GetStrategyGroup implements UpdateTeamStrategyParams.
func (s *SaveTeamStrategyParams) GetStrategyGroup() do.StrategyGroup {
	return s.strategyGroup
}

// GetStrategyType implements UpdateTeamStrategyParams.
func (s *SaveTeamStrategyParams) GetStrategyType() vobj.StrategyType {
	return s.StrategyType
}

func (s *SaveTeamStrategyParams) WithStrategyGroup(strategyGroup do.StrategyGroup) {
	s.strategyGroup = strategyGroup
}

func (s *SaveTeamStrategyParams) WithReceiverRoutes(receiverRoutes []do.NoticeGroup) {
	s.receiverRoutes = receiverRoutes
}

func (s *SaveTeamStrategyParams) WithStrategy(strategy do.Strategy) {
	s.strategyDo = strategy
}

func (s *SaveTeamStrategyParams) Validate() error {
	if s.StrategyGroupID <= 0 {
		return merr.ErrorParams("strategy group id is required")
	}
	if validate.IsNil(s.strategyGroup) || s.strategyGroup.GetID() != s.StrategyGroupID {
		return merr.ErrorParams("strategy group is not found")
	}
	if strings.TrimSpace(s.Name) == "" {
		return merr.ErrorParams("name is required")
	}
	if !s.StrategyType.Exist() {
		return merr.ErrorParams("strategy type is invalid")
	}
	if utf8.RuneCountInString(s.Remark) > 255 {
		return merr.ErrorParams("remark is too long")
	}
	if s.ID > 0 && validate.IsNil(s.strategyDo) {
		return merr.ErrorParams("strategy is not found")
	}
	if s.ID > 0 && s.strategyDo.GetID() != s.ID {
		return merr.ErrorParams("strategy is not found")
	}
	if validate.IsNotNil(s.strategyDo) {
		if s.strategyDo.GetStatus().IsEnable() && s.strategyDo.GetStrategyGroupID() != s.StrategyGroupID {
			return merr.ErrorBadRequest("enabled strategy cannot modify strategy group")
		}
	}

	return nil
}

type UpdateTeamStrategiesStatusParams struct {
	StrategyIds []uint32
	Status      vobj.GlobalStatus
}

type ListTeamStrategyParams struct {
	*PaginationRequest
	Keyword       string
	Status        vobj.GlobalStatus
	GroupIds      []uint32
	StrategyTypes []vobj.StrategyType
}

func (l *ListTeamStrategyParams) ToListReply(items []do.Strategy) *ListTeamStrategyReply {
	return &ListTeamStrategyReply{
		PaginationReply: l.ToReply(),
		Items:           items,
	}
}

type ListTeamStrategyReply = ListReply[do.Strategy]

type SubscribeTeamStrategiesParams struct {
	*PaginationRequest
	Subscribers []uint32
	NoticeType  vobj.NoticeType
}

func (s *SubscribeTeamStrategiesParams) ToListReply(items []do.TeamStrategySubscriber) *SubscribeTeamStrategiesReply {
	return &SubscribeTeamStrategiesReply{
		PaginationReply: s.ToReply(),
		Items:           items,
	}
}

type SubscribeTeamStrategiesReply = ListReply[do.TeamStrategySubscriber]

type SubscribeTeamStrategyParams struct {
	StrategyId uint32
	NoticeType vobj.NoticeType
}

func (s *SubscribeTeamStrategyParams) Validate() error {
	if s.StrategyId <= 0 {
		return merr.ErrorParams("strategy id is required")
	}
	if !s.NoticeType.Exist() {
		return merr.ErrorParams("notice type is invalid")
	}
	return nil
}
