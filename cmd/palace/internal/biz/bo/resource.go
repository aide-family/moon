package bo

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

type Resource struct {
}

type BatchUpdateResourceStatusReq struct {
	IDs    []uint32          `json:"ids"`
	Status vobj.GlobalStatus `json:"status"`
}

type ListResourceReq struct {
	Statuses []vobj.GlobalStatus `json:"statuses"`
	Keyword  string              `json:"keyword"`
	*PaginationRequest
}

func (r *ListResourceReq) ToListResourceReply(resources []*system.Resource) *ListResourceReply {
	return &ListResourceReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(resources, func(resource *system.Resource) do.Resource { return resource }),
	}
}

type ListResourceReply = ListReply[do.Resource]

type SaveResource interface {
	GetID() uint32
	GetName() string
	GetPath() string
	GetStatus() vobj.GlobalStatus
	GetAllow() vobj.ResourceAllow
	GetRemark() string
}

// SaveResourceReq represents the request for saving a resource
type SaveResourceReq struct {
	ID       uint32
	Name     string
	Path     string
	Status   vobj.GlobalStatus
	Allow    vobj.ResourceAllow
	Remark   string
	resource do.Resource
}

func (r *SaveResourceReq) GetID() uint32 {
	if r == nil {
		return 0
	}
	if r.resource == nil {
		return r.ID
	}
	return r.resource.GetID()
}

func (r *SaveResourceReq) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *SaveResourceReq) GetPath() string {
	if r == nil {
		return ""
	}
	return r.Path
}

func (r *SaveResourceReq) GetStatus() vobj.GlobalStatus {
	if r == nil {
		return vobj.GlobalStatusUnknown
	}
	return r.Status
}

func (r *SaveResourceReq) GetAllow() vobj.ResourceAllow {
	if r == nil {
		return vobj.ResourceAllowUnknown
	}
	return r.Allow
}

func (r *SaveResourceReq) GetRemark() string {
	if r == nil {
		return ""
	}
	return r.Remark
}

func (r *SaveResourceReq) WithResource(resource do.Resource) SaveResource {
	r.resource = resource
	return r
}

type SaveMenu interface {
	GetID() uint32
	GetName() string
	GetPath() string
	GetStatus() vobj.GlobalStatus
	GetIcon() string
	GetParent() do.Menu
	GetType() vobj.MenuType
	GetResources() []do.Resource
}

// SaveMenuReq represents the request for saving a menu
type SaveMenuReq struct {
	resources   []do.Resource
	menu        do.Menu
	parent      do.Menu
	ID          uint32
	Name        string
	Path        string
	Status      vobj.GlobalStatus
	Icon        string
	ParentID    uint32
	Type        vobj.MenuType
	ResourceIds []uint32
}

func (r *SaveMenuReq) GetID() uint32 {
	if r == nil {
		return 0
	}
	return r.ID
}

func (r *SaveMenuReq) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *SaveMenuReq) GetPath() string {
	if r == nil {
		return ""
	}
	return r.Path
}

func (r *SaveMenuReq) GetStatus() vobj.GlobalStatus {
	if r == nil {
		return vobj.GlobalStatusUnknown
	}
	return r.Status
}

func (r *SaveMenuReq) GetIcon() string {
	if r == nil {
		return ""
	}
	return r.Icon
}

func (r *SaveMenuReq) GetParent() do.Menu {
	if r == nil {
		return nil
	}
	return r.parent
}

func (r *SaveMenuReq) GetType() vobj.MenuType {
	if r == nil {
		return vobj.MenuTypeUnknown
	}
	return r.Type
}

func (r *SaveMenuReq) GetResources() []do.Resource {
	if r == nil {
		return nil
	}
	return r.resources
}

func (r *SaveMenuReq) WithResources(resources []do.Resource) SaveMenu {
	r.resources = slices.MapFilter(resources, func(resource do.Resource) (do.Resource, bool) {
		if validate.IsNil(resource) || resource.GetID() <= 0 {
			return nil, false
		}
		return resource, true
	})
	return r
}

func (r *SaveMenuReq) WithMenu(menu do.Menu) SaveMenu {
	r.menu = menu
	return r
}

func (r *SaveMenuReq) WithParent(parent do.Menu) SaveMenu {
	r.parent = parent
	return r
}
