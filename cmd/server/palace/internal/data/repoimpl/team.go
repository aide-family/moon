package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/random"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// NewTeamRepository 创建团队仓库
func NewTeamRepository(data *data.Data, cacheRepo repository.Cache) repository.Team {
	return &teamRepositoryImpl{
		data:      data,
		cacheRepo: cacheRepo,
	}
}

type teamRepositoryImpl struct {
	data      *data.Data
	cacheRepo repository.Cache
}

func (l *teamRepositoryImpl) MemberList(ctx context.Context, teamID uint32) ([]*bizmodel.SysTeamMember, error) {
	bizQuery, err := getTeamIDBizQuery(l.data, teamID)
	if !types.IsNil(err) {
		return nil, err
	}
	return bizQuery.SysTeamMember.WithContext(ctx).Find()
}

func (l *teamRepositoryImpl) GetTeamConfig(ctx context.Context, teamID uint32) (*model.SysTeamConfig, error) {
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	return mainQuery.WithContext(ctx).SysTeamConfig.Where(mainQuery.SysTeamConfig.TeamID.Eq(teamID)).First()
}

func (l *teamRepositoryImpl) CreateTeamConfig(ctx context.Context, params *bo.SetTeamConfigParams) error {
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	return mainQuery.WithContext(ctx).SysTeamConfig.Create(params.ToModel(ctx))
}

func (l *teamRepositoryImpl) UpdateTeamConfig(ctx context.Context, params *bo.SetTeamConfigParams) error {
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	teamConfig := params.ToModel(ctx)
	rows, err := mainQuery.WithContext(ctx).SysTeamConfig.
		Where(mainQuery.SysTeamConfig.TeamID.Eq(middleware.GetTeamID(ctx))).
		UpdateColumnSimple(
			mainQuery.SysTeamConfig.EmailConfig.Value(teamConfig.EmailConfig),
			mainQuery.SysTeamConfig.SymmetricEncryptionConfig.Value(teamConfig.SymmetricEncryptionConfig),
			mainQuery.SysTeamConfig.AsymmetricEncryptionConfig.Value(teamConfig.AsymmetricEncryptionConfig),
		)
	if !types.IsNil(err) {
		return err
	}
	if rows.RowsAffected == 0 {
		log.Warnw("UpdateTeamConfig.RowsAffected", "rows", rows)
	}
	return nil
}

func (l *teamRepositoryImpl) CreateTeam(ctx context.Context, team *bo.CreateTeamParams) (*model.SysTeam, error) {
	sysTeamModel := &model.SysTeam{
		Name:     team.Name,
		Status:   team.Status,
		Remark:   team.Remark,
		Logo:     team.Logo,
		LeaderID: team.LeaderID,
		UUID:     random.UUIDToUpperCase(true),
	}
	sysTeamModel.WithContext(ctx)
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	// 判断团队名称是否重复
	_, err := mainQuery.SysTeam.WithContext(ctx).Where(mainQuery.SysTeam.Name.Eq(team.Name)).First()
	if !types.IsNil(err) {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if err == nil {
		return nil, merr.ErrorI18nAlertTeamNameExistErr(ctx)
	}
	// 添加管理员成员
	adminMembers := types.SliceToWithFilter(team.Admins, func(memberId uint32) (*bizmodel.SysTeamMember, bool) {
		if memberId <= 0 {
			return nil, false
		}
		if memberId == sysTeamModel.LeaderID {
			return nil, false
		}
		return &bizmodel.SysTeamMember{
			UserID: memberId,
			Status: vobj.StatusEnable,
			Role:   vobj.RoleAdmin,
		}, true
	})
	adminMembers = append(adminMembers, &bizmodel.SysTeamMember{
		UserID: sysTeamModel.LeaderID,
		Status: vobj.StatusEnable,
		Role:   vobj.RoleSuperAdmin,
	})
	err = mainQuery.Transaction(func(tx *query.Query) error {
		// 创建基础信息
		if err = tx.SysTeam.Create(sysTeamModel); !types.IsNil(err) {
			return err
		}

		return l.syncTeamBaseData(ctx, sysTeamModel, adminMembers)
	})
	if !types.IsNil(err) {
		return nil, err
	}
	l.cacheRepo.AppendTeam(ctx, sysTeamModel)
	l.cacheRepo.SyncUserTeamList(ctx, sysTeamModel.LeaderID)
	for _, memberID := range team.Admins {
		l.cacheRepo.SyncUserTeamList(ctx, memberID)
	}
	return sysTeamModel, nil
}

func (l *teamRepositoryImpl) SyncTeamBaseData(ctx context.Context, sysTeamModel *model.SysTeam, members []*bizmodel.SysTeamMember) (err error) {
	return l.syncTeamBaseData(ctx, sysTeamModel, members)
}

func (l *teamRepositoryImpl) syncTeamBaseData(ctx context.Context, sysTeamModel *model.SysTeam, members []*bizmodel.SysTeamMember) (err error) {
	teamID := sysTeamModel.ID
	// 创建团队数据库
	if err = l.data.CreateBizDatabase(teamID); !types.IsNil(err) {
		return err
	}

	// 创建团队数据库
	if err = l.data.CreateBizAlarmDatabase(teamID); !types.IsNil(err) {
		return err
	}

	mainQuery := query.Use(l.data.GetMainDB(ctx))
	sysApis, err := mainQuery.WithContext(ctx).SysAPI.Find()
	if !types.IsNil(err) {
		return err
	}

	sysDict, err := mainQuery.WithContext(ctx).SysDict.Find()
	if !types.IsNil(err) {
		return err
	}
	teamApis := types.SliceToWithFilter(sysApis, func(apiItem *model.SysAPI) (*bizmodel.SysTeamAPI, bool) {
		return &bizmodel.SysTeamAPI{
			Name:   apiItem.Name,
			Path:   apiItem.Path,
			Status: apiItem.Status,
			Remark: apiItem.Remark,
			Module: apiItem.Module,
			Domain: apiItem.Domain,
		}, true
	})

	dictList := types.SliceToWithFilter(sysDict, func(dictItem *model.SysDict) (*bizmodel.SysDict, bool) {
		return &bizmodel.SysDict{
			Name:         dictItem.Name,
			Value:        dictItem.Value,
			DictType:     dictItem.DictType,
			ColorType:    dictItem.ColorType,
			CSSClass:     dictItem.CSSClass,
			Icon:         dictItem.Icon,
			ImageURL:     dictItem.ImageURL,
			Status:       dictItem.Status,
			LanguageCode: dictItem.LanguageCode,
			Remark:       dictItem.Remark,
		}, true
	})

	bizDB, bizDbErr := l.data.GetBizGormDB(teamID)
	if !types.IsNil(bizDbErr) {
		return err
	}
	alarmDB, alarmDbErr := l.data.GetAlarmGormDB(teamID)
	if !types.IsNil(alarmDbErr) {
		return err
	}

	alarmModels := alarmmodel.Models()
	if len(alarmModels) > 0 {
		if err = alarmDB.AutoMigrate(alarmModels...); !types.IsNil(err) {
			return err
		}
	}

	modelList := bizmodel.Models()
	if len(modelList) > 0 {
		// 初始化数据表
		if err = bizDB.AutoMigrate(modelList...); !types.IsNil(err) {
			return err
		}

		return bizDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			bizQuery := bizquery.Use(tx)
			if len(members) > 0 {
				if err = bizQuery.SysTeamMember.Create(members...); !types.IsNil(err) {
					return err
				}
			}

			if len(teamApis) > 0 {
				// 迁移api数据到团队数据库
				if err = bizQuery.SysTeamAPI.Create(teamApis...); !types.IsNil(err) {
					return err
				}
			}

			if len(dictList) > 0 {
				if err = bizQuery.SysDict.Create(dictList...); !types.IsNil(err) {
					return err
				}
			}

			return nil
		})
	}

	return nil
}

func (l *teamRepositoryImpl) UpdateTeam(ctx context.Context, team *bo.UpdateTeamParams) error {
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	// 判断团队名称是否重复
	teamInfo, err := mainQuery.SysTeam.WithContext(ctx).Where(mainQuery.SysTeam.Name.Eq(team.Name)).First()
	if !types.IsNil(err) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && teamInfo.ID != team.ID {
		return merr.ErrorI18nAlertTeamNameExistErr(ctx)
	}
	_, err = mainQuery.WithContext(ctx).SysTeam.
		Where(mainQuery.SysTeam.ID.Eq(team.ID)).
		UpdateColumnSimple(
			mainQuery.SysTeam.Name.Value(team.Name),
			mainQuery.SysTeam.Remark.Value(team.Remark),
			mainQuery.SysTeam.Logo.Value(team.Logo),
			mainQuery.SysTeam.Status.Value(team.Status.GetValue()),
		)
	return err
}

func (l *teamRepositoryImpl) GetTeamDetail(ctx context.Context, teamID uint32) (*model.SysTeam, error) {
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	return mainQuery.WithContext(ctx).SysTeam.
		Where(mainQuery.SysTeam.ID.Eq(teamID)).
		Preload(field.Associations).
		First()
}

func (l *teamRepositoryImpl) GetTeamList(ctx context.Context, params *bo.QueryTeamListParams) ([]*model.SysTeam, error) {
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	teamQuery := mainQuery.WithContext(ctx).SysTeam
	if types.IsNil(params) {
		return teamQuery.Find()
	}
	if !types.TextIsNull(params.Keyword) {
		teamQuery = teamQuery.Where(mainQuery.SysTeam.Name.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		teamQuery = teamQuery.Where(mainQuery.SysTeam.Status.Eq(params.Status.GetValue()))
	}
	if params.CreatorID > 0 {
		teamQuery = teamQuery.Where(mainQuery.SysTeam.CreatorID.Eq(params.CreatorID))
	}
	if params.LeaderID > 0 {
		teamQuery = teamQuery.Where(mainQuery.SysTeam.LeaderID.Eq(params.LeaderID))
	}
	var teamIds []uint32
	queryTeamIds := false
	if params.UserID > 0 {
		queryTeamIds = true
		// 缓存用户的全部团队ID， 然后取出来使用
		teamList := l.cacheRepo.GetUserTeamList(ctx, params.UserID)
		teamIds = append(teamIds, types.SliceTo(teamList, func(team *model.SysTeam) uint32 { return team.ID })...)
	}
	if len(params.IDs) > 0 {
		queryTeamIds = true
		if len(teamIds) > 0 {
			teamIds = types.SlicesIntersection(params.IDs, teamIds)
		} else {
			teamIds = params.IDs
		}
	}
	if queryTeamIds {
		teamQuery = teamQuery.Where(mainQuery.SysTeam.ID.In(teamIds...))
	}

	return teamQuery.Order(mainQuery.SysTeam.ID.Desc()).Preload(field.Associations).Find()
}

func (l *teamRepositoryImpl) UpdateTeamStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	_, err := mainQuery.WithContext(ctx).SysTeam.Where(mainQuery.SysTeam.ID.In(ids...)).
		UpdateColumnSimple(mainQuery.SysTeam.Status.Value(status.GetValue()))
	return err
}

func (l *teamRepositoryImpl) GetUserTeamList(ctx context.Context, userID uint32) ([]*model.SysTeam, error) {
	// 从全局缓存读取数据
	list := l.cacheRepo.GetUserTeamList(ctx, userID)
	if len(list) > 0 {
		return list, nil
	}
	l.cacheRepo.SyncUserTeamList(ctx, userID)
	return l.cacheRepo.GetUserTeamList(ctx, userID), nil
}

func (l *teamRepositoryImpl) AddTeamMember(ctx context.Context, params *bo.AddTeamMemberParams) error {
	members := types.SliceToWithFilter(params.Members, func(memberItem *bo.AddTeamMemberItem) (*bizmodel.SysTeamMember, bool) {
		if types.IsNil(memberItem) {
			return nil, false
		}
		return &bizmodel.SysTeamMember{
			UserID: memberItem.UserID,
			Status: vobj.StatusEnable,
			Role:   memberItem.Role,
			TeamRoles: types.SliceToWithFilter(memberItem.RoleIDs, func(roleId uint32) (*bizmodel.SysTeamRole, bool) {
				if roleId <= 0 {
					return nil, false
				}
				return &bizmodel.SysTeamRole{
					AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: roleId}},
				}, true
			}),
		}, true
	})
	bizDB, err := l.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}

	bizQuery := bizquery.Use(bizDB)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if err := tx.SysTeamMember.WithContext(ctx).Create(members...); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return err
	}
	for _, member := range members {
		l.cacheRepo.SyncUserTeamList(ctx, member.UserID)
	}

	return nil
}

func (l *teamRepositoryImpl) RemoveTeamMember(ctx context.Context, params *bo.RemoveTeamMemberParams) error {
	if len(params.MemberIds) == 0 {
		return nil
	}
	bizDB, err := l.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		if _, err = tx.SysTeamMember.WithContext(ctx).
			Where(tx.SysTeamMember.ID.In(params.MemberIds...)).
			Delete(); !types.IsNil(err) {
			return err
		}
		if _, err = tx.SysTeamMemberRole.WithContext(ctx).
			Where(tx.SysTeamMemberRole.SysTeamMemberID.In(params.MemberIds...)).
			Delete(); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return err
	}
	for _, member := range params.MemberIds {
		l.cacheRepo.SyncUserTeamList(ctx, member)
	}
	return nil
}

func (l *teamRepositoryImpl) SetMemberAdmin(ctx context.Context, params *bo.SetMemberAdminParams) error {
	if len(params.MemberIDs) == 0 {
		return nil
	}
	bizDB, err := l.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}

	bizQuery := bizquery.Use(bizDB)
	_, err = bizQuery.WithContext(ctx).SysTeamMember.Where(
		bizQuery.SysTeamMember.ID.In(params.MemberIDs...),
	).UpdateColumnSimple(bizQuery.SysTeamMember.Role.Value(params.Role.GetValue()))
	if !types.IsNil(err) {
		return err
	}
	// 查询团队管理员列表
	members, err := bizQuery.SysTeamMember.Where(
		bizQuery.SysTeamMember.ID.In(params.MemberIDs...),
	).Find()
	if !types.IsNil(err) {
		return err
	}
	for _, member := range members {
		l.cacheRepo.SyncUserTeamList(ctx, member.UserID)
	}
	return nil
}

func (l *teamRepositoryImpl) SetMemberRole(ctx context.Context, params *bo.SetMemberRoleParams) error {
	roles := types.SliceToWithFilter(params.RoleIDs, func(roleId uint32) (*bizmodel.SysTeamRole, bool) {
		if roleId == 0 {
			return nil, false
		}
		return &bizmodel.SysTeamRole{
			AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: roleId}},
		}, true
	})
	bizDB, err := l.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)
	return bizQuery.SysTeamMember.TeamRoles.
		WithContext(ctx).
		Model(&bizmodel.SysTeamMember{AllFieldModel: bizmodel.AllFieldModel{
			AllFieldModel: model.AllFieldModel{ID: params.MemberID},
		}}).
		Replace(roles...)
}

func (l *teamRepositoryImpl) ListTeamMember(ctx context.Context, params *bo.ListTeamMemberParams) ([]*bizmodel.SysTeamMember, error) {
	var wheres []gen.Condition
	var userWheres []gen.Condition
	bizDB, err := l.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	if !types.TextIsNull(params.Keyword) {
		userWheres = append(userWheres, query.SysUser.Username.Like(params.Keyword))
	}
	if !params.Gender.IsUnknown() {
		userWheres = append(userWheres, query.SysUser.Gender.Eq(params.Gender.GetValue()))
	}
	if len(userWheres) > 0 {
		var userIds []uint32
		mainQuery := query.Use(l.data.GetMainDB(ctx))
		if err = mainQuery.WithContext(ctx).SysUser.Where(userWheres...).
			Pluck(mainQuery.SysUser.ID, &userIds); !types.IsNil(err) {
			return nil, err
		}
		wheres = append(wheres, bizQuery.SysTeamMember.UserID.In(userIds...))
	}

	if !params.Role.IsAll() {
		wheres = append(wheres, bizQuery.SysTeamMember.Role.Eq(params.Role.GetValue()))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.SysTeamMember.Status.Eq(params.Status.GetValue()))
	}
	if len(params.MemberIDs) > 0 {
		wheres = append(wheres, bizQuery.SysTeamMember.ID.In(params.MemberIDs...))
	}

	q := bizQuery.WithContext(ctx).SysTeamMember.Where(wheres...)
	if q, err = types.WithPageQuery(q, params.Page); err != nil {
		return nil, err
	}
	return q.Order(bizQuery.SysTeamMember.Role.Desc(), bizQuery.SysTeamMember.ID.Asc()).Find()
}

func (l *teamRepositoryImpl) TransferTeamLeader(ctx context.Context, params *bo.TransferTeamLeaderParams) error {
	bizDB, err := l.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)
	err = bizQuery.Transaction(func(tx *bizquery.Query) error {
		// 设置新管理员
		if _, err = tx.SysTeamMember.WithContext(ctx).Where(
			tx.SysTeamMember.ID.Eq(params.LeaderID),
		).UpdateColumnSimple(tx.SysTeamMember.Role.Value(vobj.RoleSuperAdmin.GetValue())); !types.IsNil(err) {
			return err
		}
		// 设置老管理员
		if _, err = tx.SysTeamMember.WithContext(ctx).Where(
			tx.SysTeamMember.ID.Neq(params.OldLeaderID),
		).UpdateColumnSimple(tx.SysTeamMember.Role.Value(vobj.RoleAdmin.GetValue())); !types.IsNil(err) {
			return err
		}
		// 系统团队信息
		mainQuery := query.Use(l.data.GetMainDB(ctx))
		_, err = mainQuery.SysTeam.WithContext(ctx).Where(
			mainQuery.SysTeam.ID.Eq(middleware.GetTeamID(ctx)),
		).UpdateColumnSimple(mainQuery.SysTeam.LeaderID.Value(params.LeaderID))
		if !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return err
	}
	members, err := bizQuery.WithContext(ctx).SysTeamMember.Where(
		bizQuery.SysTeamMember.ID.In(params.OldLeaderID, params.LeaderID),
	).Find()
	if !types.IsNil(err) {
		return err
	}
	for _, member := range members {
		l.cacheRepo.SyncUserTeamList(ctx, member.UserID)
	}
	return nil
}

func (l *teamRepositoryImpl) GetUserTeamByID(ctx context.Context, userID, teamID uint32) (*bizmodel.SysTeamMember, error) {
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	return bizQuery.WithContext(ctx).SysTeamMember.Where(
		bizQuery.SysTeamMember.UserID.Eq(userID),
	).First()
}

func (l *teamRepositoryImpl) UpdateTeamMemberStatus(ctx context.Context, status vobj.Status, u ...uint32) error {
	bizDB, err := l.data.GetBizGormDB(middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)
	_, err = bizQuery.WithContext(ctx).SysTeamMember.Where(
		bizQuery.SysTeamMember.ID.In(u...),
	).UpdateColumnSimple(bizQuery.SysTeamMember.Status.Value(status.GetValue()))
	return err
}
