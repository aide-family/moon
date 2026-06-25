package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// SysFunctionBo 系统功能业务对象
type SysFunctionBo struct {
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

// ToAPIV1SysFunctionItem 转换为 proto SysFunctionItem
func (b *SysFunctionBo) ToAPIV1SysFunctionItem() *goddessv1.SysFunctionItem {
	return &goddessv1.SysFunctionItem{
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

// CreateSysFunctionBo 新增系统功能业务对象
type CreateSysFunctionBo struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	FunctionRemark    string
	Creator           int64
}

// NewCreateSysFunctionBo 从 proto 请求构造新增系统功能业务对象
func NewCreateSysFunctionBo(req *goddessv1.CreateSysFunctionRequest) *CreateSysFunctionBo {
	return &CreateSysFunctionBo{
		MenuCode:          req.MenuCode,
		FunctionCode:      req.FunctionCode,
		FunctionName:      req.FunctionName,
		FunctionSortOrder: req.FunctionSortOrder,
		FunctionStatus:    req.FunctionStatus,
		FunctionRemark:    req.FunctionRemark,
	}
}

// ListSysFunctionBo 功能列表查询业务对象
type ListSysFunctionBo struct {
	MenuCode       string
	PageRequestBo  *PageRequestBo
	Keyword        string
	FunctionStatus *int32
}

// NewListSysFunctionBo 从 proto 请求构造查询业务对象
func NewListSysFunctionBo(req *goddessv1.ListSysFunctionRequest) *ListSysFunctionBo {
	return &ListSysFunctionBo{
		MenuCode:       req.MenuCode,
		PageRequestBo:  NewPageRequestBo(req.Page, req.PageSize),
		Keyword:        req.Keyword,
		FunctionStatus: req.FunctionStatus,
	}
}

// UpdateSysFunctionBo 编辑功能业务对象（指针字段区分未传和传零值）
type UpdateSysFunctionBo struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      *string
	FunctionSortOrder *int32
	FunctionStatus    *int32
	FunctionRemark    *string
}

// NewUpdateSysFunctionBo 从 proto 请求构造编辑功能业务对象
func NewUpdateSysFunctionBo(req *goddessv1.UpdateSysFunctionRequest) *UpdateSysFunctionBo {
	return &UpdateSysFunctionBo{
		MenuCode:          req.MenuCode,
		FunctionCode:      req.FunctionCode,
		FunctionName:      req.FunctionName,
		FunctionSortOrder: req.FunctionSortOrder,
		FunctionStatus:    req.FunctionStatus,
		FunctionRemark:    req.FunctionRemark,
	}
}
