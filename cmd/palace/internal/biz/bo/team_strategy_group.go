package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
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

func (r *ListTeamStrategyGroupParams) ToListReply(groups []do.StrategyGroup) *ListTeamStrategyGroupReply {
	return &ListTeamStrategyGroupReply{
		PaginationReply: r.ToReply(),
		Items:           groups,
	}
}

type UpdateTeamStrategyGroupStatusParams struct {
	ID     uint32
	Status vobj.GlobalStatus
}
