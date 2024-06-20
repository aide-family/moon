package build

import (
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
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
		DeletedAt: types.NewTime(time.Unix(int64(b.DeletedAt), 0)).String(),
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
		DeletedAt: types.NewTime(time.Unix(int64(b.DeletedAt), 0)).String(),
		Module:    api.ModuleType(b.Module),
		Domain:    api.DomainType(b.Domain),
	}
}
