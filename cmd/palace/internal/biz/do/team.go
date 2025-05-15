package do

import (
	"github.com/google/uuid"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/plugin/cache"
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
	GetResources() []Resource
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
	GetPosition() vobj.Role
	GetRoles() []uint32
	GetInviteUser() User
	GetTeam() Team
}
