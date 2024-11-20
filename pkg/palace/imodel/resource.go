package imodel

import (
	"github.com/aide-family/moon/pkg/vobj"
)

// IResource 资源模型
type IResource interface {
	IAllFieldModel
	// GetName 获取名称
	GetName() string
	// GetPath 获取路径
	GetPath() string
	// GetStatus 获取状态
	GetStatus() vobj.Status
	// GetRemark 获取备注
	GetRemark() string
	// GetModule 获取模块
	GetModule() int32
	// GetDomain 获取领域
	GetDomain() int32
	// GetAllow 获取放行规则
	GetAllow() vobj.Allow
}
