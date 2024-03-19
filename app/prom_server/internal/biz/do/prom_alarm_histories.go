package do

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNamePromAlarmHistory = "prom_alarm_histories"

const (
	PromAlarmHistoryFieldInstance        = "instance"
	PromAlarmHistoryFieldStatus          = "status"
	PromAlarmHistoryFieldInfo            = "info"
	PromAlarmHistoryFieldStartAt         = "start_at"
	PromAlarmHistoryFieldEndAt           = "end_at"
	PromAlarmHistoryFieldDuration        = "duration"
	PromAlarmHistoryFieldStrategyID      = "strategy_id"
	PromAlarmHistoryFieldLevelID         = "level_id"
	PromAlarmHistoryFieldMd5             = "md5"
	PromAlarmHistoryPreloadFieldStrategy = "Strategy"
	PromAlarmHistoryPreloadFieldLevel    = "Level"
)

// PromAlarmHistoryLikeInstance 根据字典名称模糊查询
func PromAlarmHistoryLikeInstance(keyword string) basescopes.ScopeMethod {
	return basescopes.WhereLikePrefixKeyword(keyword, PromAlarmHistoryFieldInstance)
}

// PromAlarmHistoryLikeInfo 根据字典名称模糊查询
func PromAlarmHistoryLikeInfo(keyword string) basescopes.ScopeMethod {
	return basescopes.WhereLikeKeyword(keyword, PromAlarmHistoryFieldInfo, PromAlarmHistoryFieldInstance)
}

// PromAlarmHistoryStartTimeRange 根据时间范围查询
func PromAlarmHistoryStartTimeRange(startTime, endTime int64) basescopes.ScopeMethod {
	if startTime == 0 || endTime == 0 || startTime > endTime {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return basescopes.BetweenColumn(PromAlarmHistoryFieldStartAt, startTime, endTime)
}

// PromAlarmHistoryEndTimeRange 根据时间范围查询
func PromAlarmHistoryEndTimeRange(startTime, endTime int64) basescopes.ScopeMethod {
	if startTime == 0 || endTime == 0 || startTime > endTime {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return basescopes.BetweenColumn(PromAlarmHistoryFieldEndAt, startTime, endTime)
}

// PromAlarmHistoryWhereInMd5 根据md5查询
func PromAlarmHistoryWhereInMd5(md5s ...string) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromAlarmHistoryFieldMd5, md5s...)
}

// PromAlarmHistoryWhereInLevelID 根据等级ID查询
func PromAlarmHistoryWhereInLevelID(levelIds ...uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromAlarmHistoryFieldLevelID, levelIds...)
}

// PromAlarmHistoryWhereInStrategyID 根据策略ID查询
func PromAlarmHistoryWhereInStrategyID(strategyIds ...uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromAlarmHistoryFieldStrategyID, strategyIds...)
}

// PromAlarmHistoryWhereStatus 根据状态查询
func PromAlarmHistoryWhereStatus(status vo.AlarmStatus) basescopes.ScopeMethod {
	if status.IsUnknown() {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return basescopes.WhereInColumn(PromAlarmHistoryFieldStatus, status)
}

const m1 = 1 * 60
const m5 = 5 * 60
const m30 = 30 * 60
const m30p1 = 30*60 + 1

// PromAlarmHistoryWhereDuration 根据持续时间查询
func PromAlarmHistoryWhereDuration(duration int64) basescopes.ScopeMethod {
	nowTimeUnix := time.Now().Unix()
	return func(db *gorm.DB) *gorm.DB {
		switch duration {
		default:
			return db
		case m1:
			return db.Where(
				fmt.Sprintf("`%s` <= ? OR (`%s` = 0 AND `%s` <= ?  AND `%s` > ?)",
					PromAlarmHistoryFieldDuration,
					PromAlarmHistoryFieldEndAt,
					PromAlarmHistoryFieldStartAt,
					PromAlarmHistoryFieldStartAt,
				), m1, nowTimeUnix, nowTimeUnix-m1)
		case m5:
			return db.Where(
				fmt.Sprintf("(`%s` <= ? AND `%s` > ?) OR (`%s` = 0 AND `%s` <= ?  AND `%s` > ?)",
					PromAlarmHistoryFieldDuration,
					PromAlarmHistoryFieldDuration,
					PromAlarmHistoryFieldEndAt,
					PromAlarmHistoryFieldStartAt,
					PromAlarmHistoryFieldStartAt,
				), m5, m1, nowTimeUnix, nowTimeUnix-m5)
		case m30:
			return db.Where(
				fmt.Sprintf("(`%s` <= ? AND `%s` > ?) OR (`%s` = 0 AND `%s` <= ?  AND `%s` > ?)",
					PromAlarmHistoryFieldDuration,
					PromAlarmHistoryFieldDuration,
					PromAlarmHistoryFieldEndAt,
					PromAlarmHistoryFieldStartAt,
					PromAlarmHistoryFieldStartAt,
				), m30, m5, nowTimeUnix, nowTimeUnix-m30)
		case m30p1:
			return db.Where(
				fmt.Sprintf("(`%s` > ?) OR (`%s` = 0 AND `%s` < ?)",
					PromAlarmHistoryFieldDuration,
					PromAlarmHistoryFieldEndAt,
					PromAlarmHistoryFieldStartAt,
				), m30, nowTimeUnix-m30)
		}
	}
}

// PromAlarmHistoryClausesOnConflict 当索引冲突, 直接更新
func PromAlarmHistoryClausesOnConflict() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: PromAlarmHistoryFieldMd5}},
			DoUpdates: clause.AssignmentColumns([]string{
				basescopes.BaseFieldStatus.String(),
				PromAlarmHistoryFieldEndAt,
				PromAlarmHistoryFieldDuration,
				PromAlarmHistoryFieldInfo,
			}),
		})
	}
}

// PromAlarmHistoryPreloadStrategy 预加载策略
func PromAlarmHistoryPreloadStrategy() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromAlarmHistoryPreloadFieldStrategy)
	}
}

// PromAlarmHistoryPreloadLevel 预加载等级
func PromAlarmHistoryPreloadLevel() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromAlarmHistoryPreloadFieldLevel)
	}
}

// PromAlarmHistoryStartAtDesc 开始时间倒序
func PromAlarmHistoryStartAtDesc() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(fmt.Sprintf("`%s` DESC", PromAlarmHistoryFieldStartAt))
	}
}

// PromAlarmHistory 报警历史数据
type PromAlarmHistory struct {
	BaseModel
	Instance   string         `gorm:"column:instance;type:varchar(64);not null;comment:instance名称;index:idx__h__instance"`
	Status     vo.AlarmStatus `gorm:"column:status;type:varchar(16);not null;comment:告警消息状态, 报警和恢复"`
	Info       string         `gorm:"column:info;type:json;not null;comment:原始告警消息"`
	StartAt    int64          `gorm:"column:start_at;type:bigint;not null;comment:报警开始时间"`
	EndAt      int64          `gorm:"column:end_at;type:bigint;not null;comment:报警恢复时间"`
	Duration   int64          `gorm:"column:duration;type:bigint;not null;comment:持续时间时间戳, 没有恢复, 时间戳是0"`
	StrategyID uint32         `gorm:"column:strategy_id;type:int unsigned;not null;index:idx__h__strategy_id,priority:1;comment:规则ID, 用于查询时候"`
	LevelID    uint32         `gorm:"column:level_id;type:int unsigned;not null;index:idx__h__level_id,priority:1;comment:报警等级ID"`
	Md5        string         `gorm:"column:md5;type:char(32);not null;unique:idx__md5,priority:1;comment:md5"`

	Strategy *PromStrategy `gorm:"foreignKey:StrategyID"`
	Level    *SysDict      `gorm:"foreignKey:LevelID"`

	// 用于回顾告警历史时候的图表查询
	Expr       string `gorm:"column:expr;type:text;not null;comment:prom ql"`
	Datasource string `gorm:"column:datasource;type:varchar(255);not null;comment:数据源;default:''"`
}

// TableName PromAlarmHistory's table name
func (*PromAlarmHistory) TableName() string {
	return TableNamePromAlarmHistory
}

// GetStrategy 获取策略
func (p *PromAlarmHistory) GetStrategy() *PromStrategy {
	if p == nil {
		return nil
	}
	return p.Strategy
}

// GetLevel 获取等级
func (p *PromAlarmHistory) GetLevel() *SysDict {
	if p == nil {
		return nil
	}
	return p.Level
}
