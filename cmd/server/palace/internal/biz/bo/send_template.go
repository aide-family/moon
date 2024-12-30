package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// CreateSendTemplate 创建发送模板
	CreateSendTemplate struct {
		// Name 模板名称
		Name string `json:"name"`
		// Content 模板内容
		Content string `json:"content"`
		// SendType 发送类型
		SendType vobj.AlarmSendType `json:"sendType"`
		// SendTarget 发送目标
		Status vobj.Status `json:"status"`
		// Remark 备注
		Remark string `json:"remark"`
	}
	// UpdateSendTemplate 更新发送模板
	UpdateSendTemplate struct {
		ID          uint32             `json:"id"`
		UpdateParam CreateSendTemplate `json:"updateParam"`
	}

	// QuerySendTemplateListParams 查询发送模板列表参数
	QuerySendTemplateListParams struct {
		Page     types.Pagination
		Keyword  string             `json:"keyword"`
		Status   vobj.Status        `json:"status"`
		SendType vobj.AlarmSendType `json:"sendType"`
	}

	// UpdateSendTemplateStatus 更新发送模板状态
	UpdateSendTemplateStatus struct {
		ID     []uint32    `json:"id"`
		Status vobj.Status `json:"status"`
	}
)
