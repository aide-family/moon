package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/random"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gen/field"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// NewTeamRepository 创建团队仓库
func NewTeamRepository(data *data.Data, cache runtimecache.RuntimeCache) repository.Team {
	return &teamRepositoryImpl{
		data:  data,
		cache: cache,
	}
}

type teamRepositoryImpl struct {
	data  *data.Data
	cache runtimecache.RuntimeCache
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
	// 判断团队名称是否重复
	_, err := query.Use(l.data.GetMainDB(ctx)).SysTeam.WithContext(ctx).Where(query.SysTeam.Name.Eq(team.Name)).First()
	if !types.IsNil(err) {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if err == nil {
		return nil, merr.ErrorI18nTeamNameExistErr(ctx)
	}

	err = query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建基础信息
		if err = tx.SysTeam.Create(sysTeamModel); !types.IsNil(err) {
			return err
		}

		return l.syncTeamBaseData(ctx, sysTeamModel, team)
	})
	if !types.IsNil(err) {
		return nil, err
	}
	runtimecache.GetRuntimeCache().AppendUserTeamList(ctx, team.LeaderID, []*model.SysTeam{sysTeamModel})
	return sysTeamModel, nil
}

func (l *teamRepositoryImpl) syncTeamBaseData(ctx context.Context, sysTeamModel *model.SysTeam, team *bo.CreateTeamParams) (err error) {
	teamID := sysTeamModel.ID
	// 创建团队数据库
	_, err = l.data.GetBizDB(ctx).Exec("CREATE DATABASE IF NOT EXISTS " + "`" + data.GenBizDatabaseName(teamID) + "`")
	if !types.IsNil(err) {
		return err
	}

	sysApis, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysAPI.Find()
	if !types.IsNil(err) {
		return err
	}
	sysMenus, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysMenu.Find()
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

	teamMenus := types.SliceToWithFilter(sysMenus, func(menuItem *model.SysMenu) (*bizmodel.SysTeamMenu, bool) {
		return &bizmodel.SysTeamMenu{
			Name:     menuItem.Name,
			Path:     menuItem.Path,
			Status:   menuItem.Status,
			Icon:     menuItem.Icon,
			ParentID: menuItem.ParentID,
			Level:    menuItem.Level,
		}, true
	})

	flag := true
	// 添加管理员成员
	adminMembers := types.SliceToWithFilter(team.Admins, func(memberId uint32) (*bizmodel.SysTeamMember, bool) {
		if memberId <= 0 {
			return nil, false
		}
		if memberId == sysTeamModel.LeaderID {
			flag = false
		}
		return &bizmodel.SysTeamMember{
			UserID: memberId,
			TeamID: teamID,
			Status: vobj.StatusEnable,
			Role:   vobj.RoleAdmin,
		}, true
	})
	if flag {
		adminMembers = append(adminMembers, &bizmodel.SysTeamMember{
			UserID: sysTeamModel.LeaderID,
			TeamID: teamID,
			Status: vobj.StatusEnable,
			Role:   vobj.RoleAdmin,
		})
	}
	bizDB, bizDbErr := l.data.GetBizGormDB(teamID)
	if !types.IsNil(bizDbErr) {
		return err
	}

	modelList := bizmodel.Models()
	if len(modelList) == 0 {
		return nil
	}

	// 初始化数据表
	if err = bizDB.AutoMigrate(modelList...); !types.IsNil(err) {
		return err
	}

	return bizDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := bizquery.Use(tx)
		if len(adminMembers) > 0 {
			if err = q.SysTeamMember.Create(adminMembers...); !types.IsNil(err) {
				return err
			}
		}

		if len(teamApis) > 0 {
			// 迁移api数据到团队数据库
			if err = q.SysTeamAPI.Create(teamApis...); !types.IsNil(err) {
				return err
			}
		}

		if len(teamMenus) > 0 {
			if err = q.SysTeamMenu.Create(teamMenus...); !types.IsNil(err) {
				return err
			}
		}

		return nil
	})
}

func (l *teamRepositoryImpl) UpdateTeam(ctx context.Context, team *bo.UpdateTeamParams) error {
	// 判断团队名称是否重复
	teamInfo, err := query.Use(l.data.GetMainDB(ctx)).SysTeam.WithContext(ctx).Where(query.SysTeam.Name.Eq(team.Name)).First()
	if !types.IsNil(err) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil && teamInfo.ID != team.ID {
		return merr.ErrorI18nTeamNameExistErr(ctx)
	}
	_, err = query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam.
		Where(query.SysTeam.ID.Eq(team.ID)).
		UpdateColumnSimple(
			query.SysTeam.Name.Value(team.Name),
			query.SysTeam.Remark.Value(team.Remark),
			query.SysTeam.Logo.Value(team.Logo),
			query.SysTeam.Status.Value(team.Status.GetValue()),
		)
	return err
}

func (l *teamRepositoryImpl) GetTeamDetail(ctx context.Context, teamID uint32) (*model.SysTeam, error) {
	return query.Use(l.data.GetMainDB(ctx)).
		WithContext(ctx).SysTeam.
		Where(query.SysTeam.ID.Eq(teamID)).
		Preload(field.Associations).
		First()
}

func (l *teamRepositoryImpl) GetTeamList(ctx context.Context, params *bo.QueryTeamListParams) ([]*model.SysTeam, error) {
	q := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam
	if types.IsNil(params) {
		return q.Find()
	}
	if !types.TextIsNull(params.Keyword) {
		q = q.Where(query.SysTeam.Name.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		q = q.Where(query.SysTeam.Status.Eq(params.Status.GetValue()))
	}
	if params.CreatorID > 0 {
		q = q.Where(query.SysTeam.CreatorID.Eq(params.CreatorID))
	}
	if params.LeaderID > 0 {
		q = q.Where(query.SysTeam.LeaderID.Eq(params.LeaderID))
	}
	var teamIds []uint32
	queryTeamIds := false
	if params.UserID > 0 {
		queryTeamIds = true
		// 缓存用户的全部团队ID， 然后取出来使用
		teamList := runtimecache.GetRuntimeCache().GetUserTeamList(ctx, params.UserID)
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
		q = q.Where(query.SysTeam.ID.In(teamIds...))
	}

	// 团队列表不再分页
	//if !types.IsNil(params.Page) {
	//	total, err := q.Count()
	//	if !types.IsNil(err) {
	//		return nil, err
	//	}
	//	params.Page.SetTotal(int(total))
	//	page := params.Page
	//	pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
	//	if pageNum <= 1 {
	//		q = q.Limit(pageSize)
	//	} else {
	//		q = q.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	//	}
	//}
	return q.Order(query.SysTeam.ID.Desc()).Preload(field.Associations).Find()
}

func (l *teamRepositoryImpl) UpdateTeamStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysTeam.Where(query.SysTeam.ID.In(ids...)).
		UpdateColumnSimple(query.SysTeam.Status.Value(status.GetValue()))
	return err
}

func (l *teamRepositoryImpl) GetUserTeamList(ctx context.Context, userID uint32) ([]*model.SysTeam, error) {
	// 从全局缓存读取数据
	return l.cache.GetUserTeamList(ctx, userID), nil
}

func (l *teamRepositoryImpl) AddTeamMember(ctx context.Context, params *bo.AddTeamMemberParams) error {
	members := types.SliceToWithFilter(params.Members, func(memberItem *bo.AddTeamMemberItem) (*bizmodel.SysTeamMember, bool) {
		if types.IsNil(memberItem) {
			return nil, false
		}
		return &bizmodel.SysTeamMember{
			UserID: memberItem.UserID,
			TeamID: params.ID,
			Status: vobj.StatusEnable,
			Role:   memberItem.Role,
			TeamRoles: types.SliceToWithFilter(memberItem.RoleIDs, func(roleId uint32) (*bizmodel.SysTeamRole, bool) {
				if roleId <= 0 {
					return nil, false
				}
				return &bizmodel.SysTeamRole{
					AllFieldModel: model.AllFieldModel{ID: roleId},
				}, true
			}),
		}, true
	})
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if !types.IsNil(err) {
		return err
	}

	err = bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if err := tx.SysTeamMember.WithContext(ctx).Create(members...); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return err
	}
	runtimecache.GetRuntimeCache().AppendTeamAdminList(ctx, params.ID, members)
	return nil
}

func (l *teamRepositoryImpl) RemoveTeamMember(ctx context.Context, params *bo.RemoveTeamMemberParams) error {
	if len(params.MemberIds) == 0 {
		return nil
	}
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if !types.IsNil(err) {
		return err
	}
	err = bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if _, err = tx.SysTeamMember.WithContext(ctx).
			Where(tx.SysTeamMember.TeamID.Eq(params.ID), tx.SysTeamMember.UserID.In(params.MemberIds...)).
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
	runtimecache.GetRuntimeCache().RemoveTeamAdminList(ctx, params.ID, params.MemberIds)
	return nil
}

func (l *teamRepositoryImpl) SetMemberAdmin(ctx context.Context, params *bo.SetMemberAdminParams) error {
	if len(params.MemberIDs) == 0 {
		return nil
	}
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if !types.IsNil(err) {
		return err
	}
	q := bizquery.Use(bizDB)
	_, err = q.WithContext(ctx).SysTeamMember.Where(
		q.SysTeamMember.TeamID.Eq(params.ID),
		q.SysTeamMember.UserID.In(params.MemberIDs...),
	).UpdateColumnSimple(q.SysTeamMember.Role.Value(params.Role.GetValue()))
	if !types.IsNil(err) {
		return err
	}
	// 查询团队管理员列表
	members, err := q.SysTeamMember.Where(
		q.SysTeamMember.TeamID.Eq(params.ID),
		q.SysTeamMember.UserID.In(params.MemberIDs...),
	).Find()
	if !types.IsNil(err) {
		return err
	}
	runtimecache.GetRuntimeCache().AppendTeamAdminList(ctx, params.ID, members)
	return nil
}

func (l *teamRepositoryImpl) SetMemberRole(ctx context.Context, params *bo.SetMemberRoleParams) error {
	roles := types.SliceToWithFilter(params.RoleIDs, func(roleId uint32) (*bizmodel.SysTeamRole, bool) {
		if roleId == 0 {
			return nil, false
		}
		return &bizmodel.SysTeamRole{
			AllFieldModel: model.AllFieldModel{ID: roleId},
		}, true
	})
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if !types.IsNil(err) {
		return err
	}
	return bizquery.Use(bizDB).SysTeamMember.TeamRoles.
		Model(&bizmodel.SysTeamMember{AllFieldModel: model.AllFieldModel{ID: params.MemberID}}).Replace(roles...)
}

func (l *teamRepositoryImpl) ListTeamMember(ctx context.Context, params *bo.ListTeamMemberParams) ([]*bizmodel.SysTeamMember, error) {
	var wheres []gen.Condition
	var userWheres []gen.Condition
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if !types.IsNil(err) {
		return nil, err
	}
	qq := bizquery.Use(bizDB)
	if !types.TextIsNull(params.Keyword) {
		userWheres = append(userWheres, query.SysUser.Username.Like(params.Keyword))
	}
	if !params.Gender.IsUnknown() {
		userWheres = append(userWheres, query.SysUser.Gender.Eq(params.Gender.GetValue()))
	}
	if len(userWheres) > 0 {
		var userIds []uint32
		if err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysUser.Where(userWheres...).
			Pluck(query.SysUser.ID, &userIds); !types.IsNil(err) {
			return nil, err
		}
		wheres = append(wheres, qq.SysTeamMember.UserID.In(userIds...))
	}

	if !params.Role.IsAll() {
		wheres = append(wheres, qq.SysTeamMember.Role.Eq(params.Role.GetValue()))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, qq.SysTeamMember.Status.Eq(params.Status.GetValue()))
	}
	if len(params.MemberIDs) > 0 {
		wheres = append(wheres, qq.SysTeamMember.UserID.In(params.MemberIDs...))
	}

	q := bizquery.Use(bizDB).WithContext(ctx).SysTeamMember.Where(wheres...)
	if !types.IsNil(params.Page) {
		total, err := q.Count()
		if !types.IsNil(err) {
			return nil, err
		}
		params.Page.SetTotal(int(total))
		page := params.Page
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			q = q.Limit(pageSize)
		} else {
			q = q.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return q.Order(qq.SysTeamMember.Role.Desc(), qq.SysTeamMember.ID.Asc()).Find()
}

func (l *teamRepositoryImpl) TransferTeamLeader(ctx context.Context, params *bo.TransferTeamLeaderParams) error {
	bizDB, err := l.data.GetBizGormDB(params.ID)
	if !types.IsNil(err) {
		return err
	}
	err = bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		// 设置新管理员
		if _, err = tx.SysTeamMember.WithContext(ctx).Where(
			tx.SysTeamMember.TeamID.Eq(params.ID),
			tx.SysTeamMember.UserID.Eq(params.LeaderID),
		).UpdateColumnSimple(tx.SysTeamMember.Role.Value(vobj.RoleSuperAdmin.GetValue())); !types.IsNil(err) {
			return err
		}
		// 设置老管理员
		if _, err = tx.SysTeamMember.WithContext(ctx).Where(
			tx.SysTeamMember.TeamID.Eq(params.ID),
			tx.SysTeamMember.UserID.Neq(params.OldLeaderID),
		).UpdateColumnSimple(tx.SysTeamMember.Role.Value(vobj.RoleAdmin.GetValue())); !types.IsNil(err) {
			return err
		}
		// 系统团队信息
		if _, err = query.Use(l.data.GetMainDB(ctx)).SysTeam.WithContext(ctx).Where(
			query.SysTeam.ID.Eq(params.ID),
		).UpdateColumnSimple(query.SysTeam.LeaderID.Value(params.LeaderID)); !types.IsNil(err) {
			return err
		}
		return nil
	})
	if !types.IsNil(err) {
		return err
	}
	members, err := bizquery.Use(bizDB).WithContext(ctx).SysTeamMember.Where(
		bizquery.Use(bizDB).SysTeamMember.TeamID.Eq(params.ID),
		bizquery.Use(bizDB).SysTeamMember.UserID.In(params.OldLeaderID, params.LeaderID),
	).Find()
	if !types.IsNil(err) {
		return err
	}
	runtimecache.GetRuntimeCache().AppendTeamAdminList(ctx, params.ID, members)
	return nil
}

func (l *teamRepositoryImpl) GetUserTeamByID(ctx context.Context, userID, teamID uint32) (*bizmodel.SysTeamMember, error) {
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return nil, err
	}
	q := bizquery.Use(bizDB)
	return q.WithContext(ctx).SysTeamMember.Where(
		q.SysTeamMember.TeamID.Eq(teamID),
		q.SysTeamMember.UserID.Eq(userID),
	).First()
}
