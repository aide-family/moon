package biz

import (
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
)

func NewDepartmentBiz(departmentRepo repo.DepartmentRepo) *DepartmentBiz {
	return &DepartmentBiz{
		departmentRepo: departmentRepo,
	}
}

// DepartmentBiz .
type DepartmentBiz struct {
	departmentRepo repo.DepartmentRepo
}
