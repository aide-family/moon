package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

type SaveTeamStrategyGroupParams struct {
	ID     uint32
	Name   string
	Remark string
}

type ListTeamStrategyGroupParams struct {
	*PaginationRequest
	Keyword string
	Status  []vobj.GlobalStatus
}

type ListTeamStrategyGroupReply = ListReply[do.StrategyGroup]

func (r *ListTeamStrategyGroupParams) ToListTeamStrategyGroupReply(groups []*team.StrategyGroup) *ListTeamStrategyGroupReply {
	return &ListTeamStrategyGroupReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(groups, func(group *team.StrategyGroup) do.StrategyGroup { return group }),
	}
}

type UpdateTeamStrategyGroupStatusParams struct {
	ID     uint32
	Status vobj.GlobalStatus
}
