package types

import (
	"github.com/aide-family/moon/api"
)

var _ Pagination = (*page)(nil)

type (
	page struct {
		PageNum  int `json:"pageNum"`
		PageSize int `json:"pageSize"`
		total    int
	}

	Pagination interface {
		GetPageNum() int
		GetPageSize() int
		GetTotal() int
		SetTotal(total int)
	}
)

// NewPage 创建一个分页器
func NewPage(pageNum, pageSize int) Pagination {
	return &page{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

// NewPagination 获取分页器
func NewPagination(page *api.PaginationReq) Pagination {
	return NewPage(int(page.GetPageNum()), int(page.GetPageSize()))
}

func (l *page) GetPageNum() int {
	return l.PageNum
}

func (l *page) GetPageSize() int {
	return l.PageSize
}

func (l *page) GetTotal() int {
	return l.total
}

func (l *page) SetTotal(total int) {
	l.total = total
}
