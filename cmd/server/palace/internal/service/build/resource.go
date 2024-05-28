package build

import (
	"time"

	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/admin"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
)

type ResourceBuild struct {
	*model.SysAPI
}

func NewResourceBuild(resource *model.SysAPI) *ResourceBuild {
	return &ResourceBuild{
		SysAPI: resource,
	}
}

func (b *ResourceBuild) ToApi() *admin.ResourceItem {
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
		DeletedAt: types.NewTime(time.Unix(b.DeletedAt, 0)).String(),
		Module:    api.ModuleType(b.Module),
		Domain:    api.DomainType(b.Domain),
	}
}

type TeamResourceBuild struct {
	*bizmodel.SysTeamAPI
}

func NewTeamResourceBuild(resource *bizmodel.SysTeamAPI) *TeamResourceBuild {
	return &TeamResourceBuild{
		SysTeamAPI: resource,
	}
}

func (b *TeamResourceBuild) ToApi() *admin.ResourceItem {
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
		DeletedAt: types.NewTime(time.Unix(b.DeletedAt, 0)).String(),
		Module:    api.ModuleType(b.Module),
		Domain:    api.DomainType(b.Domain),
	}
}
