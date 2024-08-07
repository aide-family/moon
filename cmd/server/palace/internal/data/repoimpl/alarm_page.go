package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen/field"
)

// NewAlarmPageRepository 创建告警页面管理操作
func NewAlarmPageRepository(data *data.Data) repository.AlarmPage {
	return &alarmPageRepositoryImpl{data: data}
}

type alarmPageRepositoryImpl struct {
	data *data.Data
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
		Preload(field.Associations).
		Find()
	if err != nil {
		return err
	}
	oldAlarmPageSelfList, err := bizQuery.AlarmPageSelf.WithContext(ctx).
		Where(bizQuery.AlarmPageSelf.UserID.Eq(userID)).
		Preload(field.Associations).
		Find()
	if err != nil {
		return err
	}
	alarmPageListMap := types.ToMap(alarmPageList, func(item *bizmodel.SysDict) uint32 {
		return item.ID
	})
	oldAlarmPageSelfListMap := types.ToMap(oldAlarmPageSelfList, func(item *bizmodel.AlarmPageSelf) uint32 {
		return item.AlarmPageID
	})
	addAlarmPageSelfList := make([]*bizmodel.AlarmPageSelf, 0, len(alarmPageIDs))
	deleteAlarmPageSelfList := make([]uint32, 0, len(oldAlarmPageSelfList))
	for sort, alarmPageID := range alarmPageIDs {
		// 如果告警页面不存在，则删除
		if _, ok := alarmPageListMap[alarmPageID]; !ok {
			deleteAlarmPageSelfList = append(deleteAlarmPageSelfList, alarmPageID)
			continue
		}
		// 如果告警页面存在，则跳过
		if _, ok := oldAlarmPageSelfListMap[alarmPageID]; ok {
			continue
		}
		// 如果告警页面不存在，则添加
		addAlarmPageSelfList = append(addAlarmPageSelfList, &bizmodel.AlarmPageSelf{
			UserID:      userID,
			MemberID:    memberDetail.ID,
			Sort:        uint32(sort),
			AlarmPageID: alarmPageID,
		})
	}
	if len(deleteAlarmPageSelfList) == 0 && len(addAlarmPageSelfList) == 0 {
		return nil
	}

	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		if len(deleteAlarmPageSelfList) > 0 {
			_, err = tx.AlarmPageSelf.WithContext(ctx).
				Where(tx.AlarmPageSelf.UserID.Eq(userID), tx.AlarmPageSelf.AlarmPageID.In(deleteAlarmPageSelfList...)).
				Delete()
			if err != nil {
				return err
			}
		}
		if len(addAlarmPageSelfList) > 0 {
			if err = tx.AlarmPageSelf.WithContext(ctx).Create(addAlarmPageSelfList...); err != nil {
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
