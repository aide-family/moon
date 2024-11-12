package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (

	// GetAlarmSendHistoryParams 获取告警发送历史参数
	GetAlarmSendHistoryParams struct {
		// 告警发送历史ID
		ID uint32 `json:"id"`
	}

	// QueryAlarmSendHistoryListParams 查询告警发送历史列表请求参数
	QueryAlarmSendHistoryListParams struct {
		Keyword       string           `json:"keyword"`
		Page          types.Pagination `json:"page"`
		SendStatus    []vobj.SendStatus
		StartSendTime string `json:"startTime"`
		EndSendTime   string `json:"endTime"`
	}

	// RetryAlarmSendParams 重试告警发送请求参数
	RetryAlarmSendParams struct {
		// 告警发送历史ID
		RequestID string `json:"requestId"`
	}

	// CreateAlarmSendParams 创建告警发送请求参数
	CreateAlarmSendParams struct {
		AlarmGroupID uint32          `json:"alarmGroupId"`
		SendData     string          `json:"sendData"`
		RequestID    string          `json:"requestId"`
		RetryNumber  int             `json:"retryNumber"`
		SendStatus   vobj.SendStatus `json:"sendStatus"`
		TeamID       uint32          `json:"teamId"`
		SendTime     *types.Time     `json:"sendTime"`
		Route        string          `json:"route"`
	}

	UpdateAlarmSendParams struct {
		ID          uint32 `json:"id"`
		UpdateParam *CreateAlarmSendParams
	}
)
