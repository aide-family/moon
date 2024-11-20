package repoimpl

import (
	"context"
	"strconv"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// NewTeamRoleRepository 创建团队角色仓库
func NewTeamRoleRepository(data *data.Data) repository.TeamRole {
	return &teamRoleRepositoryImpl{
		data: data,
	}
}

type teamRoleRepositoryImpl struct {
	data *data.Data
}

func (l *teamRoleRepositoryImpl) CreateTeamRole(ctx context.Context, teamRole *bo.CreateTeamRoleParams) (*bizmodel.SysTeamRole, error) {
	teamID := middleware.GetTeamID(ctx)
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	apis, err := bizQuery.SysTeamAPI.WithContext(ctx).Where(bizQuery.SysTeamAPI.ID.In(teamRole.Permissions...)).Find()
	if !types.IsNil(err) {
		return nil, err
	}
	sysTeamRoleModel := &bizmodel.SysTeamRole{
		Name:   teamRole.Name,
		Status: teamRole.Status,
		Remark: teamRole.Remark,
		Apis:   apis,
	}
	sysTeamRoleModel.WithContext(ctx)

	err = bizDB.Transaction(func(tx *gorm.DB) error {
		// 创建角色
		if err = bizquery.Use(tx).SysTeamRole.WithContext(ctx).Create(sysTeamRoleModel); !types.IsNil(err) {
			return err
		}
		roleIDStr := strconv.FormatUint(uint64(sysTeamRoleModel.ID), 10)
		if len(apis) == 0 {
			return nil
		}
		_, err = l.data.GetCasbinByTx(tx).AddPolicies(types.SliceTo(apis, func(apiItem *bizmodel.SysTeamAPI) []string {
			return []string{roleIDStr, apiItem.Path, "http"}
		}))
		return err
	})

	if !types.IsNil(err) {
		return nil, err
	}

	return sysTeamRoleModel, l.data.GetCasBin(teamID).LoadPolicy()
}

func (l *teamRoleRepositoryImpl) UpdateTeamRole(ctx context.Context, teamRole *bo.UpdateTeamRoleParams) error {
	teamID := middleware.GetTeamID(ctx)
	// 查询角色
	sysTeamRoleModel, err := l.GetTeamRole(ctx, teamRole.ID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastRoleNotFound(ctx)
		}
		return err
	}
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)
	apis, err := bizQuery.SysTeamAPI.WithContext(ctx).Where(bizQuery.SysTeamAPI.ID.In(teamRole.Permissions...)).Find()
	if !types.IsNil(err) {
		return err
	}
	roleIDStr := strconv.FormatUint(uint64(sysTeamRoleModel.ID), 10)
	defer l.data.GetCasBin(teamID).LoadPolicy()
	return bizDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		casbinEnforce := l.data.GetCasbinByTx(tx)
		// 删除角色权限
		if _, err = casbinEnforce.DeletePermission(roleIDStr); !types.IsNil(err) {
			return err
		}
		queryTx := bizquery.Use(tx)
		if len(apis) > 0 {
			_, err = casbinEnforce.AddPolicies(types.SliceTo(apis, func(apiItem *bizmodel.SysTeamAPI) []string {
				return []string{roleIDStr, apiItem.Path, "http"}
			}))

			if err = queryTx.SysTeamRole.Apis.WithContext(ctx).Model(sysTeamRoleModel).Replace(apis...); !types.IsNil(err) {
				return err
			}
		} else {
			if err = queryTx.SysTeamRole.Apis.WithContext(ctx).Model(sysTeamRoleModel).Clear(); !types.IsNil(err) {
				return err
			}
		}

		if _, err = queryTx.SysTeamRole.WithContext(ctx).Where(queryTx.SysTeamRole.ID.Eq(sysTeamRoleModel.ID)).UpdateColumnSimple(
			queryTx.SysTeamRole.Name.Value(teamRole.Name),
			queryTx.SysTeamRole.Remark.Value(teamRole.Remark),
		); !types.IsNil(err) {
			return err
		}
		return err
	})
}

func (l *teamRoleRepositoryImpl) DeleteTeamRole(ctx context.Context, id uint32) error {
	teamID := middleware.GetTeamID(ctx)
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return err
	}
	defer l.data.GetCasBin(teamID).LoadPolicy()
	return bizDB.Transaction(func(tx *gorm.DB) error {
		_, err = l.data.GetCasbinByTx(tx).DeletePermission(strconv.Itoa(int(id)))
		if !types.IsNil(err) {
			return err
		}
		queryTx := bizquery.Use(tx)
		if _, err = queryTx.SysTeamRole.WithContext(ctx).
			Where(queryTx.SysTeamRole.ID.Eq(id)).Delete(); !types.IsNil(err) {
			return err
		}
		return err
	})
}

func (l *teamRoleRepositoryImpl) GetTeamRole(ctx context.Context, id uint32) (*bizmodel.SysTeamRole, error) {
	teamID := middleware.GetTeamID(ctx)
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	return bizQuery.SysTeamRole.WithContext(ctx).
		Where(bizQuery.SysTeamRole.ID.Eq(id)).
		Preload(bizQuery.SysTeamRole.Apis, bizQuery.SysTeamRole.Members).First()
}

func (l *teamRoleRepositoryImpl) ListTeamRole(ctx context.Context, params *bo.ListTeamRoleParams) ([]*bizmodel.SysTeamRole, error) {
	bizDB, err := l.data.GetBizGormDB(params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	teamRoleQuery := bizQuery.SysTeamRole.WithContext(ctx)
	if !types.TextIsNull(params.Keyword) {
		teamRoleQuery = teamRoleQuery.Where(bizQuery.SysTeamRole.Name.Like(params.Keyword))
	}
	if teamRoleQuery, err = types.WithPageQuery(teamRoleQuery, params.Page); err != nil {
		return nil, err
	}
	return teamRoleQuery.Order(bizQuery.SysTeamRole.ID.Desc()).Find()
}

func (l *teamRoleRepositoryImpl) GetTeamRoleByUserID(ctx context.Context, userID, teamID uint32) ([]*bizmodel.SysTeamRole, error) {
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	// 查询member信息
	memberDetail, err := bizQuery.SysTeamMember.WithContext(ctx).Where(bizQuery.SysTeamMember.UserID.Eq(userID)).First()
	if !types.IsNil(err) {
		return nil, err
	}

	return bizQuery.SysTeamMember.TeamRoles.WithContext(ctx).Model(memberDetail).Find()
}

func (l *teamRoleRepositoryImpl) UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}
	teamID := middleware.GetTeamID(ctx)
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return err
	}
	bizQuery := bizquery.Use(bizDB)
	roleList, err := bizQuery.SysTeamRole.WithContext(ctx).
		Where(bizQuery.SysTeamRole.ID.In(ids...)).
		Preload(bizQuery.SysTeamRole.Apis).
		Find()
	if !types.IsNil(err) {
		return err
	}
	casbinRules := make([][]string, 0, len(roleList))
	for _, roleItem := range roleList {
		roleItemIDStr := strconv.FormatUint(uint64(roleItem.ID), 10)
		for _, apiItem := range roleItem.Apis {
			casbinRules = append(casbinRules, []string{roleItemIDStr, apiItem.Path, "http"})
		}
	}
	idStrList := types.SliceTo(ids, func(id uint32) string {
		return strconv.FormatUint(uint64(id), 10)
	})
	defer l.data.GetCasBin(teamID).LoadPolicy()
	return bizDB.Transaction(func(tx *gorm.DB) error {
		queryTx := bizquery.Use(tx)
		if _, err = queryTx.SysTeamRole.WithContext(ctx).
			Where(queryTx.SysTeamRole.ID.In(ids...)).
			UpdateColumnSimple(queryTx.SysTeamRole.Status.Value(status.GetValue())); !types.IsNil(err) {
			return err
		}
		// 启用则创建权限
		if status.IsEnable() && len(casbinRules) > 0 {
			_, err = l.data.GetCasbinByTx(tx).AddPolicies(casbinRules)
			if !types.IsNil(err) {
				return err
			}
		} else {
			// 禁用则删除权限
			for _, roleStr := range idStrList {
				_, err = l.data.GetCasbinByTx(tx).DeletePermission(roleStr)
				if !types.IsNil(err) {
					return err
				}
			}
		}
		return nil
	})
}

func (l *teamRoleRepositoryImpl) CheckRbac(_ context.Context, teamID uint32, roleIDs []uint32, path string) (bool, error) {
	enforce := l.data.GetCasBin(teamID)
	_ = enforce.LoadPolicy()
	enforceParams := make([][]any, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		roleStr := strconv.FormatUint(uint64(roleID), 10)
		enforceParams = append(enforceParams, []any{roleStr, path, "http"})
	}
	has, err := enforce.BatchEnforce(enforceParams)
	if !types.IsNil(err) {
		log.Errorw("check rbac error", "", "path", path, "err", err)
		return false, err
	}
	for _, ok := range has {
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (l *teamRoleRepositoryImpl) GetBizTeamRolesByIds(ctx context.Context, teamID uint32, roleIds []uint32) ([]*bizmodel.SysTeamRole, error) {
	bizQuery, err := getTeamIDBizQuery(l.data, teamID)
	if err != nil {
		return nil, err
	}
	return bizQuery.SysTeamRole.WithContext(ctx).Where(bizQuery.SysTeamRole.ID.In(roleIds...)).Find()
}
