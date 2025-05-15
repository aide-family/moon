package team

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
)

var _ do.NoticeMember = (*NoticeMember)(nil)

const tableNameNoticeMember = "team_notice_members"

type NoticeMember struct {
	do.TeamModel
	NoticeGroupID uint32          `gorm:"column:notice_group_id;type:int(10) unsigned;not null;comment:通知组ID" json:"noticeGroupID"`
	UserID        uint32          `gorm:"column:user_id;type:int(10) unsigned;not null;comment:用户ID" json:"userID"`
	NoticeType    vobj.NoticeType `gorm:"column:notice_type;type:int(10) unsigned;not null;comment:通知类型" json:"noticeType"`
	NoticeGroup   *NoticeGroup    `gorm:"foreignKey:NoticeGroupID;references:ID" json:"noticeGroup"`
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
	if n == nil || n.NoticeGroup == nil {
		return nil
	}
	return n.NoticeGroup
}

func (n *NoticeMember) TableName() string {
	return tableNameNoticeMember
}
