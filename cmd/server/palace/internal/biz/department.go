package biz

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
)

func NewDepartmentBiz(departmentRepo repository.Department) *DepartmentBiz {
	return &DepartmentBiz{
		departmentRepo: departmentRepo,
	}
}

// DepartmentBiz .
type DepartmentBiz struct {
	departmentRepo repository.Department
}
