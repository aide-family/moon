package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/strategy"
)

const TableNamePromStrategy = "prom_strategies"

// PromStrategy mapped from table <prom_strategies>
type PromStrategy struct {
	query.BaseModel
	GroupID      uint                  `gorm:"column:group_id;type:int unsigned;not null;comment:所属规则组ID"`
	Alert        string                `gorm:"column:alert;type:varchar(64);not null;comment:规则名称"`
	Expr         string                `gorm:"column:expr;type:text;not null;comment:prom ql"`
	For          string                `gorm:"column:for;type:varchar(64);not null;default:10s;comment:持续时间"`
	Labels       *strategy.Labels      `gorm:"column:labels;type:json;not null;comment:标签"`
	Annotations  *strategy.Annotations `gorm:"column:annotations;type:json;not null;comment:告警文案"`
	AlertLevelID uint                  `gorm:"column:alert_level_id;type:int;not null;index:idx__alert_level_id,priority:1;comment:告警等级dict ID"`
	Status       valueobj.Status       `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态: 1启用;2禁用"`
	Remark       string                `gorm:"column:remark;type:varchar(255);not null;comment:描述信息"`

	AlarmPages []*PromAlarmPage   `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:AlarmPageID;many2many:prom_strategy_alarm_pages"`
	Categories []*PromDict        `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:DictID;many2many:prom_strategy_categories"`
	AlertLevel *PromDict          `gorm:"foreignKey:AlertLevelID"`
	GroupInfo  *PromStrategyGroup `gorm:"foreignKey:GroupID"`

	// 通知对象
	PromNotifies []*PromAlarmNotify `gorm:"many2many:prom_strategy_notifies;comment:通知对象"`
	// 告警升级后的通知对象
	PromNotifyUpgrade []*PromAlarmNotify `gorm:"many2many:prom_strategy_notify_upgrades;comment:告警升级后的通知对象"`
	// 最大抑制时长(s)
	MaxSuppress int64 `gorm:"column:max_suppress;type:bigint;not null;default:0;comment:最大抑制时长(s)"`
	// 是否发送告警恢复通知
	SendRecover bool `gorm:"column:send_recover;type:tinyint;not null;default:0;comment:是否发送告警恢复通知"`
	// 发送告警时间间隔(s), 默认为for的10倍时间分钟, 用于长时间未消警情况
	SendInterval int64 `gorm:"column:send_interval;type:bigint;not null;default:0;comment:发送告警时间间隔(s), 默认为for的10倍时间分钟, 用于长时间未消警情况"`
}

// TableName PromStrategy's table name
func (*PromStrategy) TableName() string {
	return TableNamePromStrategy
}

// GetAlertLevel 获取告警等级
func (p *PromStrategy) GetAlertLevel() *PromDict {
	if p.AlertLevel == nil {
		return &PromDict{}
	}
	return p.AlertLevel
}

// GetAlarmPages 获取告警页面
func (p *PromStrategy) GetAlarmPages() []*PromAlarmPage {
	if p.AlarmPages == nil {
		return []*PromAlarmPage{}
	}
	return p.AlarmPages
}
