package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (

	// CreateAlarmHookParams 创建hook参数
	CreateAlarmHookParams struct {
		// Hook的名称
		Name string `json:"name"`
		// hook说明信息
		Remark string `json:"remark"`
		// url
		URL string `json:"url"`
		// secret
		Secret string `json:"secret"`
		// hook app
		HookApp vobj.HookAPP `json:"hookApp"`
		// status
		Status vobj.Status `json:"status"`
	}

	// QueryAlarmHookListParams 查询hook列表
	QueryAlarmHookListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
		Name    string
		Status  vobj.Status
		Apps    []vobj.HookAPP
	}

	// UpdateAlarmHookParams 更新hook参数
	UpdateAlarmHookParams struct {
		ID          uint32                 `json:"id"`
		UpdateParam *CreateAlarmHookParams `json:"updateParam"`
	}

	// UpdateAlarmHookStatusParams 更新hook状态
	UpdateAlarmHookStatusParams struct {
		IDs    []uint32 `json:"ids"`
		Status vobj.Status
	}
)
