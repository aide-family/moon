package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/query"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
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
	apis := types.SliceToWithFilter(teamRole.Permissions, func(apiId uint32) (*model.SysAPI, bool) {
		if apiId <= 0 {
			return nil, false
		}
		return &model.SysAPI{ID: apiId}, true
	})

	err := query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		// 创建角色
		if err := tx.SysTeamRole.WithContext(ctx).Create(sysTeamRoleModel); err != nil {
			return err
		}
		// 添加api关联
		return tx.SysTeamRole.Apis.WithContext(ctx).Model(sysTeamRoleModel).Append(apis...)
	})
	return sysTeamRoleModel, err
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
	apis := types.SliceToWithFilter(teamRole.Permissions, func(apiId uint32) (*model.SysAPI, bool) {
		if apiId <= 0 {
			return nil, false
		}
		return &model.SysAPI{ID: apiId}, true
	})
	return query.Use(l.data.GetMainDB(ctx)).Transaction(func(tx *query.Query) error {
		if _, err = tx.SysTeamRole.WithContext(ctx).Where(tx.SysTeamRole.ID.Eq(sysTeamRoleModel.ID)).UpdateColumnSimple(
			tx.SysTeamRole.Name.Value(teamRole.Name),
			tx.SysTeamRole.Remark.Value(teamRole.Remark),
		); err != nil {
			return err
		}

		return tx.SysTeamRole.Apis.WithContext(ctx).Model(sysTeamRoleModel).Replace(apis...)
	})
}

func (l *teamRoleRepoImpl) DeleteTeamRole(ctx context.Context, id uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
		Where(query.SysTeamRole.ID.Eq(id)).Delete()
	return err
}

func (l *teamRoleRepoImpl) GetTeamRole(ctx context.Context, id uint32) (*model.SysTeamRole, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
		Where(query.SysTeamRole.ID.Eq(id)).First()
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
	).Model(&model.SysTeamMember{}).Find()
}

func (l *teamRoleRepoImpl) UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).SysTeamRole.WithContext(ctx).
		Where(query.SysTeamRole.ID.In(ids...)).
		UpdateColumnSimple(query.SysTeamRole.Status.Value(status))
	return err
}
