package do

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
)

type Menu interface {
	Base
	GetName() string
	GetPath() string
	GetStatus() vobj.GlobalStatus
	GetIcon() string
	GetParentID() uint32
	GetType() vobj.MenuType
	GetResources() []Resource
	GetParent() Menu
}
