package bo

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateTemplateStrategyParams struct {
		*model.StrategyTemplate
	}

	UpdateTemplateStrategyParams struct {
		*model.StrategyTemplate
	}

	QueryTemplateStrategyListParams struct {
		Page   types.Pagination
		Alert  string
		Status vobj.Status
	}

	CreateStrategyAlarmLevel struct {
		//*model.StrategyAlarmLevel
	}

	UpdateStrategyAlarmLevel struct {
		//*model.StrategyAlarmLevel
	}

	QueryStrategyAlarmLevelListParams struct {
		Page    types.Pagination
		Keyword string
		Status  vobj.Status
	}
)
