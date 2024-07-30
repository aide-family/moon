package build

import (
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

// ResourceBuilder 资源构造器
type ResourceBuilder struct {
	*model.SysAPI
}

// NewResourceBuilder 创建资源构造器
func NewResourceBuilder(resource *model.SysAPI) *ResourceBuilder {
	return &ResourceBuilder{
		SysAPI: resource,
	}
}

// ToAPI 转换为API模型
func (b *ResourceBuilder) ToAPI() *admin.ResourceItem {
	if types.IsNil(b) || types.IsNil(b.SysAPI) {
		return nil
	}

	return &admin.ResourceItem{
		Id:        b.ID,
		Name:      b.Name,
		Path:      b.Path,
		Status:    api.Status(b.Status),
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		DeletedAt: types.NewTime(time.Unix(int64(b.DeletedAt), 0)).String(),
		Module:    api.ModuleType(b.Module),
		Domain:    api.DomainType(b.Domain),
	}
}

// TeamResourceBuilder 团队资源构造器
type TeamResourceBuilder struct {
	*bizmodel.SysTeamAPI
}

// NewTeamResourceBuilder 创建团队资源构造器
func NewTeamResourceBuilder(resource *bizmodel.SysTeamAPI) *TeamResourceBuilder {
	return &TeamResourceBuilder{
		SysTeamAPI: resource,
	}
}

// ToAPI 转换为API模型
func (b *TeamResourceBuilder) ToAPI() *admin.ResourceItem {
	if types.IsNil(b) || types.IsNil(b.SysTeamAPI) {
		return nil
	}

	return &admin.ResourceItem{
		Id:        b.ID,
		Name:      b.Name,
		Path:      b.Path,
		Status:    api.Status(b.Status),
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		DeletedAt: types.NewTime(time.Unix(int64(b.DeletedAt), 0)).String(),
		Module:    api.ModuleType(b.Module),
		Domain:    api.DomainType(b.Domain),
	}
}
