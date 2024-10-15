package imodel

import (
	"github.com/aide-family/moon/pkg/vobj"
)

type IResource interface {
	IAllFieldModel
	GetName() string
	GetPath() string
	GetStatus() vobj.Status
	GetRemark() string
	GetModule() int32
	GetDomain() int32
	GetAllow() vobj.Allow
}
