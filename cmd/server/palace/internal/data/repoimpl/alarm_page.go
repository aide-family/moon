package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gorm/clause"

	"gorm.io/gen/field"
)

// NewAlarmPageRepository 创建告警页面管理操作
func NewAlarmPageRepository(data *data.Data) repository.AlarmPage {
	return &alarmPageRepositoryImpl{data: data}
}

type alarmPageRepositoryImpl struct {
	data *data.Data
}

type AlertLevelCount struct {
	LevelID uint32 `gorm:"column:level_id"`
	Count   int64  `gorm:"column:count"`
}

func (a *alarmPageRepositoryImpl) GetAlertCounts(ctx context.Context, pageIDs []uint32) map[uint32]int64 {
	bizQuery, err := getBizQuery(ctx, a.data)
	if err != nil {
		return nil
	}

	alarmPageSelfQuery := bizQuery.SysDict
	alarmPageSelves, err := alarmPageSelfQuery.WithContext(ctx).
		Where(alarmPageSelfQuery.ID.In(pageIDs...)).
		Where(alarmPageSelfQuery.DictType.Eq(vobj.DictTypeAlarmPage.GetValue())).
		Preload(alarmPageSelfQuery.StrategyLevels).Find()
	if err != nil {
		return nil
	}
	pageLevelMap := make(map[uint32][]uint32, len(alarmPageSelves))
	levelIds := make([]uint32, 0, len(alarmPageSelves)*3)
	for _, alarmPageSelf := range alarmPageSelves {
		pageLevels := types.SliceTo(alarmPageSelf.StrategyLevels, func(item *bizmodel.StrategyLevel) uint32 {
			return item.ID
		})
		pageLevelMap[alarmPageSelf.ID] = pageLevels
		levelIds = append(levelIds, pageLevels...)
	}

	if len(levelIds) == 0 {
		return nil
	}
	// 统计实时告警这些等级的告警数量
	var alertLevelCounts []AlertLevelCount
	bizAlarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil
	}
	realtimeAlarmQuery := bizAlarmQuery.RealtimeAlarm
	err = realtimeAlarmQuery.WithContext(ctx).Where(realtimeAlarmQuery.LevelID.In(levelIds...)).
		Select(realtimeAlarmQuery.LevelID, realtimeAlarmQuery.LevelID.Count().As("count")).
		Group(realtimeAlarmQuery.LevelID).
		Scan(&alertLevelCounts)
	if err != nil {
		return nil
	}

	alertLevelCountsMap := make(map[uint32]int64, len(alertLevelCounts))
	for _, alertLevelCount := range alertLevelCounts {
		alertLevelCountsMap[alertLevelCount.LevelID] = alertLevelCount.Count
	}

	alertCounts := make(map[uint32]int64, len(alarmPageSelves))
	for pageID, pageLevels := range pageLevelMap {
		alertCounts[pageID] = 0
		for _, pageLevel := range pageLevels {
			if count, ok := alertLevelCountsMap[pageLevel]; ok {
				alertCounts[pageID] += count
			}
		}
	}

	return alertCounts
}

func (a *alarmPageRepositoryImpl) ReplaceAlarmPages(ctx context.Context, userID uint32, alarmPageIDs []uint32) error {
	bizQuery, err := getBizQuery(ctx, a.data)
	if err != nil {
		return err
	}
	if len(alarmPageIDs) == 0 {
		_, err = bizQuery.AlarmPageSelf.WithContext(ctx).
			Where(bizQuery.AlarmPageSelf.UserID.Eq(userID)).
			Delete()
		return err
	}
	memberDetail, err := bizQuery.SysTeamMember.WithContext(ctx).
		Where(bizQuery.SysTeamMember.UserID.Eq(userID)).
		Preload(field.Associations).
		First()
	if err != nil {
		return err
	}
	alarmPageList, err := bizQuery.SysDict.WithContext(ctx).
		Where(bizQuery.SysDict.ID.In(alarmPageIDs...)).
		Where(bizQuery.SysDict.DictType.Eq(vobj.DictTypeAlarmPage.GetValue())).
		Preload(field.Associations).
		Find()
	if err != nil {
		return err
	}
	if len(alarmPageList) != len(alarmPageIDs) {
		return merr.ErrorI18nAlertSelectAlertPageErr(ctx)
	}

	addAlarmPageSelfList := make([]*bizmodel.AlarmPageSelf, 0, len(alarmPageIDs))
	for sort, alarmPageID := range alarmPageIDs {
		// 如果告警页面不存在，则添加
		addAlarmPageSelfList = append(addAlarmPageSelfList, &bizmodel.AlarmPageSelf{
			UserID:      userID,
			MemberID:    memberDetail.ID,
			Sort:        uint32(sort),
			AlarmPageID: alarmPageID,
		})
	}

	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		_, err = tx.AlarmPageSelf.WithContext(ctx).Where(tx.AlarmPageSelf.UserID.Eq(userID)).Delete()
		if err != nil {
			return err
		}
		if len(addAlarmPageSelfList) > 0 {
			if err = tx.AlarmPageSelf.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(addAlarmPageSelfList...); err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *alarmPageRepositoryImpl) ListAlarmPages(ctx context.Context, userID uint32) ([]*bizmodel.AlarmPageSelf, error) {
	bizQuery, err := getBizQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	return bizQuery.AlarmPageSelf.WithContext(ctx).
		Where(bizQuery.AlarmPageSelf.UserID.Eq(userID)).
		Preload(field.Associations).
		Find()
}
