package system

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

var _ do.TeamAudit = (*TeamAudit)(nil)

const tableNameTeamAudit = "sys_team_audits"

type TeamAudit struct {
	do.CreatorModel
	TeamID  uint32           `gorm:"column:team_id;type:int unsigned;not null;comment:团队ID" json:"team_id,omitempty"`
	Status  vobj.StatusAudit `gorm:"column:status;type:tinyint(2);not null;comment:审批状态" json:"status"`
	Action  vobj.AuditAction `gorm:"column:action;type:tinyint(2);not null;comment:操作" json:"action"`
	Reason  string           `gorm:"column:reason;type:varchar(255);not null;comment:原因" json:"reason"`
	Creator *User            `gorm:"foreignKey:CreatorID;references:ID" json:"creator"`
	Team    *Team            `gorm:"foreignKey:TeamID;references:ID" json:"team"`
}

func (u *TeamAudit) GetTeamID() uint32 {
	if u == nil {
		return 0
	}
	return u.TeamID
}

func (u *TeamAudit) GetStatus() vobj.StatusAudit {
	if u == nil {
		return vobj.AuditStatusUnknown
	}
	return u.Status
}

func (u *TeamAudit) GetAction() vobj.AuditAction {
	if u == nil {
		return vobj.AuditActionUnknown
	}
	return u.Action
}

func (u *TeamAudit) GetReason() string {
	if u == nil {
		return ""
	}
	return u.Reason
}

func (u *TeamAudit) GetTeam() do.Team {
	if u == nil {
		return nil
	}
	return u.Team
}

func (u *TeamAudit) TableName() string {
	return tableNameTeamAudit
}
