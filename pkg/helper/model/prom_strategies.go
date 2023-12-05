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
	GroupID      uint32                `gorm:"column:group_id;type:int unsigned;not null;comment:所属规则组ID"`
	Alert        string                `gorm:"column:alert;type:varchar(64);not null;comment:规则名称"`
	Expr         string                `gorm:"column:expr;type:text;not null;comment:prom ql"`
	For          string                `gorm:"column:for;type:varchar(64);not null;default:10s;comment:持续时间"`
	Labels       *strategy.Labels      `gorm:"column:labels;type:json;not null;comment:标签"`
	Annotations  *strategy.Annotations `gorm:"column:annotations;type:json;not null;comment:告警文案"`
	AlertLevelID uint32                `gorm:"column:alert_level_id;type:int;not null;index:idx__alert_level_id,priority:1;comment:告警等级dict ID"`
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
	if p == nil {
		return nil
	}
	return p.AlertLevel
}

// GetAlarmPages 获取告警页面
func (p *PromStrategy) GetAlarmPages() []*PromAlarmPage {
	if p == nil {
		return nil
	}
	return p.AlarmPages
}

// GetLabels 获取标签
func (p *PromStrategy) GetLabels() *strategy.Labels {
	if p == nil {
		return nil
	}
	return p.Labels
}

// GetAnnotations 获取告警文案
func (p *PromStrategy) GetAnnotations() *strategy.Annotations {
	if p == nil {
		return nil
	}
	return p.Annotations
}

// GetCategories 获取分类
func (p *PromStrategy) GetCategories() []*PromDict {
	if p == nil {
		return nil
	}
	return p.Categories
}

// GetPromNotifies 获取通知对象
func (p *PromStrategy) GetPromNotifies() []*PromAlarmNotify {
	if p == nil {
		return nil
	}
	return p.PromNotifies
}

// GetPromNotifyUpgrade 获取告警升级后的通知对象
func (p *PromStrategy) GetPromNotifyUpgrade() []*PromAlarmNotify {
	if p == nil {
		return nil
	}
	return p.PromNotifyUpgrade
}

// GetGroupInfo 获取所属规则组
func (p *PromStrategy) GetGroupInfo() *PromStrategyGroup {
	if p == nil {
		return nil
	}
	return p.GroupInfo
}
