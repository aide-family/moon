package impl

import (
	"context"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/middler/permission"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamRole(data *data.Data) repository.TeamRole {
	return &teamRoleRepoImpl{
		Data: data,
	}
}

type teamRoleRepoImpl struct {
	*data.Data
}

func (t *teamRoleRepoImpl) Find(ctx context.Context, ids []uint32) ([]do.TeamRole, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("team id not found")
	}
	roleQuery := getMainQuery(ctx, t).TeamRole
	wrapper := []gen.Condition{
		roleQuery.TeamID.Eq(teamID),
		roleQuery.ID.In(ids...),
	}
	roles, err := roleQuery.WithContext(ctx).Where(wrapper...).Find()
	if err != nil {
		return nil, err
	}
	roleDos := slices.Map(roles, func(role *system.TeamRole) do.TeamRole { return role })
	return roleDos, nil
}

func (t *teamRoleRepoImpl) Get(ctx context.Context, id uint32) (do.TeamRole, error) {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("team id not found")
	}

	roleQuery := getMainQuery(ctx, t).TeamRole
	wrapper := []gen.Condition{
		roleQuery.TeamID.Eq(teamID),
		roleQuery.ID.Eq(id),
	}
	role, err := roleQuery.WithContext(ctx).Where(wrapper...).Preload(field.Associations).First()
	if err != nil {
		return nil, teamRoleNotFound(err)
	}
	return role, nil
}

func (t *teamRoleRepoImpl) List(ctx context.Context, req *bo.ListRoleReq) (*bo.ListTeamRoleReply, error) {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("team id not found")
	}
	roleQuery := getMainQuery(ctx, t).TeamRole
	wrapper := roleQuery.WithContext(ctx).Where(roleQuery.TeamID.Eq(teamID))

	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(roleQuery.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(roleQuery.Name.Like(req.Keyword))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	roles, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(roles, func(role *system.TeamRole) do.TeamRole { return role })
	return req.ToTeamRoleListReply(rows), nil
}

func (t *teamRoleRepoImpl) Create(ctx context.Context, role bo.Role) error {
	teamDo := &system.TeamRole{
		Name:   role.GetName(),
		Remark: role.GetRemark(),
		Status: role.GetStatus(),
	}
	teamDo.WithContext(ctx)

	bizRoleQuery := getMainQuery(ctx, t).TeamRole
	if err := bizRoleQuery.WithContext(ctx).Create(teamDo); err != nil {
		return err
	}
	menus := slices.MapFilter(role.GetMenus(), func(menu do.Menu) (*system.Menu, bool) {
		if validate.IsNil(menu) || menu.GetID() <= 0 {
			return nil, false
		}
		return &system.Menu{
			BaseModel: do.BaseModel{ID: menu.GetID()},
		}, true
	})
	if len(menus) == 0 {
		return nil
	}
	return bizRoleQuery.Menus.Model(teamDo).Append(menus...)
}

func (t *teamRoleRepoImpl) Update(ctx context.Context, role bo.Role) error {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return merr.ErrorPermissionDenied("team id not found")
	}
	bizRoleQuery := getMainQuery(ctx, t).TeamRole
	wrapper := []gen.Condition{
		bizRoleQuery.TeamID.Eq(teamID),
		bizRoleQuery.ID.Eq(role.GetID()),
	}
	mutations := []field.AssignExpr{
		bizRoleQuery.Name.Value(role.GetName()),
		bizRoleQuery.Remark.Value(role.GetRemark()),
		bizRoleQuery.Status.Value(role.GetStatus().GetValue()),
	}
	_, err := bizRoleQuery.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(mutations...)
	if err != nil {
		return err
	}

	roleDo := &system.TeamRole{
		TeamModel: do.TeamModel{
			CreatorModel: do.CreatorModel{
				BaseModel: do.BaseModel{ID: role.GetID()},
			},
		},
	}
	menuDos := slices.MapFilter(role.GetMenus(), func(menu do.Menu) (*system.Menu, bool) {
		if validate.IsNil(menu) || menu.GetID() <= 0 {
			return nil, false
		}
		return &system.Menu{
			BaseModel: do.BaseModel{ID: menu.GetID()},
		}, true
	})
	menuMutation := bizRoleQuery.Menus.WithContext(ctx).Model(roleDo)
	if len(menuDos) == 0 {
		return menuMutation.Clear()
	}
	return menuMutation.Replace(menuDos...)
}

func (t *teamRoleRepoImpl) Delete(ctx context.Context, id uint32) error {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return merr.ErrorPermissionDenied("team id not found")
	}
	bizRoleQuery := getMainQuery(ctx, t).TeamRole
	wrapper := []gen.Condition{
		bizRoleQuery.TeamID.Eq(teamID),
		bizRoleQuery.ID.Eq(id),
	}
	_, err := bizRoleQuery.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

func (t *teamRoleRepoImpl) UpdateStatus(ctx context.Context, req *bo.UpdateRoleStatusReq) error {
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return merr.ErrorPermissionDenied("team id not found")
	}
	bizRoleQuery := getMainQuery(ctx, t).TeamRole
	wrapper := []gen.Condition{
		bizRoleQuery.TeamID.Eq(teamID),
		bizRoleQuery.ID.Eq(req.RoleID),
	}
	_, err := bizRoleQuery.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(bizRoleQuery.Status.Value(req.Status.GetValue()))
	return err
}
