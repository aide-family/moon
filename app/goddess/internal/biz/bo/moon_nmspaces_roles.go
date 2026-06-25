package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesRoleItemBo 租户角色业务对象
type NmspacesRoleItemBo struct {
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	RoleSortOrder int32
	RoleRemark    string
	Creator       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ToAPIV1NmspacesRoleItem 将业务对象转换为 proto 响应对象
func (b *NmspacesRoleItemBo) ToAPIV1NmspacesRoleItem() *goddessv1.NmspacesRoleItem {
	return &goddessv1.NmspacesRoleItem{
		RoleCode:      b.RoleCode,
		RoleName:      b.RoleName,
		RoleDataScope: b.RoleDataScope,
		RoleStatus:    b.RoleStatus,
		RoleSortOrder: b.RoleSortOrder,
		RoleRemark:    b.RoleRemark,
		Creator:       b.Creator,
		CreatedAt:     timex.FormatTime(&b.CreatedAt),
		UpdatedAt:     timex.FormatTime(&b.UpdatedAt),
	}
}

// ListNmspacesRoleBo 租户角色列表查询业务对象
type ListNmspacesRoleBo struct {
	*PageRequestBo
	NamespaceUID snowflake.ID
	Keyword      string
	RoleCode     string
	RoleStatus   *int32
	SortBy       string
	SortOrder    string
}

// NewListNmspacesRoleBo 从 proto 请求构造查询业务对象
func NewListNmspacesRoleBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesRoleRequest) *ListNmspacesRoleBo {
	return &ListNmspacesRoleBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		NamespaceUID:  namespaceUID,
		Keyword:       req.Keyword,
		RoleCode:      req.RoleCode,
		RoleStatus:    req.RoleStatus,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
	}
}

// ToAPIV1ListNmspacesRoleReply 将分页结果转换为 proto 响应
func ToAPIV1ListNmspacesRoleReply(pageResponseBo *PageResponseBo[*NmspacesRoleItemBo]) *goddessv1.ListNmspacesRoleReply {
	items := make([]*goddessv1.NmspacesRoleItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1NmspacesRoleItem())
	}
	return &goddessv1.ListNmspacesRoleReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

// CreateNmspacesRoleBo 新增租户角色业务对象
type CreateNmspacesRoleBo struct {
	NamespaceUID  snowflake.ID
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	RoleSortOrder int32
	RoleRemark    string
	Creator       snowflake.ID
}

// NewCreateNmspacesRoleBo 从 proto 请求构造新增角色业务对象
func NewCreateNmspacesRoleBo(namespaceUID snowflake.ID, creator snowflake.ID, req *goddessv1.CreateNmspacesRoleRequest) *CreateNmspacesRoleBo {
	return &CreateNmspacesRoleBo{
		NamespaceUID:  namespaceUID,
		RoleCode:      req.RoleCode,
		RoleName:      req.RoleName,
		RoleDataScope: req.RoleDataScope,
		RoleStatus:    req.RoleStatus,
		RoleSortOrder: req.RoleSortOrder,
		RoleRemark:    req.RoleRemark,
		Creator:       creator,
	}
}

// UpdateNmspacesRoleBo 编辑租户角色业务对象
type UpdateNmspacesRoleBo struct {
	NamespaceUID  snowflake.ID
	RoleCode      string
	RoleName      *string
	RoleDataScope *int32
	RoleStatus    *int32
	RoleSortOrder *int32
	RoleRemark    *string
}

// NewUpdateNmspacesRoleBo 从 proto 请求构造编辑角色业务对象
func NewUpdateNmspacesRoleBo(namespaceUID snowflake.ID, req *goddessv1.UpdateNmspacesRoleRequest) *UpdateNmspacesRoleBo {
	return &UpdateNmspacesRoleBo{
		NamespaceUID:  namespaceUID,
		RoleCode:      req.RoleCode,
		RoleName:      req.RoleName,
		RoleDataScope: req.RoleDataScope,
		RoleStatus:    req.RoleStatus,
		RoleSortOrder: req.RoleSortOrder,
		RoleRemark:    req.RoleRemark,
	}
}
