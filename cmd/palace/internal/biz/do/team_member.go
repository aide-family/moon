package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

type TeamMember interface {
	cache.Object
	TeamBase
	GetTeamMemberID() uint32
	GetMemberName() string
	GetRemark() string
	GetUserID() uint32
	GetInviterID() uint32
	GetPosition() vobj.Role
	GetStatus() vobj.MemberStatus
	GetRoles() []TeamRole
	GetUser() User
	GetInviter() User
}
