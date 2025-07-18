package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewRoleRepo(d *data.Data, logger log.Logger) repository.Role {
	return &roleRepoImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.role")),
	}
}

type roleRepoImpl struct {
	*data.Data
	helper *log.Helper
}

func (r *roleRepoImpl) Find(ctx context.Context, ids []uint32) ([]do.Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	roleQuery := getMainQuery(ctx, r).Role
	roles, err := roleQuery.WithContext(ctx).Where(roleQuery.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(roles, func(role *system.Role) do.Role { return role }), nil
}

func (r *roleRepoImpl) Get(ctx context.Context, id uint32) (do.Role, error) {
	roleQuery := getMainQuery(ctx, r).Role
	role, err := roleQuery.WithContext(ctx).Where(roleQuery.ID.Eq(id)).First()
	if err != nil {
		return nil, roleNotFound(err)
	}
	return role, nil
}

func (r *roleRepoImpl) List(ctx context.Context, req *bo.ListRoleReq) (*bo.ListRoleReply, error) {
	roleQuery := getMainQuery(ctx, r).Role
	wrapper := roleQuery.WithContext(ctx)

	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(roleQuery.Status.Eq(req.Status.GetValue()))
	}
	if !validate.TextIsNull(req.Keyword) {
		ors := []gen.Condition{
			roleQuery.Name.Like(req.Keyword),
			roleQuery.Remark.Like(req.Keyword),
		}
		wrapper = wrapper.Where(roleQuery.Or(ors...))
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
	rows := slices.Map(roles, func(role *system.Role) do.Role { return role })
	return req.ToListReply(rows), nil
}

func (r *roleRepoImpl) Create(ctx context.Context, role bo.Role) error {
	roleDo := &system.Role{
		Name:   role.GetName(),
		Remark: role.GetRemark(),
		Status: role.GetStatus(),
	}
	roleDo.WithContext(ctx)
	roleMutation := getMainQuery(ctx, r).Role
	if err := roleMutation.WithContext(ctx).Create(roleDo); err != nil {
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
	return roleMutation.Menus.Model(roleDo).Append(menus...)
}

func (r *roleRepoImpl) Update(ctx context.Context, role bo.Role) error {
	menus := slices.MapFilter(role.GetMenus(), func(menu do.Menu) (*system.Menu, bool) {
		if validate.IsNil(menu) || menu.GetID() <= 0 {
			return nil, false
		}
		return &system.Menu{
			BaseModel: do.BaseModel{ID: menu.GetID()},
		}, true
	})
	roleMutation := getMainQuery(ctx, r).Role

	mutations := []field.AssignExpr{
		roleMutation.Name.Value(role.GetName()),
		roleMutation.Remark.Value(role.GetRemark()),
		roleMutation.Status.Value(role.GetStatus().GetValue()),
	}
	wrapper := []gen.Condition{
		roleMutation.ID.Eq(role.GetID()),
	}
	_, err := roleMutation.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(mutations...)
	if err != nil {
		return err
	}
	roleDo := &system.Role{
		CreatorModel: do.CreatorModel{
			BaseModel: do.BaseModel{
				ID: role.GetID(),
			},
		},
	}
	menusAssociation := roleMutation.Menus.WithContext(ctx).Model(roleDo)
	if len(menus) == 0 {
		return menusAssociation.Clear()
	}
	return menusAssociation.Replace(menus...)
}

func (r *roleRepoImpl) Delete(ctx context.Context, id uint32) error {
	roleMutation := getMainQuery(ctx, r).Role
	wrapper := []gen.Condition{
		roleMutation.ID.Eq(id),
	}
	_, err := roleMutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

func (r *roleRepoImpl) UpdateStatus(ctx context.Context, req *bo.UpdateRoleStatusReq) error {
	roleMutation := getMainQuery(ctx, r).Role
	wrapper := []gen.Condition{
		roleMutation.ID.Eq(req.RoleID),
	}
	_, err := roleMutation.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(roleMutation.Status.Value(req.Status.GetValue()))
	return err
}

func (r *roleRepoImpl) UpdateUsers(ctx context.Context, req bo.UpdateRoleUsers) error {
	roleMutation := getMainQuery(ctx, r).Role
	roleDo := &system.Role{
		CreatorModel: do.CreatorModel{
			BaseModel: do.BaseModel{
				ID: req.GetRole().GetID(),
			},
		},
	}
	roleMutation.WithContext(ctx)
	users := slices.MapFilter(req.GetUsers(), func(user do.User) (*system.User, bool) {
		if validate.IsNil(user) || user.GetID() <= 0 {
			return nil, false
		}
		return &system.User{
			BaseModel: do.BaseModel{
				ID: user.GetID(),
			},
		}, true
	})

	usersAssociation := roleMutation.Users.WithContext(ctx).Model(roleDo)
	if len(users) == 0 {
		return usersAssociation.Clear()
	}
	return usersAssociation.Replace(users...)
}
