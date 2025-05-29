package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.NoticeMember = (*NoticeMember)(nil)

const tableNameNoticeMember = "team_notice_members"

type NoticeMember struct {
	do.TeamModel
	NoticeGroupID uint32          `gorm:"column:notice_group_id;type:int(10) unsigned;not null;comment:notice group ID" json:"noticeGroupID"`
	UserID        uint32          `gorm:"column:user_id;type:int(10) unsigned;not null;comment:user ID" json:"userID"`
	NoticeType    vobj.NoticeType `gorm:"column:notice_type;type:int(10) unsigned;not null;comment:notice type" json:"noticeType"`
	NoticeGroup   *NoticeGroup    `gorm:"foreignKey:NoticeGroupID;references:ID" json:"noticeGroup"`
	DutyCycle     []*TimeEngine   `gorm:"many2many:team_notice_member_duty_cycles" json:"dutyCycle"`
}

func (n *NoticeMember) GetMember() do.TeamMember {
	if n == nil {
		return nil
	}
	return do.GetTeamMember(n.GetUserID())
}

func (n *NoticeMember) GetUserID() uint32 {
	if n == nil {
		return 0
	}
	return n.UserID
}

func (n *NoticeMember) GetNoticeGroupID() uint32 {
	if n == nil {
		return 0
	}
	return n.NoticeGroupID
}

func (n *NoticeMember) GetNoticeType() vobj.NoticeType {
	if n == nil {
		return vobj.NoticeTypeUnknown
	}
	return n.NoticeType
}

func (n *NoticeMember) GetNoticeGroup() do.NoticeGroup {
	if n == nil {
		return nil
	}
	return n.NoticeGroup
}

func (n *NoticeMember) GetDutyCycle() []do.TimeEngine {
	if n == nil {
		return nil
	}
	return slices.Map(n.DutyCycle, func(e *TimeEngine) do.TimeEngine { return e })
}

func (n *NoticeMember) TableName() string {
	return tableNameNoticeMember
}
