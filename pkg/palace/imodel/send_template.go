package imodel

import "github.com/aide-family/moon/pkg/vobj"

// ISendTemplate 发送模板
type ISendTemplate interface {
	// GetID 获取id
	GetID() uint32
	// GetName 获取模板名称
	GetName() string
	// GetContent 获取模板内容
	GetContent() string
	// GetSendType 获取发送类型
	GetSendType() vobj.AlarmSendType
	// GetRemark 获取备注
	GetRemark() string
	// GetStatus 获取状态
	GetStatus() vobj.Status
}
