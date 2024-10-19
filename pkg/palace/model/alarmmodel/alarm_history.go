package alarmmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameAlarmHistories = "alarm_histories"

// AlarmHistory 告警历史
type AlarmHistory struct {
	model.EasyModel
	// 告警摘要
	Summary string `gorm:"column:summary;type:varchar(255);not null;comment:告警摘要"`
	// 告警明细
	Description string `gorm:"column:description;type:text;not null;comment:告警明细"`
	// 告警消息状态
	AlertStatus vobj.AlertStatus `gorm:"column:status;type:varchar(16);not null;comment:告警消息状态"`
	// 告警时间
	StartsAt string `gorm:"column:starts_at;type:varchar(100);not null;comment:告警时间"`
	// 恢复时间
	EndsAt string `gorm:"column:ends_at;type:varchar(100);not null;comment:恢复时间"`
	// 告警表达式
	Expr string `gorm:"column:expr;type:text;not null;comment:prom ql"`
	// 指纹
	Fingerprint string `gorm:"column:fingerprint;type:varchar(255);not null;comment:fingerprint;uniqueIndex"`
	// 标签
	Labels *vobj.Labels `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	// 注解
	Annotations vobj.Annotations `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	// 告警原始数据ID
	RawInfoID uint32 `gorm:"column:raw_info_id;type:int;comment:告警原始数据id;uniqueIndex:idx__notice__raw_info_id,priority:1" json:"rawInfoId"`
	// 附加信息
	HistoryDetails *HistoryDetails `gorm:"foreignKey:AlarmHistoryID;comment:附加信息"`
}

// String json string
func (a *AlarmHistory) String() string {
	bs, _ := json.Marshal(a)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (a *AlarmHistory) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

// MarshalBinary redis存储实现
func (a *AlarmHistory) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

// TableName AlarmHistory's table name
func (a *AlarmHistory) TableName() string {
	return tableNameAlarmHistories
}

// GetHistoryDetails 获取附加信息
func (a *AlarmHistory) GetHistoryDetails() *HistoryDetails {
	if types.IsNil(a) || types.IsNil(a.HistoryDetails) {
		return &HistoryDetails{}
	}
	return a.HistoryDetails
}
