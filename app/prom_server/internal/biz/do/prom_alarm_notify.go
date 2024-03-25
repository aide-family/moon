package do

import (
	"fmt"

	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNamePromNotify = "prom_alarm_notifies"

const (
	PromAlarmNotifyFiledName     = basescopes.BaseFieldName
	PromAlarmNotifyFiledStatus   = basescopes.BaseFieldStatus
	PromAlarmNotifyFiledRemark   = basescopes.BaseFieldRemark
	PromAlarmNotifyFiledCreateBy = "create_by"

	PromAlarmNotifyPreloadFieldChatGroups         = "ChatGroups"
	PromAlarmNotifyPreloadFieldBeNotifyMembers    = "BeNotifyMembers"
	PromAlarmNotifyPreloadFieldExternalNotifyObjs = "ExternalNotifyObjs"
)

// PromAlarmNotifyPreloadChatGroups 预加载通知组
func PromAlarmNotifyPreloadChatGroups(chatGroupIds ...uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(chatGroupIds) > 0 {
			return db.Preload(PromAlarmNotifyPreloadFieldChatGroups, basescopes.WhereInColumn(basescopes.BaseFieldID, chatGroupIds...))
		}
		return db.Preload(PromAlarmNotifyPreloadFieldChatGroups)
	}
}

// PromAlarmNotifyPreloadBeNotifyMembers 预加载被通知成员
func PromAlarmNotifyPreloadBeNotifyMembers(beNotifyMemberIds ...uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Preload(fmt.Sprintf("%s.%s", PromAlarmNotifyPreloadFieldBeNotifyMembers, PromNotifyMemberPreloadFieldMember))
		if len(beNotifyMemberIds) > 0 {
			return db.Preload(PromAlarmNotifyPreloadFieldBeNotifyMembers, basescopes.WhereInColumn(basescopes.BaseFieldID, beNotifyMemberIds...))
		}
		return db.Preload(PromAlarmNotifyPreloadFieldBeNotifyMembers)
	}
}

// PromAlarmNotify 告警通知对象
type PromAlarmNotify struct {
	BaseModel
	Name            string                   `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__an__name,priority:1;comment:通知名称"`
	Status          vobj.Status              `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark          string                   `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	ChatGroups      []*PromAlarmChatGroup    `gorm:"many2many:prom_notify_chat_groups;comment:通知组"`
	BeNotifyMembers []*PromAlarmNotifyMember `gorm:"foreignKey:PromAlarmNotifyID;comment:被通知成员"`
	// 外部体系通知对象(不在用户体系内的人和hook), 多对多
	ExternalNotifyObjs []*ExternalNotifyObj `gorm:"many2many:prom_alarm_notify_external_notify_objs;comment:外部体系通知对象"`
	// 创建人ID
	CreateBy uint32 `gorm:"column:create_by;type:int;not null;comment:创建人ID"`
}

func (*PromAlarmNotify) TableName() string {
	return TableNamePromNotify
}

// GetChatGroups 获取通知组
func (p *PromAlarmNotify) GetChatGroups() []*PromAlarmChatGroup {
	if p == nil {
		return nil
	}
	return p.ChatGroups
}

// GetBeNotifyMembers 获取被通知成员
func (p *PromAlarmNotify) GetBeNotifyMembers() []*PromAlarmNotifyMember {
	if p == nil {
		return nil
	}
	return p.BeNotifyMembers
}

// GetExternalNotifyObjs 获取外部体系通知对象
func (p *PromAlarmNotify) GetExternalNotifyObjs() []*ExternalNotifyObj {
	if p == nil {
		return nil
	}
	return p.ExternalNotifyObjs
}
