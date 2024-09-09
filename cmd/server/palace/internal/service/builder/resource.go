package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	resourceapi "github.com/aide-family/moon/api/admin/resource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

var _ IResourceModuleBuilder = (*resourceModuleBuilder)(nil)

type (
	resourceModuleBuilder struct {
		ctx context.Context
	}

	IResourceModuleBuilder interface {
		WithListResourceRequest(*resourceapi.ListResourceRequest) IListResourceRequestBuilder
		DoResourceBuilder() IDoResourceBuilder
	}

	IListResourceRequestBuilder interface {
		ToBo() *bo.QueryResourceListParams
	}

	listResourceRequestBuilder struct {
		ctx context.Context
		*resourceapi.ListResourceRequest
	}

	IDoResourceBuilder interface {
		ToAPI(*model.SysAPI) *adminapi.ResourceItem
		ToAPIs([]*model.SysAPI) []*adminapi.ResourceItem
		ToSelect(*model.SysAPI) *adminapi.SelectItem
		ToSelects([]*model.SysAPI) []*adminapi.SelectItem
	}

	doResourceBuilder struct {
		ctx context.Context
	}
)

func (d *doResourceBuilder) ToAPI(sysAPI *model.SysAPI) *adminapi.ResourceItem {
	if types.IsNil(d) || types.IsNil(sysAPI) {
		return nil
	}

	return &adminapi.ResourceItem{
		Id:        sysAPI.ID,
		Name:      sysAPI.Name,
		Path:      sysAPI.Path,
		Status:    api.Status(sysAPI.Status),
		Remark:    sysAPI.Remark,
		CreatedAt: sysAPI.CreatedAt.String(),
		UpdatedAt: sysAPI.UpdatedAt.String(),
		Module:    api.ModuleType(sysAPI.Module),
		Domain:    api.DomainType(sysAPI.Domain),
	}
}

func (d *doResourceBuilder) ToAPIs(apis []*model.SysAPI) []*adminapi.ResourceItem {
	if types.IsNil(d) || types.IsNil(apis) {
		return nil
	}

	return types.SliceTo(apis, func(api *model.SysAPI) *adminapi.ResourceItem {
		return d.ToAPI(api)
	})
}

func (d *doResourceBuilder) ToSelect(sysAPI *model.SysAPI) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(sysAPI) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    sysAPI.ID,
		Label:    sysAPI.Name,
		Children: nil,
		Disabled: sysAPI.DeletedAt > 0 || !sysAPI.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: sysAPI.Remark,
		},
	}
}

func (d *doResourceBuilder) ToSelects(apis []*model.SysAPI) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(apis) {
		return nil
	}

	return types.SliceTo(apis, func(api *model.SysAPI) *adminapi.SelectItem {
		return d.ToSelect(api)
	})
}

func (l *listResourceRequestBuilder) ToBo() *bo.QueryResourceListParams {
	if types.IsNil(l) || types.IsNil(l.ListResourceRequest) {
		return nil
	}

	return &bo.QueryResourceListParams{
		Keyword: l.GetKeyword(),
		Page:    types.NewPagination(l.GetPagination()),
	}
}

func (r *resourceModuleBuilder) WithListResourceRequest(request *resourceapi.ListResourceRequest) IListResourceRequestBuilder {
	return &listResourceRequestBuilder{ctx: r.ctx, ListResourceRequest: request}
}

func (r *resourceModuleBuilder) DoResourceBuilder() IDoResourceBuilder {
	return &doResourceBuilder{ctx: r.ctx}
}
