package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

// NewAlarmGroupRepository 创建策略分组仓库
func NewAlarmGroupRepository(data *data.Data) repository.AlarmGroup {
	return &alarmGroupRepositoryImpl{
		data: data,
	}
}

type (
	alarmGroupRepositoryImpl struct {
		data *data.Data
	}
)

func (a *alarmGroupRepositoryImpl) CreateAlarmGroup(ctx context.Context, params *bo.CreateAlarmNoticeGroupParams) (*bizmodel.AlarmNoticeGroup, error) {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return nil, err
	}

	alarmGroupModel := createAlarmGroupParamsToModel(ctx, params)

	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.AlarmNoticeGroup.WithContext(ctx).Create(alarmGroupModel); err != nil {
			return err
		}
		noticeMembers := createAlarmNoticeUsersToModel(ctx, params.NoticeMembers, alarmGroupModel.ID)
		if err := tx.AlarmNoticeMember.WithContext(ctx).Create(noticeMembers...); err != nil {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	return alarmGroupModel, nil

}

func (a *alarmGroupRepositoryImpl) UpdateAlarmGroup(ctx context.Context, params *bo.UpdateAlarmNoticeGroupParams) error {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return err
	}
	noticeMembers := createAlarmNoticeUsersToModel(ctx, params.UpdateParam.NoticeMembers, params.ID)
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		//告警组关联通知人中间表操作
		groupModel := &bizmodel.AlarmNoticeGroup{AllFieldModel: model.AllFieldModel{ID: params.ID}}
		noticeParams := params.UpdateParam.NoticeMembers
		// 告警通知人与hook参数为空则清空
		if !types.IsNil(noticeParams) && len(noticeParams) > 0 {
			// 替换通知人员关联信息
			if err := tx.AlarmNoticeGroup.NoticeMembers.Model(groupModel).Replace(noticeMembers...); err != nil {
				return err
			}
		} else {
			// 清除通知人员关联信息
			if _, err := tx.AlarmNoticeMember.WithContext(ctx).Where(tx.AlarmNoticeMember.AlarmGroupID.Eq(params.ID)).Delete(); err != nil {
				return err
			}
		}

		if !types.IsNil(params.UpdateParam.HookIds) {
			// 替换告警hook关联信息
			hookModels := types.SliceTo(params.UpdateParam.HookIds, func(hookID uint32) *bizmodel.AlarmHook {
				return &bizmodel.AlarmHook{AllFieldModel: model.AllFieldModel{ID: hookID}}
			})
			if err := tx.AlarmNoticeGroup.AlarmHooks.Model(groupModel).Replace(hookModels...); err != nil {
				return err
			}
		} else {
			// 清除告警hook信息
			if _, err := tx.AlarmHook.WithContext(ctx).Where(tx.AlarmHook.AlarmGroupID.Eq(params.ID)).Delete(); err != nil {
				return err
			}
		}

		// 更新告警分组
		if _, err = tx.AlarmNoticeGroup.WithContext(ctx).Where(tx.AlarmNoticeGroup.ID.Eq(params.ID)).UpdateSimple(
			tx.AlarmNoticeGroup.Name.Value(params.UpdateParam.Name),
			tx.AlarmNoticeGroup.Remark.Value(params.UpdateParam.Remark),
		); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

func (a *alarmGroupRepositoryImpl) DeleteAlarmGroup(ctx context.Context, alarmID uint32) error {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return err
	}
	return bizQuery.Transaction(func(tx *bizquery.Query) error {
		// 清除通知人员关联信息
		if _, err := tx.AlarmNoticeMember.WithContext(ctx).Where(tx.AlarmNoticeMember.AlarmGroupID.Eq(alarmID)).Delete(); err != nil {
			return err
		}

		// 清除告警hook信息
		if _, err := tx.AlarmHook.WithContext(ctx).Where(tx.AlarmHook.AlarmGroupID.Eq(alarmID)).Delete(); err != nil {
			return err
		}

		if _, err = tx.AlarmNoticeGroup.WithContext(ctx).Where(bizQuery.AlarmNoticeGroup.ID.Eq(alarmID)).Delete(); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

func (a *alarmGroupRepositoryImpl) GetAlarmGroup(ctx context.Context, alarmID uint32) (*bizmodel.AlarmNoticeGroup, error) {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return bizQuery.AlarmNoticeGroup.WithContext(ctx).Where(bizQuery.AlarmNoticeGroup.ID.Eq(alarmID)).Preload(field.Associations).First()
}

func (a *alarmGroupRepositoryImpl) AlarmGroupPage(ctx context.Context, params *bo.QueryAlarmNoticeGroupListParams) ([]*bizmodel.AlarmNoticeGroup, error) {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.AlarmNoticeGroup.WithContext(ctx)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Name) {
		wheres = append(wheres, bizQuery.AlarmNoticeGroup.Name.Like(params.Name))
	}

	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.AlarmNoticeGroup.Status.Eq(params.Status.GetValue()))
	}
	if !types.TextIsNull(params.Keyword) {
		bizWrapper = bizWrapper.Or(bizQuery.AlarmNoticeGroup.Name.Like(params.Keyword))
		bizWrapper = bizWrapper.Or(bizQuery.AlarmNoticeGroup.Remark.Like(params.Keyword))
	}
	bizWrapper = bizWrapper.Where(wheres...)

	if err := types.WithPageQuery[bizquery.IAlarmNoticeGroupDo](bizWrapper, params.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(bizQuery.AlarmNoticeGroup.ID.Desc()).Find()
}

func (a *alarmGroupRepositoryImpl) UpdateStatus(ctx context.Context, params *bo.UpdateAlarmNoticeGroupStatusParams) error {
	if len(params.IDs) == 0 {
		return nil
	}

	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return err
	}

	_, err = bizQuery.AlarmNoticeGroup.WithContext(ctx).Where(bizQuery.AlarmNoticeGroup.ID.In(params.IDs...)).Update(bizQuery.AlarmNoticeGroup.Status, params.Status)
	return nil
}

func (a *alarmGroupRepositoryImpl) MyAlarmGroups(ctx context.Context, params *bo.MyAlarmGroupListParams) ([]*bizmodel.AlarmNoticeGroup, error) {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return nil, err
	}
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	memberID := claims.GetMember()

	// 查询当前账号告警通知
	alarmNotices, err := bizQuery.AlarmNoticeMember.WithContext(ctx).Where(bizQuery.AlarmNoticeMember.MemberID.In(memberID)).Find()
	if !types.IsNil(err) || len(alarmNotices) == 0 {
		return nil, err
	}

	var alarmGroupIds []uint32

	alarmGroupIds = append(alarmGroupIds, types.SliceTo(alarmNotices, func(member *bizmodel.AlarmNoticeMember) uint32 { return member.AlarmGroupID })...)

	bizWrapper := bizQuery.AlarmNoticeGroup.WithContext(ctx)
	var wheres []gen.Condition

	wheres = append(wheres, bizQuery.AlarmNoticeGroup.ID.In(alarmGroupIds...))

	if !types.TextIsNull(params.Name) {
		wheres = append(wheres, bizQuery.AlarmNoticeGroup.Name.Like(params.Name))
	}

	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.AlarmNoticeGroup.Status.Eq(params.Status.GetValue()))
	}
	if !types.TextIsNull(params.Keyword) {
		bizWrapper = bizWrapper.Or(bizQuery.AlarmNoticeGroup.Name.Like(params.Keyword))
		bizWrapper = bizWrapper.Or(bizQuery.AlarmNoticeGroup.Remark.Like(params.Keyword))
	}
	bizWrapper = bizWrapper.Where(wheres...)
	if err := types.WithPageQuery[bizquery.IAlarmNoticeGroupDo](bizWrapper, params.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(bizQuery.AlarmNoticeGroup.ID.Desc()).Find()
}

// convert bo params to model
func createAlarmGroupParamsToModel(ctx context.Context, params *bo.CreateAlarmNoticeGroupParams) *bizmodel.AlarmNoticeGroup {
	alarmGroup := &bizmodel.AlarmNoticeGroup{
		Name:   params.Name,
		Status: params.Status,
		Remark: params.Remark,
		AlarmHooks: types.SliceTo(params.HookIds, func(hookID uint32) *bizmodel.AlarmHook {
			return &bizmodel.AlarmHook{AllFieldModel: model.AllFieldModel{ID: hookID}}
		}),
	}
	alarmGroup.WithContext(ctx)
	return alarmGroup
}

func createAlarmNoticeUsersToModel(ctx context.Context, params []*bo.CreateNoticeMemberParams, alarmGroupID uint32) []*bizmodel.AlarmNoticeMember {
	return types.SliceToWithFilter(params, func(noticeUser *bo.CreateNoticeMemberParams) (*bizmodel.AlarmNoticeMember, bool) {
		if noticeUser.MemberID <= 0 {
			return nil, false
		}
		resUser := &bizmodel.AlarmNoticeMember{
			AlarmNoticeType: noticeUser.NotifyType,
			MemberID:        noticeUser.MemberID,
			AlarmGroupID:    alarmGroupID,
		}
		resUser.WithContext(ctx)
		return resUser, true
	})
}
