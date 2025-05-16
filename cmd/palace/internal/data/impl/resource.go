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
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewResourceRepo(d *data.Data, logger log.Logger) repository.Resource {
	return &resourceImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.resource")),
	}
}

type resourceImpl struct {
	*data.Data
	helper *log.Helper
}

func (r *resourceImpl) Find(ctx context.Context, ids []uint32) ([]do.Resource, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	mainQuery := getMainQuery(ctx, r)
	resource := mainQuery.Resource
	resourceDos, err := resource.WithContext(ctx).Where(resource.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(resourceDos, func(resource *system.Resource) do.Resource { return resource }), nil
}

func (r *resourceImpl) CreateResource(ctx context.Context, req bo.SaveResource) error {
	resourceDo := &system.Resource{
		Name:   req.GetName(),
		Path:   req.GetPath(),
		Status: req.GetStatus(),
		Allow:  req.GetAllow(),
		Remark: req.GetRemark(),
	}
	resourceDo.WithContext(ctx)
	resourceMutation := getMainQuery(ctx, r).Resource
	return resourceMutation.WithContext(ctx).Create(resourceDo)
}

func (r *resourceImpl) UpdateResource(ctx context.Context, req bo.SaveResource) error {
	resourceMutation := getMainQuery(ctx, r).Resource
	wrapper := []gen.Condition{
		resourceMutation.ID.Eq(req.GetID()),
	}
	mutations := []field.AssignExpr{
		resourceMutation.Name.Value(req.GetName()),
		resourceMutation.Path.Value(req.GetPath()),
		resourceMutation.Status.Value(req.GetStatus().GetValue()),
		resourceMutation.Allow.Value(req.GetAllow().GetValue()),
		resourceMutation.Remark.Value(req.GetRemark()),
	}
	_, err := resourceMutation.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(mutations...)
	return err
}

func (r *resourceImpl) CreateMenu(ctx context.Context, req bo.SaveMenu) error {
	resources := slices.MapFilter(req.GetResources(), func(resource do.Resource) (*system.Resource, bool) {
		if validate.IsNil(resource) || resource.GetID() <= 0 {
			return nil, false
		}
		return &system.Resource{
			BaseModel: do.BaseModel{ID: resource.GetID()},
		}, true
	})

	menuDo := &system.Menu{
		Name:      req.GetName(),
		Path:      req.GetPath(),
		Status:    req.GetStatus(),
		Icon:      req.GetIcon(),
		Type:      req.GetType(),
		Resources: resources,
	}
	if validate.IsNotNil(req.GetParent()) {
		menuDo.ParentID = req.GetParent().GetID()
	}
	menuDo.WithContext(ctx)
	menuMutation := getMainQuery(ctx, r).Menu
	return menuMutation.WithContext(ctx).Create(menuDo)
}

func (r *resourceImpl) UpdateMenu(ctx context.Context, req bo.SaveMenu) error {
	menuMutation := getMainQuery(ctx, r).Menu
	wrapper := []gen.Condition{
		menuMutation.ID.Eq(req.GetID()),
	}
	mutations := []field.AssignExpr{
		menuMutation.Name.Value(req.GetName()),
		menuMutation.Path.Value(req.GetPath()),
		menuMutation.Status.Value(req.GetStatus().GetValue()),
		menuMutation.Icon.Value(req.GetIcon()),
		menuMutation.Type.Value(req.GetType().GetValue()),
	}
	if validate.IsNotNil(req.GetParent()) {
		mutations = append(mutations, menuMutation.ParentID.Value(req.GetParent().GetID()))
	}
	_, err := menuMutation.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(mutations...)
	if err != nil {
		return err
	}
	resources := slices.MapFilter(req.GetResources(), func(resource do.Resource) (*system.Resource, bool) {
		if validate.IsNil(resource) || resource.GetID() <= 0 {
			return nil, false
		}
		return &system.Resource{
			BaseModel: do.BaseModel{ID: resource.GetID()},
		}, true
	})
	menuDo := &system.Menu{
		BaseModel: do.BaseModel{ID: req.GetID()},
	}
	menuMutation.WithContext(ctx)
	resourcesAssociation := menuMutation.Resources.WithContext(ctx).Model(menuDo)
	if len(resources) == 0 {
		return resourcesAssociation.Clear()
	}
	return resourcesAssociation.Replace(resources...)
}

func (r *resourceImpl) GetMenuByID(ctx context.Context, id uint32) (do.Menu, error) {
	mainQuery := getMainQuery(ctx, r)
	menu := mainQuery.Menu
	menuDo, err := menu.WithContext(ctx).Where(menu.ID.Eq(id)).First()
	if err != nil {
		return nil, menuNotFound(err)
	}
	return menuDo, nil
}

func (r *resourceImpl) GetResources(ctx context.Context) ([]do.Resource, error) {
	mainQuery := getMainQuery(ctx, r)
	resources, err := mainQuery.Resource.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(resources, func(resource *system.Resource) do.Resource { return resource }), nil
}

func (r *resourceImpl) GetResourceByID(ctx context.Context, id uint32) (do.Resource, error) {
	mainQuery := getMainQuery(ctx, r)
	resource := mainQuery.Resource
	resourceDo, err := resource.WithContext(ctx).Where(resource.ID.Eq(id)).First()
	if err != nil {
		return nil, resourceNotFound(err)
	}
	return resourceDo, nil
}

func (r *resourceImpl) GetResourceByOperation(ctx context.Context, operation string) (do.Resource, error) {
	mainQuery := getMainQuery(ctx, r)
	resource := mainQuery.Resource
	resourceDo, err := resource.WithContext(ctx).Where(resource.Path.Eq(operation)).First()
	if err != nil {
		return nil, resourceNotFound(err)
	}
	return resourceDo, nil
}

func (r *resourceImpl) BatchUpdateResourceStatus(ctx context.Context, ids []uint32, status vobj.GlobalStatus) error {
	if len(ids) == 0 {
		return nil
	}
	mainQuery := getMainQuery(ctx, r)
	resource := mainQuery.Resource
	_, err := resource.WithContext(ctx).
		Where(resource.ID.In(ids...)).
		Update(resource.Status, int8(status))
	return err
}

func (r *resourceImpl) ListResources(ctx context.Context, req *bo.ListResourceReq) (*bo.ListResourceReply, error) {
	mainQuery := getMainQuery(ctx, r)
	resource := mainQuery.Resource
	resourceQuery := resource.WithContext(ctx)
	if len(req.Statuses) > 0 {
		resourceQuery = resourceQuery.Where(resource.Status.In(slices.Map(req.Statuses, func(status vobj.GlobalStatus) int8 { return status.GetValue() })...))
	}
	if !validate.TextIsNull(req.Keyword) {
		resourceQuery = resourceQuery.Where(resource.Name.Like(req.Keyword))
	}
	if req.PaginationRequest != nil {
		total, err := resourceQuery.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		resourceQuery = resourceQuery.Offset(req.Offset()).Limit(int(req.Limit))
	}
	resources, err := resourceQuery.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(resources, func(resource *system.Resource) do.Resource { return resource })
	return req.ToListReply(rows), nil
}

func (r *resourceImpl) GetMenusByUserID(ctx context.Context, userID uint32) ([]do.Menu, error) {
	mainQuery := getMainQuery(ctx, r)
	user := mainQuery.User
	userQuery := user.WithContext(ctx).Where(user.ID.Eq(userID)).Preload(user.Roles.Menus)
	userDo, err := userQuery.First()
	if err != nil {
		return nil, userNotFound(err)
	}
	menus := make([]do.Menu, 0, len(userDo.Roles))
	for _, role := range userDo.Roles {
		menus = append(menus, slices.Map(role.Menus, func(menu *system.Menu) do.Menu { return menu })...)
	}
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return menus, nil
	}
	teamMember := mainQuery.TeamMember
	teamQuery := teamMember.WithContext(ctx).Where(teamMember.TeamID.Eq(teamId)).Preload(teamMember.Roles.Menus)
	teamMemberDo, err := teamQuery.First()
	if err != nil {
		return nil, teamMemberNotFound(err)
	}
	for _, role := range teamMemberDo.Roles {
		menus = append(menus, slices.Map(role.Menus, func(menu *system.Menu) do.Menu { return menu })...)
	}
	return menus, nil
}

func (r *resourceImpl) GetMenus(ctx context.Context, t vobj.MenuType) ([]do.Menu, error) {
	mainQuery := getMainQuery(ctx, r)
	menu := mainQuery.Menu
	menus, err := menu.WithContext(ctx).Where(menu.Type.Eq(t.GetValue())).Find()
	if err != nil {
		return nil, err
	}
	return slices.Map(menus, func(menu *system.Menu) do.Menu { return menu }), nil
}
