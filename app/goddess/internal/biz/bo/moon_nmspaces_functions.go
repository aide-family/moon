package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesFunctionBo 租户功能业务对象
type NmspacesFunctionBo struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	FunctionRemark    string
	Creator           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ToAPIV1NmspacesFunctionItem 转换为 proto NmspacesFunctionItem
func (b *NmspacesFunctionBo) ToAPIV1NmspacesFunctionItem() *goddessv1.NmspacesFunctionItem {
	return &goddessv1.NmspacesFunctionItem{
		MenuCode:          b.MenuCode,
		FunctionCode:      b.FunctionCode,
		FunctionName:      b.FunctionName,
		FunctionSortOrder: b.FunctionSortOrder,
		FunctionStatus:    b.FunctionStatus,
		FunctionRemark:    b.FunctionRemark,
		Creator:           b.Creator,
		CreatedAt:         timex.FormatTime(&b.CreatedAt),
		UpdatedAt:         timex.FormatTime(&b.UpdatedAt),
	}
}

// CreateNmspacesFunctionBo 新增租户功能业务对象
type CreateNmspacesFunctionBo struct {
	NamespaceUID      snowflake.ID
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	FunctionRemark    string
	Creator           snowflake.ID
}

// NewCreateNmspacesFunctionBoFromSysFunction 从系统功能业务对象构造新增租户功能业务对象
func NewCreateNmspacesFunctionBoFromSysFunction(namespaceUID snowflake.ID, creator snowflake.ID, sysFunc *SysFunctionBo) *CreateNmspacesFunctionBo {
	return &CreateNmspacesFunctionBo{
		NamespaceUID:      namespaceUID,
		MenuCode:          sysFunc.MenuCode,
		FunctionCode:      sysFunc.FunctionCode,
		FunctionName:      sysFunc.FunctionName,
		FunctionSortOrder: sysFunc.FunctionSortOrder,
		FunctionStatus:    sysFunc.FunctionStatus,
		FunctionRemark:    sysFunc.FunctionRemark,
		Creator:           creator,
	}
}

// ListNmspacesFunctionBo 租户功能列表查询业务对象
type ListNmspacesFunctionBo struct {
	*PageRequestBo
	NamespaceUID   snowflake.ID
	MenuCode       string
	Keyword        string
	FunctionStatus *int32
}

// NewListNmspacesFunctionBo 从 proto 请求构造
func NewListNmspacesFunctionBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesFunctionRequest) *ListNmspacesFunctionBo {
	return &ListNmspacesFunctionBo{
		PageRequestBo:  NewPageRequestBo(req.Page, req.PageSize),
		NamespaceUID:   namespaceUID,
		MenuCode:       req.MenuCode,
		Keyword:        req.Keyword,
		FunctionStatus: req.FunctionStatus,
	}
}

// UpdateNmspacesFunctionBo 编辑租户功能业务对象
type UpdateNmspacesFunctionBo struct {
	NamespaceUID      snowflake.ID
	MenuCode          string
	FunctionCode      string
	FunctionName      *string
	FunctionSortOrder *int32
	FunctionStatus    *int32
	FunctionRemark    *string
}

// NewUpdateNmspacesFunctionBo 从 proto 请求构造（仅获取功能状态字段）
func NewUpdateNmspacesFunctionBo(namespaceUID snowflake.ID, req *goddessv1.UpdateNmspacesFunctionRequest) *UpdateNmspacesFunctionBo {
	return &UpdateNmspacesFunctionBo{
		NamespaceUID:   namespaceUID,
		MenuCode:       req.MenuCode,
		FunctionCode:   req.FunctionCode,
		FunctionStatus: req.FunctionStatus,
	}
}
