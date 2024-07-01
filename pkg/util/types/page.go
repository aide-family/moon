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

	PageQuery[T any] interface {
		Limit(limit int) T
		Offset(offset int) T
		Count() (count int64, err error)
	}
)

// WithPageQuery 分页查询
func WithPageQuery[T any](q PageQuery[T], page Pagination) error {
	if IsNil(q) || IsNil(page) {
		return nil
	}
	total, err := q.Count()
	if !IsNil(err) {
		return err
	}
	page.SetTotal(int(total))
	pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
	if pageNum <= 1 {
		q.Limit(pageSize)
	} else {
		q.Offset((pageNum - 1) * pageSize)
		q.Limit(pageSize)
	}
	return nil
}

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
