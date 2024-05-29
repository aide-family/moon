package repoimpl

import (
	"context"
	"strconv"

	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel/bizquery"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewTeamRoleRepository(data *data.Data) repository.TeamRole {
	return &teamRoleRepositoryImpl{
		data: data,
	}
}

type teamRoleRepositoryImpl struct {
	data *data.Data
}

func (l *teamRoleRepositoryImpl) CreateTeamRole(ctx context.Context, teamRole *bo.CreateTeamRoleParams) (*bizmodel.SysTeamRole, error) {
	bizDB, err := l.data.GetBizGormDB(teamRole.TeamID)
	if err != nil {
		return nil, err
	}
	q := bizquery.Use(bizDB)
	apis, err := bizquery.Use(bizDB).SysTeamAPI.WithContext(ctx).Where(q.SysTeamAPI.ID.In(teamRole.Permissions...)).Find()
	if err != nil {
		return nil, err
	}
	sysTeamRoleModel := &bizmodel.SysTeamRole{
		TeamID: teamRole.TeamID,
		Name:   teamRole.Name,
		Status: teamRole.Status.GetValue(),
		Remark: teamRole.Remark,
		Apis:   apis,
	}

	err = bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		// 创建角色
		if err = tx.SysTeamRole.WithContext(ctx).Create(sysTeamRoleModel); err != nil {
			return err
		}
		roleIdStr := strconv.FormatUint(uint64(sysTeamRoleModel.ID), 10)
		return tx.CasbinRule.WithContext(ctx).
			Create(types.SliceTo(apis, func(apiItem *bizmodel.SysTeamAPI) *bizmodel.CasbinRule {
				return &bizmodel.CasbinRule{
					Ptype: "p",
					V0:    roleIdStr,
					V1:    apiItem.Path,
					V2:    "http",
				}
			})...)
	})
	if err != nil {
		return nil, err
	}

	return sysTeamRoleModel, l.data.GetCasbin(teamRole.TeamID).LoadPolicy()
}

func (l *teamRoleRepositoryImpl) UpdateTeamRole(ctx context.Context, teamRole *bo.UpdateTeamRoleParams) error {
	// 查询角色
	sysTeamRoleModel, err := l.GetTeamRole(ctx, teamRole.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.TeamRoleNotFoundErr
		}
		return err
	}
	bizDB, err := l.data.GetBizGormDB(sysTeamRoleModel.TeamID)
	if err != nil {
		return err
	}
	q := bizquery.Use(bizDB)
	apis, err := q.SysTeamAPI.WithContext(ctx).Where(q.SysTeamAPI.ID.In(teamRole.Permissions...)).Find()
	if err != nil {
		return err
	}
	roleIdStr := strconv.FormatUint(uint64(sysTeamRoleModel.ID), 10)
	return bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if err = tx.SysTeamRole.Apis.WithContext(ctx).Model(sysTeamRoleModel).Replace(apis...); err != nil {
			return err
		}

		if _, err = tx.SysTeamRole.WithContext(ctx).Where(tx.SysTeamRole.ID.Eq(sysTeamRoleModel.ID)).UpdateColumnSimple(
			tx.SysTeamRole.Name.Value(teamRole.Name),
			tx.SysTeamRole.Remark.Value(teamRole.Remark),
		); err != nil {
			return err
		}

		// 删除角色权限
		_, err = tx.CasbinRule.WithContext(ctx).Where(tx.CasbinRule.V0.Eq(roleIdStr)).Delete()
		if err != nil {
			return err
		}
		if len(apis) == 0 {
			return nil
		}
		if err = tx.CasbinRule.WithContext(ctx).
			Create(types.SliceTo(apis, func(apiItem *bizmodel.SysTeamAPI) *bizmodel.CasbinRule {
				return &bizmodel.CasbinRule{
					Ptype: "p",
					V0:    roleIdStr,
					V1:    apiItem.Path,
					V2:    "http",
				}
			})...); err != nil {
			return err
		}

		return l.data.GetCasbin(sysTeamRoleModel.TeamID).LoadPolicy()
	})
}

func (l *teamRoleRepositoryImpl) DeleteTeamRole(ctx context.Context, id uint32) error {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	bizDB, err := l.data.GetBizGormDB(claims.GetTeam())
	if err != nil {
		return err
	}
	return bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if _, err = tx.SysTeamRole.WithContext(ctx).
			Where(tx.SysTeamRole.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		if _, err = tx.CasbinRule.WithContext(ctx).
			Where(tx.CasbinRule.V0.Eq(strconv.FormatUint(uint64(id), 10))).Delete(); err != nil {
			return err
		}
		return l.data.GetCasbin(claims.GetTeam()).LoadPolicy()
	})
}

func (l *teamRoleRepositoryImpl) GetTeamRole(ctx context.Context, id uint32) (*bizmodel.SysTeamRole, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, bo.UnLoginErr
	}
	bizDB, err := l.data.GetBizGormDB(claims.GetTeam())
	if err != nil {
		return nil, err
	}
	q := bizquery.Use(bizDB)
	return q.SysTeamRole.WithContext(ctx).
		Where(q.SysTeamRole.ID.Eq(id)).Preload(q.SysTeamRole.Apis).First()
}

func (l *teamRoleRepositoryImpl) ListTeamRole(ctx context.Context, params *bo.ListTeamRoleParams) ([]*bizmodel.SysTeamRole, error) {
	bizDB, err := l.data.GetBizGormDB(params.TeamID)
	if err != nil {
		return nil, err
	}
	qq := bizquery.Use(bizDB)
	q := qq.SysTeamRole.WithContext(ctx).
		Where(qq.SysTeamRole.TeamID.Eq(params.TeamID))
	if !types.TextIsNull(params.Keyword) {
		q = q.Where(qq.SysTeamRole.Name.Like(params.Keyword))
	}
	return q.Find()
}

func (l *teamRoleRepositoryImpl) GetTeamRoleByUserID(ctx context.Context, userID, teamID uint32) ([]*bizmodel.SysTeamRole, error) {
	bizDB, err := l.data.GetBizGormDB(teamID)
	if err != nil {
		return nil, err
	}
	q := bizquery.Use(bizDB)
	// 查询member信息
	memberDetail, err := q.SysTeamMember.WithContext(ctx).Where(q.SysTeamMember.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	return q.SysTeamMember.TeamRoles.WithContext(ctx).Model(memberDetail).Find()
}

func (l *teamRoleRepositoryImpl) UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return bo.UnLoginErr
	}
	bizDB, err := l.data.GetBizGormDB(claims.GetTeam())
	if err != nil {
		return err
	}
	q := bizquery.Use(bizDB)
	roleList, err := q.SysTeamRole.WithContext(ctx).
		Where(q.SysTeamRole.ID.In(ids...)).
		Preload(q.SysTeamRole.Apis).
		Find()
	if err != nil {
		return err
	}
	casbinRules := make([]*bizmodel.CasbinRule, 0, len(roleList))
	for _, roleItem := range roleList {
		roleItemIdStr := strconv.FormatUint(uint64(roleItem.ID), 10)
		for _, apiItem := range roleItem.Apis {
			casbinRules = append(casbinRules, &bizmodel.CasbinRule{
				Ptype: "p",
				V0:    roleItemIdStr,
				V1:    apiItem.Path,
				V2:    "http",
			})
		}
	}
	idStrList := types.SliceTo(ids, func(id uint32) string {
		return strconv.FormatUint(uint64(id), 10)
	})
	return bizquery.Use(bizDB).Transaction(func(tx *bizquery.Query) error {
		if _, err = tx.SysTeamRole.WithContext(ctx).
			Where(tx.SysTeamRole.ID.In(ids...)).
			UpdateColumnSimple(tx.SysTeamRole.Status.Value(status.GetValue())); err != nil {
			return err
		}
		// 启用则创建权限
		if status.IsEnable() && len(casbinRules) > 0 {
			if err = tx.CasbinRule.WithContext(ctx).Create(casbinRules...); err != nil {
				return err
			}
		} else {
			// 禁用则删除权限
			if _, err = tx.CasbinRule.WithContext(ctx).
				Where(tx.CasbinRule.V0.In(idStrList...)).
				Delete(); err != nil {
				return err
			}
		}

		return l.data.GetCasbin(claims.GetTeam()).LoadPolicy()
	})
}

func (l *teamRoleRepositoryImpl) CheckRbac(_ context.Context, teamId uint32, roleIds []uint32, path string) (bool, error) {
	enforce := l.data.GetCasbin(teamId)
	for _, roleId := range roleIds {
		roleStr := strconv.FormatUint(uint64(roleId), 10)
		has, err := enforce.Enforce(roleStr, path, path)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}
