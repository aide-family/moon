package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
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
