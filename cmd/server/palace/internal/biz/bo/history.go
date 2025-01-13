package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// QueryAlarmHistoryListParams 查询告警历史列表请求参数
	QueryAlarmHistoryListParams struct {
		Keyword     string             `json:"keyword"`
		Page        types.Pagination   `json:"page"`
		AlertStatus []vobj.AlertStatus `json:"alertStatus"`
		// 告警时间范围
		EventAtStart string
		EventAtEnd   string
		// 告警恢复时间
		ResolvedAtStart string
		ResolvedAtEnd   string
		AlarmPage       uint32
		StrategyIds     []uint32
	}

	// GetAlarmHistoryParams 获取告警告警历史参数
	GetAlarmHistoryParams struct {
		// 告警ID
		ID uint32 `json:"id"`
	}
)
