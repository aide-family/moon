package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func ToDictItem(dictItem do.TeamDict) *common.TeamDictItem {
	if validate.IsNil(dictItem) {
		return nil
	}
	return &common.TeamDictItem{
		TeamId:    dictItem.GetTeamID(),
		DictId:    dictItem.GetID(),
		CreatedAt: timex.Format(dictItem.GetCreatedAt()),
		UpdatedAt: timex.Format(dictItem.GetUpdatedAt()),
		Key:       dictItem.GetKey(),
		Value:     dictItem.GetValue(),
		Lang:      dictItem.GetLang(),
		Color:     dictItem.GetColor(),
		DictType:  common.DictType(dictItem.GetType()),
		Status:    common.GlobalStatus(dictItem.GetStatus().GetValue()),
	}
}

func ToDictItems(dictItems []do.TeamDict) []*common.TeamDictItem {
	return slices.Map(dictItems, ToDictItem)
}
