package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// GetRealTimeAlarmParams 获取实时告警参数
	GetRealTimeAlarmParams struct {
		// 告警ID
		RealtimeAlarmID uint32
		// 告警指纹
		Fingerprint string
	}

	// GetRealTimeAlarmsParams 获取实时告警列表参数
	GetRealTimeAlarmsParams struct {
		// 分页参数
		Pagination types.Pagination
		// 告警时间范围
		EventAtStart int64
		EventAtEnd   int64
		// 告警恢复时间
		ResolvedAtStart int64
		ResolvedAtEnd   int64
		// 告警级别
		AlarmLevels []uint32
		// 告警状态
		AlarmStatuses []vobj.AlertStatus
		// 关键字
		Keyword string
		// 告警页面
		AlarmPageID uint32
		// 我的告警
		MyAlarm bool
	}
)
