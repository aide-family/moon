package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNamePromNotifyMember = "prom_alarm_notify_members"

type PromAlarmNotifyMember struct {
	BaseModel
	PromAlarmNotifyID uint32         `gorm:"column:prom_alarm_notify_id;type:int unsigned;not null;index:idx__nm__prom_alarm_notify_id,priority:1;comment:通知ID"`
	Status            vo.Status      `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	NotifyTypes       vo.NotifyTypes `gorm:"column:notify_types;type:json;not null;comment:通知方式"`
	MemberId          uint32         `gorm:"column:member_id;type:int unsigned;not null;index:idx__nm__member_id,priority:1;comment:成员ID"`
	Member            *SysUser       `gorm:"foreignKey:MemberId;comment:成员"`
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

// GetNotifyTypes 获取通知方式
func (p *PromAlarmNotifyMember) GetNotifyTypes() vo.NotifyTypes {
	if p == nil {
		return nil
	}
	return p.NotifyTypes
}
