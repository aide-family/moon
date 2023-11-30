package model

import (
	"database/sql/driver"
	"encoding/json"

	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromNotifyMember = "prom_notify_members"

type NotifyTypes []int32

func (l *NotifyTypes) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}

	str, err := json.Marshal(l)
	return string(str), err
}

func (l *NotifyTypes) Scan(src any) error {
	return json.Unmarshal(src.([]byte), l)
}

type PromNotifyMember struct {
	query.BaseModel
	Status      int32       `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`
	NotifyTypes NotifyTypes `gorm:"column:notify_types;type:json;not null;comment:通知方式" json:"notifyTypes"`
	MemberId    uint        `gorm:"column:member_id;type:int unsigned;not null;index:idx_member_id,priority:1;comment:成员ID" json:"memberId"`
	Member      *SysUser    `gorm:"foreignKey:MemberId;references:ID;joinForeignKey:MemberId;joinReferences:ID;comment:成员" json:"member"`
}

func (*PromNotifyMember) TableName() string {
	return TableNamePromNotifyMember
}
