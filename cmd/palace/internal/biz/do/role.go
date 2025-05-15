package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Role interface {
	Creator
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetUsers() []User
	GetMenus() []Menu
}
