package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
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
