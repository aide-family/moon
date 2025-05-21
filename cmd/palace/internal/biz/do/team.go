package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/google/uuid"
)

type Team interface {
	cache.Object
	Creator
	GetName() string
	GetRemark() string
	GetLogo() string
	GetStatus() vobj.TeamStatus
	GetLeaderID() uint32
	GetUUID() uuid.UUID
	GetCapacity() vobj.TeamCapacity
	GetBizDBConfig() *config.Database
	GetAlarmDBConfig() *config.Database
	GetLeader() User
	GetAdmins() []User
	GetMenus() []Menu
}

type TeamAudit interface {
	Creator
	GetTeamID() uint32
	GetStatus() vobj.StatusAudit
	GetAction() vobj.AuditAction
	GetReason() string
	GetTeam() Team
}

type TeamInviteUser interface {
	Creator
	GetTeamID() uint32
	GetInviteUserID() uint32
	GetPosition() vobj.Position
	GetRoles() []uint32
	GetInviteUser() User
	GetTeam() Team
}
