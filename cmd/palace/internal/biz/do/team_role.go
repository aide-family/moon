package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type TeamRole interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetMembers() []TeamMember
	GetMenus() []Menu
}
