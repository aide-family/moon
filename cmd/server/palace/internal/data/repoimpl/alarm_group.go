package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

// NewAlarmGroupRepository 创建策略分组仓库
func NewAlarmGroupRepository(data *data.Data, rabbitConn *data.RabbitConn) repository.AlarmGroup {
	return &alarmGroupRepositoryImpl{
		data:       data,
		rabbitConn: rabbitConn,
	}
}

type (
	alarmGroupRepositoryImpl struct {
		data       *data.Data
		rabbitConn *data.RabbitConn
	}
)

func (a *alarmGroupRepositoryImpl) checkHooks(ctx context.Context, hookIds []uint32) error {
	if len(hookIds) == 0 {
		return nil
	}
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return err
	}
	// 校验hook是否存在
	hookCount, err := bizQuery.AlarmHook.WithContext(ctx).
		Where(bizQuery.AlarmHook.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.AlarmHook.ID.In(hookIds...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(hookCount) != len(hookIds) {
		return merr.ErrorI18nToastAlarmHookNotFound(ctx)
	}
	return nil
}

func (a *alarmGroupRepositoryImpl) checkMembers(ctx context.Context, memberIds []uint32) error {
	if len(memberIds) == 0 {
		return nil
	}
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return err
	}
	// 校验成员是否存在
	memberCount, err := bizQuery.SysTeamMember.WithContext(ctx).
		Where(bizQuery.SysTeamMember.Status.Eq(vobj.StatusEnable.GetValue())).
		Where(bizQuery.SysTeamMember.ID.In(memberIds...)).Count()
	if !types.IsNil(err) {
		return err
	}
	if int(memberCount) != len(memberIds) {
		return merr.ErrorI18nToastTeamMemberNotFound(ctx)
	}
	return nil
}

// 校验告警组名称是否存在
func (a *alarmGroupRepositoryImpl) checkAlarmGroupName(ctx context.Context, name string, id uint32) error {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return err
	}
	alarmGroupDo, err := bizQuery.AlarmNoticeGroup.WithContext(ctx).Where(bizQuery.AlarmNoticeGroup.Name.Eq(name)).First()
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if (id > 0 && alarmGroupDo.ID != id) || id == 0 {
		return merr.ErrorI18nAlertAlertGroupNameDuplicate(ctx)
	}
	return nil
}

func (a *alarmGroupRepositoryImpl) CreateAlarmGroup(ctx context.Context, params *bo.CreateAlarmNoticeGroupParams) (*bizmodel.AlarmNoticeGroup, error) {
	if err := a.checkAlarmGroupName(ctx, params.Name, 0); !types.IsNil(err) {
		return nil, err
	}
	if err := a.checkHooks(ctx, params.HookIds); !types.IsNil(err) {
		return nil, err
	}
	memberIds := types.SliceTo(params.NoticeMembers, func(member *bo.CreateNoticeMemberParams) uint32 {
		return member.MemberID
	})
	if err := a.checkMembers(ctx, memberIds); !types.IsNil(err) {
		return nil, err
	}

	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return nil, err
	}

	alarmGroupModel := createAlarmGroupParamsToModel(ctx, params)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.AlarmNoticeGroup.WithContext(ctx).Create(alarmGroupModel); err != nil {
			return err
		}
		if len(params.NoticeMembers) > 0 {
			noticeMembers := createAlarmNoticeUsersToModel(ctx, params.NoticeMembers, alarmGroupModel.ID)
			if err := tx.AlarmNoticeMember.WithContext(ctx).Create(noticeMembers...); err != nil {
				return err
			}
		}
		return nil
	})
	if !types.IsNil(err) {
		return nil, err
	}
	if len(params.TimeEngineIds) > 0 {
		timeEngines := types.SliceTo(params.TimeEngineIds, func(timeEngineID uint32) *bizmodel.TimeEngine {
			return &bizmodel.TimeEngine{AllFieldModel: bizmodel.AllFieldModel{
				AllFieldModel: model.AllFieldModel{ID: timeEngineID},
				TeamID:        middleware.GetTeamID(ctx),
			}}
		})
		if err := bizQuery.AlarmNoticeGroup.TimeEngines.Model(alarmGroupModel).Append(timeEngines...); err != nil {
			return nil, err
		}
	}
	go func() {
		defer after.RecoverX()
		ctx := types.CopyValueCtx(ctx)
		if err := a.rabbitConn.SyncTeam(ctx, middleware.GetTeamID(ctx)); !types.IsNil(err) {
			log.Errorw("method", "SyncTeam", "error", err)
		}
	}()
	return alarmGroupModel, nil
}

func (a *alarmGroupRepositoryImpl) UpdateAlarmGroup(ctx context.Context, params *bo.UpdateAlarmNoticeGroupParams) error {
	if params.UpdateParam == nil {
		panic("UpdateAlarmGroup method params UpdateParam field is nil")
	}
	if err := a.checkAlarmGroupName(ctx, params.UpdateParam.Name, params.ID); !types.IsNil(err) {
		return err
	}
	memberIds := types.SliceTo(params.UpdateParam.NoticeMembers, func(member *bo.CreateNoticeMemberParams) uint32 {
		return member.MemberID
	})
	if err := a.checkMembers(ctx, memberIds); !types.IsNil(err) {
		return err
	}

	if err := a.checkHooks(ctx, params.UpdateParam.HookIds); !types.IsNil(err) {
		return err
	}

	bizDB, err := a.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}
	noticeMembers := createAlarmNoticeUsersToModel(ctx, params.UpdateParam.NoticeMembers, params.ID)
	// 替换告警hook关联信息
	hookModels := types.SliceTo(params.UpdateParam.HookIds, func(hookID uint32) *bizmodel.AlarmHook {
		return &bizmodel.AlarmHook{AllFieldModel: bizmodel.AllFieldModel{
			AllFieldModel: model.AllFieldModel{ID: hookID},
			TeamID:        middleware.GetTeamID(ctx),
		}}
	})
	// 告警组关联通知人中间表操作
	groupModel := &bizmodel.AlarmNoticeGroup{AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: params.ID}}}
	defer func() {
		go func() {
			defer after.RecoverX()
			ctx := types.CopyValueCtx(ctx)
			if err := a.rabbitConn.SyncTeam(ctx, middleware.GetTeamID(ctx)); !types.IsNil(err) {
				log.Errorw("method", "SyncTeam", "error", err)
			}
		}()
		if len(params.UpdateParam.TimeEngineIds) > 0 {
			timeEngines := types.SliceTo(params.UpdateParam.TimeEngineIds, func(timeEngineID uint32) *bizmodel.TimeEngine {
				return &bizmodel.TimeEngine{AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: timeEngineID}}}
			})
			bizQuery, err := getBizQuery(ctx, a.data)
			if !types.IsNil(err) {
				return
			}
			if err := bizQuery.AlarmNoticeGroup.TimeEngines.Model(groupModel).Replace(timeEngines...); err != nil {
				return
			}
		}
	}()
	return bizDB.Transaction(func(tx *gorm.DB) error {
		bizQueryTx := bizquery.Use(tx)
		// 告警通知人与hook参数为空则清空
		if len(noticeMembers) > 0 {
			// 清除通知人员关联信息
			if _, err := bizQueryTx.AlarmNoticeMember.WithContext(ctx).Where(bizQueryTx.AlarmNoticeMember.AlarmGroupID.Eq(params.ID)).Delete(); err != nil {
				return err
			}

			// 替换通知人员关联信息
			if err := bizQueryTx.AlarmNoticeMember.WithContext(ctx).Create(noticeMembers...); err != nil {
				return err
			}
		} else {
			// 清除通知人员关联信息
			if _, err := bizQueryTx.AlarmNoticeMember.WithContext(ctx).Where(bizQueryTx.AlarmNoticeMember.AlarmGroupID.Eq(params.ID)).Delete(); err != nil {
				return err
			}
		}

		if len(hookModels) > 0 {
			if err := bizQueryTx.AlarmNoticeGroup.AlarmHooks.Model(groupModel).Replace(hookModels...); err != nil {
				return err
			}
		} else {
			// 清除告警hook信息
			if err := bizQueryTx.AlarmNoticeGroup.AlarmHooks.Model(groupModel).Clear(); err != nil {
				return err
			}
		}

		// 更新告警分组
		if _, err = bizQueryTx.AlarmNoticeGroup.WithContext(ctx).Where(bizQueryTx.AlarmNoticeGroup.ID.Eq(params.ID)).UpdateSimple(
			bizQueryTx.AlarmNoticeGroup.Name.Value(params.UpdateParam.Name),
			bizQueryTx.AlarmNoticeGroup.Remark.Value(params.UpdateParam.Remark),
		); !types.IsNil(err) {
			return err
		}
		return nil
	})
}

func (a *alarmGroupRepositoryImpl) DeleteAlarmGroup(ctx context.Context, alarmID uint32) error {
	bizDB, err := a.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}
	defer func() {
		go func() {
			defer after.RecoverX()
			ctx := types.CopyValueCtx(ctx)
			if err := a.rabbitConn.SyncTeam(ctx, middleware.GetTeamID(ctx)); !types.IsNil(err) {
				log.Errorw("method", "SyncTeam", "error", err)
			}
		}()
	}()
	return bizDB.Transaction(func(tx *gorm.DB) error {
		bizQueryTx := bizquery.Use(tx)
		// 清除通知人员关联信息
		if _, err = bizQueryTx.AlarmNoticeMember.WithContext(ctx).Where(bizQueryTx.AlarmNoticeMember.AlarmGroupID.Eq(alarmID)).Delete(); err != nil {
			return err
		}

		// 清除告警hook信息
		if err = bizQueryTx.AlarmNoticeGroup.AlarmHooks.WithContext(ctx).Model(&bizmodel.AlarmNoticeGroup{AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: alarmID}}}).Clear(); err != nil {
			return err
		}

		if _, err = bizQueryTx.AlarmNoticeGroup.WithContext(ctx).Where(bizQueryTx.AlarmNoticeGroup.ID.Eq(alarmID)).Delete(); !types.IsNil(err) {
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
	return bizQuery.AlarmNoticeGroup.WithContext(ctx).Where(bizQuery.AlarmNoticeGroup.ID.Eq(alarmID)).
		Preload(field.Associations, bizQuery.AlarmNoticeGroup.NoticeMembers.Member).
		Preload(bizQuery.AlarmNoticeGroup.TimeEngines).
		First()
}

func (a *alarmGroupRepositoryImpl) GetAlarmGroupsByIDs(ctx context.Context, ids []uint32) ([]*bizmodel.AlarmNoticeGroup, error) {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return bizQuery.AlarmNoticeGroup.WithContext(ctx).Where(bizQuery.AlarmNoticeGroup.ID.In(ids...)).
		Preload(field.Associations, bizQuery.AlarmNoticeGroup.NoticeMembers.Member).
		Preload(bizQuery.AlarmNoticeGroup.TimeEngines).
		Find()
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
	if bizWrapper, err = types.WithPageQuery(bizWrapper, params.Page); err != nil {
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
	if !types.IsNil(err) {
		return err
	}
	go func() {
		defer after.RecoverX()
		ctx := types.CopyValueCtx(ctx)
		if err := a.rabbitConn.SyncTeam(ctx, middleware.GetTeamID(ctx)); !types.IsNil(err) {
			log.Errorw("method", "SyncTeam", "error", err)
		}
	}()
	return nil
}

func (a *alarmGroupRepositoryImpl) MyAlarmGroups(ctx context.Context, params *bo.MyAlarmGroupListParams) ([]*bizmodel.AlarmNoticeGroup, error) {
	bizQuery, err := getBizQuery(ctx, a.data)
	if !types.IsNil(err) {
		return nil, err
	}

	memberID := middleware.GetTeamMemberID(ctx)

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
	if bizWrapper, err = types.WithPageQuery(bizWrapper, params.Page); err != nil {
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
			return &bizmodel.AlarmHook{AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: hookID}}}
		}),
	}
	alarmGroup.WithContext(ctx)
	return alarmGroup
}

func createAlarmNoticeUsersToModel(ctx context.Context, params []*bo.CreateNoticeMemberParams, alarmGroupID uint32) []*bizmodel.AlarmNoticeMember {
	if len(params) == 0 {
		return nil
	}
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
