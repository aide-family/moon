package do

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNamePromAlarmRealtime = "prom_alarm_realtime"

const (
	PromAlarmRealtimeFieldStrategyID               = "strategy_id"
	PromAlarmRealtimeFieldLevelID                  = "level_id"
	PromAlarmRealtimeFieldInstance                 = "instance"
	PromAlarmRealtimeFieldNote                     = "note"
	PromAlarmRealtimeFieldStatus                   = "status"
	PromAlarmRealtimeFieldEventAt                  = "event_at"
	PromAlarmRealtimeFieldNotifiedAt               = "notified_at"
	PromAlarmRealtimeFieldHistoryID                = "history_id"
	PromAlarmRealtimePreloadFieldStrategy          = "Strategy"
	PromAlarmRealtimePreloadFieldLevel             = "Level"
	PromAlarmRealtimePreloadFieldHistory           = "History"
	PromAlarmRealtimePreloadFieldIntervenes        = "AlarmIntervenes"
	PromAlarmRealtimePreloadFieldBeenNotifyMembers = "BeenNotifyMembers"
	PromAlarmRealtimePreloadFieldBeenChatGroups    = "BeenChatGroups"
	PromAlarmRealtimePreloadFieldAlarmUpgradeInfo  = "AlarmUpgradeInfo"
	PromAlarmRealtimePreloadFieldAlarmSuppressInfo = "AlarmSuppressInfo"
)

// PromAlarmRealtimeLike 查询关键字
func PromAlarmRealtimeLike(keyword string) basescopes.ScopeMethod {
	return basescopes.WhereLikeKeyword(keyword, PromAlarmRealtimeFieldNote, PromAlarmRealtimeFieldInstance)
}

// PromAlarmRealtimeEventAtDesc 事件时间倒序
func PromAlarmRealtimeEventAtDesc() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(PromAlarmRealtimeFieldEventAt + " " + basescopes.DESC)
	}
}

// PromAlarmRealtimeInHistoryIds 查询历史ID列表
func PromAlarmRealtimeInHistoryIds(historyIds ...uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromAlarmRealtimeFieldHistoryID, historyIds...)
}

// PromAlarmRealtimeClauseOnConflict 冲突处理
func PromAlarmRealtimeClauseOnConflict() clause.Expression {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: PromAlarmRealtimeFieldHistoryID}},
		UpdateAll: true,
	}
}

// PromAlarmRealtimePreloadLevel 预加载级别
func PromAlarmRealtimePreloadLevel() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromAlarmRealtimePreloadFieldLevel)
	}
}

// PromAlarmRealtimeInStrategyIds 查询策略ID列表
func PromAlarmRealtimeInStrategyIds(strategyIds ...uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromAlarmRealtimeFieldStrategyID, strategyIds...)
}

// PromAlarmRealtimeInLevelIds 查询级别ID列表
func PromAlarmRealtimeInLevelIds(levelIds ...uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromAlarmRealtimeFieldLevelID, levelIds...)
}

// PromAlarmRealtimeBetweenEventAt 查询时间范围
func PromAlarmRealtimeBetweenEventAt(min, max int64) basescopes.ScopeMethod {
	if min > max || min == max || min == 0 || max == 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return basescopes.BetweenColumn(PromAlarmRealtimeFieldEventAt, min, max)
}

// PromAlarmRealtimePreloadStrategy 预加载关联策略
func PromAlarmRealtimePreloadStrategy() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromAlarmRealtimePreloadFieldStrategy)
	}
}

// PromAlarmRealtime 实时告警信息
type PromAlarmRealtime struct {
	BaseModel
	// StrategyID 发生这条告警的具体策略信息
	StrategyID uint32        `gorm:"column:strategy_id;type:int unsigned;not null;index:idx__ar__strategy_id,priority:1;comment:策略ID"`
	Strategy   *PromStrategy `gorm:"foreignKey:StrategyID"`
	LevelId    uint32        `gorm:"column:level_id;type:int unsigned;not null;index:idx__ar__level_id,priority:1;comment:告警等级ID"`
	Level      *SysDict      `gorm:"foreignKey:LevelId"`
	// Instance 发生这条告警的具体实例信息
	Instance string `gorm:"column:instance;type:varchar(64);not null;index:idx__ar__instance,priority:1;comment:instance名称"`
	Note     string `gorm:"column:note;type:varchar(255);not null;comment:告警内容"`
	// Status 告警状态: 1告警;2恢复
	Status vobj.AlarmStatus `gorm:"column:status;type:tinyint;not null;default:1;comment:告警状态: 1告警;2恢复"`
	// EventAt 告警时间
	EventAt int64 `gorm:"column:event_at;type:bigint;not null;comment:告警时间"`
	// 通知对象, 记录事件发生时候实际的通知人员
	BeenNotifyMembers []*PromAlarmBeenNotifyMember    `gorm:"foreignKey:RealtimeAlarmID;comment:已通知成员"`
	BeenChatGroups    []*PromAlarmBeenNotifyChatGroup `gorm:"foreignKey:RealtimeAlarmID;comment:已通知群组"`
	NotifiedAt        int64                           `gorm:"column:notified_at;type:bigint;not null;default:0;comment:通知时间"`
	// HistoryID 对应的报警历史数据
	HistoryID uint32            `gorm:"column:history_id;type:int unsigned;not null;unique:idx__history_id,priority:1;comment:历史记录ID"`
	History   *PromAlarmHistory `gorm:"foreignKey:HistoryID"`
	// Intervenes 运维介入信息
	AlarmIntervenes []*PromAlarmIntervene `gorm:"foreignKey:RealtimeAlarmID"`
	// AlarmUpgradeInfo 告警升级信息
	AlarmUpgradeInfo *PromAlarmUpgrade `gorm:"foreignKey:RealtimeAlarmID"`
	// AlarmSuppressInfo 告警抑制信息
	AlarmSuppressInfo *PromAlarmSuppress `gorm:"foreignKey:RealtimeAlarmID"`
}

// TableName 表名
func (*PromAlarmRealtime) TableName() string {
	return TableNamePromAlarmRealtime
}

// GetStrategy 获取策略信息
func (p *PromAlarmRealtime) GetStrategy() *PromStrategy {
	if p.Strategy == nil {
		p.Strategy = &PromStrategy{}
		p.Strategy.ID = p.StrategyID
	}
	return p.Strategy
}

// GetAlarmIntervenes 获取运维介入信息
func (p *PromAlarmRealtime) GetAlarmIntervenes() []*PromAlarmIntervene {
	if p == nil {
		return nil
	}
	return p.AlarmIntervenes
}

// GetBeenNotifyMembers 获取通知对象
func (p *PromAlarmRealtime) GetBeenNotifyMembers() []*PromAlarmBeenNotifyMember {
	if p == nil {
		return nil
	}
	return p.BeenNotifyMembers
}

// GetBeenChatGroups 获取通知群组
func (p *PromAlarmRealtime) GetBeenChatGroups() []*PromAlarmBeenNotifyChatGroup {
	if p == nil {
		return nil
	}
	return p.BeenChatGroups
}

// GetHistory 获取历史记录
func (p *PromAlarmRealtime) GetHistory() *PromAlarmHistory {
	if p == nil {
		return nil
	}
	return p.History
}

// GetAlarmUpgradeInfo 获取告警升级信息
func (p *PromAlarmRealtime) GetAlarmUpgradeInfo() *PromAlarmUpgrade {
	if p == nil {
		return nil
	}
	return p.AlarmUpgradeInfo
}

// GetAlarmSuppressInfo 获取告警抑制信息
func (p *PromAlarmRealtime) GetAlarmSuppressInfo() *PromAlarmSuppress {
	if p == nil {
		return nil
	}
	return p.AlarmSuppressInfo
}

// GetLevel 获取告警等级
func (p *PromAlarmRealtime) GetLevel() *SysDict {
	if p == nil {
		return nil
	}
	return p.Level
}
