package do

import (
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNamePromNotifyMember = "prom_alarm_notify_members"

const (
	PromNotifyMemberFieldPromAlarmNotifyID = "prom_alarm_notify_id"
	PromNotifyMemberFieldStatus            = "status"
	PromNotifyMemberFieldNotifyTypes       = "notify_types"
	PromNotifyMemberFieldMemberId          = "member_id"
	PromNotifyMemberPreloadFieldMember     = "Member"
)

// PromAlarmNotifyMemberWherePromAlarmNotifyID .
func PromAlarmNotifyMemberWherePromAlarmNotifyID(promAlarmNotifyID uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if promAlarmNotifyID == 0 {
			return db
		}
		return db.Where(PromNotifyMemberFieldPromAlarmNotifyID, promAlarmNotifyID)
	}
}

type PromAlarmNotifyMember struct {
	BaseModel
	PromAlarmNotifyID uint32          `gorm:"column:prom_alarm_notify_id;type:int unsigned;not null;index:idx__nm__prom_alarm_notify_id,priority:1;comment:通知ID"`
	Status            vobj.Status     `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	NotifyType        vobj.NotifyType `gorm:"column:notify_types;type:tinyint;not null;comment:通知方式"`
	MemberId          uint32          `gorm:"column:member_id;type:int unsigned;not null;index:idx__nm__member_id,priority:1;comment:成员ID"`
	Member            *SysUser        `gorm:"foreignKey:MemberId;comment:成员"`
}

func (*PromAlarmNotifyMember) TableName() string {
	return TableNamePromNotifyMember
}

// GetMember 获取成员
func (p *PromAlarmNotifyMember) GetMember() *SysUser {
	if p == nil {
		return nil
	}
	return p.Member
}
