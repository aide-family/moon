package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (

	// CreateAlarmNoticeGroupParams 创建告警组请求参数
	CreateAlarmNoticeGroupParams struct {
		// 告警组名称
		Name string `json:"name,omitempty"`
		// 告警组说明信息
		Remark string `json:"remark,omitempty"`
		// 告警组状态
		Status vobj.Status `json:"status,omitempty"`
		// 告警分组通知人
		NoticeUsers []*CreateNoticeUserParams `json:"noticeUsers,omitempty"`
		// hook ids
		HookIds []uint32 `json:"hookIds"`
	}

	// CreateNoticeUserParams 创建通知人参数
	CreateNoticeUserParams struct {
		// 用户id
		UserID uint32
		// 通知方式
		NotifyType vobj.NotifyType
	}

	// UpdateAlarmNoticeGroupStatusParams 更新告警组状态请求参数
	UpdateAlarmNoticeGroupStatusParams struct {
		IDs    []uint32 `json:"ids"`
		Status vobj.Status
	}

	// UpdateAlarmNoticeGroupParams 更新告警组请求参数
	UpdateAlarmNoticeGroupParams struct {
		ID          uint32 `json:"id"`
		UpdateParam *CreateAlarmNoticeGroupParams
	}

	// QueryAlarmNoticeGroupListParams 查询告警组列表请求参数
	QueryAlarmNoticeGroupListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
		Name    string
		Status  vobj.Status
	}

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
