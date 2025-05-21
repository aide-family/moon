package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
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
	Status  vobj.GlobalStatus
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

type SelectTeamStrategyGroupRequest struct {
	*PaginationRequest
	Keyword string
	Status  []vobj.GlobalStatus
}

func (r *SelectTeamStrategyGroupRequest) ToSelectReply(groups []do.StrategyGroup) *SelectTeamStrategyGroupReply {
	return &SelectTeamStrategyGroupReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(groups, func(g do.StrategyGroup) SelectItem {
			return &selectItem{
				Value:    g.GetID(),
				Label:    g.GetName(),
				Disabled: g.GetDeletedAt() > 0 || !g.GetStatus().IsEnable(),
				Extra: &selectItemExtra{
					Remark: g.GetRemark(),
				},
			}
		}),
	}
}

type SelectTeamStrategyGroupReply = ListReply[SelectItem]
