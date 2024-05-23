package repoimpl

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/query"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewTeamRoleRepo(data *data.Data) repo.TeamRoleRepo {
	return &teamRoleRepoImpl{
		data: data,
	}
}

type teamRoleRepoImpl struct {
	data *data.Data
}

func (l *teamRoleRepoImpl) CreateTeamRole(ctx context.Context, teamRole *bo.CreateTeamRoleParams) (*model.SysTeamRole, error) {
	sysTeamRoleModel := &model.SysTeamRole{
		TeamID: teamRole.TeamID,
		Name:   teamRole.Name,
		Status: teamRole.Status,
		Remark: teamRole.Remark,
	}
	apis, err := query.Use(l.data.GetMainDB(ctx)).SysAPI.WithContext(ctx).Where(query.SysAPI.ID.In(teamRole.Permissions...)).Find()
	if err != nil {
		return nil, err
	}
	err = query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建角色
		if err := tx.SysTeamRole.WithContext(ctx).Create(sysTeamRoleModel); err != nil {
			return err
		}
		// 添加api关联
		if err := tx.SysTeamRole.Apis.WithContext(ctx).Model(sysTeamRoleModel).Append(apis...); err != nil {
			return err
		}
		roleIdStr := strconv.FormatUint(uint64(sysTeamRoleModel.ID), 10)
		return tx.CasbinRule.WithContext(ctx).
			Create(types.SliceTo(apis, func(apiItem *model.SysAPI) *model.CasbinRule {
				return &model.CasbinRule{
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

	return sysTeamRoleModel, l.data.GetCasbin().LoadPolicy()
}

func (l *teamRoleRepoImpl) UpdateTeamRole(ctx context.Context, teamRole *bo.UpdateTeamRoleParams) error {
	// 查询角色
	sysTeamRoleModel, err := l.GetTeamRole(ctx, teamRole.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bo.TeamRoleNotFoundErr
		}
		return err
	}
	apis, err := query.Use(l.data.GetMainDB(ctx)).SysAPI.WithContext(ctx).Where(query.SysAPI.ID.In(teamRole.Permissions...)).Find()
	if err != nil {
		return err
	}
	roleIdStr := strconv.FormatUint(uint64(sysTeamRoleModel.ID), 10)
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		if _, err = tx.SysTeamRole.WithContext(ctx).Where(tx.SysTeamRole.ID.Eq(sysTeamRoleModel.ID)).UpdateColumnSimple(
			tx.SysTeamRole.Name.Value(teamRole.Name),
			tx.SysTeamRole.Remark.Value(teamRole.Remark),
		); err != nil {
			return err
		}

		if err = tx.SysTeamRole.Apis.WithContext(ctx).Model(sysTeamRoleModel).Replace(apis...); err != nil {
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
		if err := tx.CasbinRule.WithContext(ctx).
			Create(types.SliceTo(apis, func(apiItem *model.SysAPI) *model.CasbinRule {
				return &model.CasbinRule{
					Ptype: "p",
					V0:    roleIdStr,
					V1:    apiItem.Path,
					V2:    "http",
				}
			})...); err != nil {
			return err
		}

		return l.data.GetCasbin().LoadPolicy()
	})
}

func (l *teamRoleRepoImpl) DeleteTeamRole(ctx context.Context, id uint32) error {
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		if _, err := query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
			Where(query.SysTeamRole.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		if _, err := tx.CasbinRule.WithContext(ctx).
			Where(tx.CasbinRule.V0.Eq(strconv.FormatUint(uint64(id), 10))).Delete(); err != nil {
			return err
		}
		return l.data.GetCasbin().LoadPolicy()
	})
}

func (l *teamRoleRepoImpl) GetTeamRole(ctx context.Context, id uint32) (*model.SysTeamRole, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
		Where(query.SysTeamRole.ID.Eq(id)).Preload(query.SysTeamRole.Apis).First()
}

func (l *teamRoleRepoImpl) ListTeamRole(ctx context.Context, params *bo.ListTeamRoleParams) ([]*model.SysTeamRole, error) {
	q := query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
		Where(query.SysTeamRole.TeamID.Eq(params.TeamID))
	if !types.TextIsNull(params.Keyword) {
		q = q.Where(query.SysTeamRole.Name.Like(params.Keyword))
	}
	return q.Find()
}

func (l *teamRoleRepoImpl) GetTeamRoleByUserID(ctx context.Context, userID, teamID uint32) ([]*model.SysTeamRole, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysTeamMember.TeamRoles.
		WithContext(ctx).Where(
		query.SysTeamMember.UserID.Eq(userID),
		query.SysTeamMember.TeamID.Eq(teamID),
	).Model(&model.SysTeamMember{TeamID: teamID, UserID: userID}).Find()
}

func (l *teamRoleRepoImpl) UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	roleList, err := query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
		Where(query.SysTeamRole.ID.In(ids...)).
		Preload(query.SysTeamRole.Apis).
		Find()
	if err != nil {
		return err
	}
	casbinRules := make([]*model.CasbinRule, 0, len(roleList))
	for _, roleItem := range roleList {
		roleItemIdStr := strconv.FormatUint(uint64(roleItem.ID), 10)
		for _, apiItem := range roleItem.Apis {
			casbinRules = append(casbinRules, &model.CasbinRule{
				Ptype: "p",
				V0:    roleItemIdStr,
				V1:    apiItem.Path,
				V2:    "http",
			})
		}
	}
	idStrs := types.SliceTo(ids, func(id uint32) string {
		return strconv.FormatUint(uint64(id), 10)
	})
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		if _, err := query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
			Where(query.SysTeamRole.ID.In(ids...)).
			UpdateColumnSimple(query.SysTeamRole.Status.Value(status)); err != nil {
			return err
		}
		// 启用则创建权限
		if status.IsEnable() && len(casbinRules) > 0 {
			if err := tx.CasbinRule.WithContext(ctx).Create(casbinRules...); err != nil {
				return err
			}
			return nil
		}
		// 禁用则删除权限
		if _, err := tx.CasbinRule.WithContext(ctx).
			Where(tx.CasbinRule.V0.In(idStrs...)).
			Delete(); err != nil {
			return err
		}
		return l.data.GetCasbin().LoadPolicy()
	})
}
