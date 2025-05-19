package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	api "github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToTeamStrategyGroupItem(group do.StrategyGroup) *common.TeamStrategyGroupItem {
	if validate.IsNil(group) {
		return nil
	}
	return &common.TeamStrategyGroupItem{
		Name:                group.GetName(),
		Remark:              group.GetRemark(),
		GroupId:             group.GetID(),
		Status:              common.GlobalStatus(group.GetStatus().GetValue()),
		StrategyCount:       0,
		EnableStrategyCount: 0,
		CreatedAt:           timex.Format(group.GetCreatedAt()),
		UpdatedAt:           timex.Format(group.GetUpdatedAt()),
		Creator:             ToUserBaseItem(group.GetCreator()),
	}
}

func ToTeamStrategyGroupItems(groups []do.StrategyGroup) []*common.TeamStrategyGroupItem {
	return slices.Map(groups, ToTeamStrategyGroupItem)
}

func ToSelectTeamStrategyGroupParams(req *api.SelectTeamStrategyGroupRequest) *bo.SelectTeamStrategyGroupRequest {
	if req == nil {
		return nil
	}
	return &bo.SelectTeamStrategyGroupRequest{
		PaginationRequest: ToPaginationRequest(req.Pagination),
		Status: slices.Map(req.Status, func(status common.GlobalStatus) vobj.GlobalStatus {
			return vobj.GlobalStatus(status)
		}),
		Keyword: req.Keyword,
	}
}
