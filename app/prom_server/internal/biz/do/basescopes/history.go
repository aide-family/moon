package basescopes

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	HistoryTableFieldInstance    Field = "instance"
	HistoryTableFieldStartAt     Field = "start_at"
	HistoryTableFieldEndAt       Field = "end_at"
	HistoryTableFieldMD5         Field = "md5"
	HistoryTableFieldAlarmPageID Field = "alarm_page_id"
	HistoryTableFieldDuration    Field = "duration"
	HistoryTableFieldInfo        Field = "info"
)

// LikeInstance 根据字典名称模糊查询
func LikeInstance(keyword string) ScopeMethod {
	return WhereLikePrefixKeyword(keyword, HistoryTableFieldInstance)
}

// TimeRange 根据时间范围查询
func TimeRange(startTime, endTime int64) ScopeMethod {
	return BetweenColumn(HistoryTableFieldStartAt, startTime, endTime)
}

// WhereInMd5 根据md5查询
func WhereInMd5(md5s ...string) ScopeMethod {
	return WhereInColumn(HistoryTableFieldMD5, md5s...)
}

// WhereAlarmPages 根据告警页面查询
func WhereAlarmPages(ids []uint) ScopeMethod {
	return WhereInColumn(HistoryTableFieldAlarmPageID, ids...)
}

// ClausesOnConflict 当索引冲突, 直接更新
func ClausesOnConflict() ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: HistoryTableFieldMD5.String()}},
			DoUpdates: clause.AssignmentColumns([]string{
				BaseFieldStatus.String(),
				HistoryTableFieldEndAt.String(),
				HistoryTableFieldDuration.String(),
				HistoryTableFieldInfo.String(),
			}),
		})
	}
}
