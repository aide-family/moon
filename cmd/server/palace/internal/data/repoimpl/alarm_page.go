package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gorm/clause"

	"gorm.io/gen/field"
)

// NewAlarmPageRepository 创建告警页面管理操作
func NewAlarmPageRepository(data *data.Data) repository.AlarmPage {
	return &alarmPageRepositoryImpl{data: data}
}

// alarmPageRepositoryImpl 告警页面管理操作
type alarmPageRepositoryImpl struct {
	data *data.Data
}

// AlertPageCount 告警页面统计
type AlertPageCount struct {
	PageID uint32 `gorm:"column:page_id"`
	Count  int64  `gorm:"column:count"`
}

// GetAlertCounts 获取告警数量
func (a *alarmPageRepositoryImpl) GetAlertCounts(ctx context.Context, pageIDs []uint32) map[int32]int64 {
	if len(pageIDs) == 0 {
		return nil
	}

	// 统计实时告警这些等级的告警数量
	bizAlarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil
	}
	realtimeAlarmPageQuery := bizAlarmQuery.RealtimeAlarmPage
	realtimeAlarmList, err := realtimeAlarmPageQuery.WithContext(ctx).Where(realtimeAlarmPageQuery.PageID.In(pageIDs...)).
		Select(realtimeAlarmPageQuery.PageID, realtimeAlarmPageQuery.RealtimeAlarmID).
		Group(realtimeAlarmPageQuery.RealtimeAlarmID).
		Find()
	if err != nil {
		return nil
	}

	alertCounts := make(map[int32]int64, len(realtimeAlarmList))
	for _, item := range realtimeAlarmList {
		alertCounts[int32(item.PageID)] += 1
	}
	alertCounts[-1] = a.countMyAlarm(ctx)

	return alertCounts
}

// countMyAlarm 统计我的告警数量
func (a *alarmPageRepositoryImpl) countMyAlarm(ctx context.Context) int64 {
	bizQuery, err := getBizQuery(ctx, a.data)
	if err != nil {
		return 0
	}

	var alarmNoticeGroupIDs []uint32
	if err = bizQuery.AlarmNoticeMember.WithContext(ctx).
		Where(bizQuery.AlarmNoticeMember.MemberID.Eq(middleware.GetTeamMemberID(ctx))).
		Select(bizQuery.AlarmNoticeMember.AlarmGroupID).Group(bizQuery.AlarmNoticeMember.AlarmGroupID).
		Scan(&alarmNoticeGroupIDs); err != nil {
		return 0
	}
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return 0
	}
	var realtimeAlarmIDs []uint32
	if err = alarmQuery.WithContext(ctx).RealtimeAlarmReceiver.
		Where(alarmQuery.RealtimeAlarmReceiver.AlarmNoticeGroupID.In(alarmNoticeGroupIDs...)).
		Select(alarmQuery.RealtimeAlarmReceiver.RealtimeAlarmID).Group(alarmQuery.RealtimeAlarmReceiver.RealtimeAlarmID).
		Scan(&realtimeAlarmIDs); err != nil {
		return 0
	}
	count, err := alarmQuery.WithContext(ctx).RealtimeAlarm.
		Where(alarmQuery.RealtimeAlarm.Status.Eq(vobj.AlertStatusFiring.GetValue())).
		Where(alarmQuery.RealtimeAlarm.ID.In(realtimeAlarmIDs...)).Count()
	if err != nil {
		return 0
	}
	return count
}

// ReplaceAlarmPages 替换告警页面
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

// ListAlarmPages 获取告警页面
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
