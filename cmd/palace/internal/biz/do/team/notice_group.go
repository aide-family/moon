package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.NoticeGroup = (*NoticeGroup)(nil)

const tableNameNoticeGroup = "team_notice_groups"

type NoticeGroup struct {
	do.TeamModel
	Name          string            `gorm:"column:name;type:varchar(64);not null;comment:名称" json:"name"`
	Remark        string            `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status        vobj.GlobalStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Members       []*NoticeMember   `gorm:"many2many:team_notice_group_members" json:"members"`
	Hooks         []*NoticeHook     `gorm:"many2many:team_notice_group_hooks" json:"hooks"`
	EmailConfigID uint32            `gorm:"column:email_config_id;type:int(10);not null;comment:邮件配置ID" json:"emailConfigId"`
	EmailConfig   *EmailConfig      `gorm:"foreignKey:EmailConfigID;references:ID" json:"emailConfig"`
	SMSConfigID   uint32            `gorm:"column:sms_config_id;type:int(10);not null;comment:短信配置ID" json:"smsConfigId"`
	SMSConfig     *SmsConfig        `gorm:"foreignKey:SMSConfigID;references:ID" json:"smsConfig"`
}

func (n *NoticeGroup) GetEmailConfig() do.TeamEmailConfig {
	if n == nil {
		return nil
	}
	return n.EmailConfig
}

func (n *NoticeGroup) GetSMSConfig() do.TeamSMSConfig {
	if n == nil {
		return nil
	}
	return n.SMSConfig
}

func (n *NoticeGroup) GetName() string {
	if n == nil {
		return ""
	}
	return n.Name
}

func (n *NoticeGroup) GetRemark() string {
	if n == nil {
		return ""
	}
	return n.Remark
}

func (n *NoticeGroup) GetStatus() vobj.GlobalStatus {
	if n == nil {
		return vobj.GlobalStatusUnknown
	}
	return n.Status
}

func (n *NoticeGroup) GetHooks() []do.NoticeHook {
	if n == nil {
		return nil
	}
	return slices.Map(n.Hooks, func(h *NoticeHook) do.NoticeHook { return h })
}

func (n *NoticeGroup) GetNoticeMembers() []do.NoticeMember {
	if n == nil {
		return nil
	}
	return slices.Map(n.Members, func(m *NoticeMember) do.NoticeMember { return m })
}

func (n *NoticeGroup) TableName() string {
	return tableNameNoticeGroup
}
