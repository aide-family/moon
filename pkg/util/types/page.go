package types

import (
	"reflect"

	"gorm.io/gen"
)

var _ Pagination = (*page)(nil)

type (
	// page 分页器
	page struct {
		PageNum  int32 `json:"pageNum"`
		PageSize int32 `json:"pageSize"`
		total    int64
	}

	// IPaginationReq 分页请求参数接口
	IPaginationReq interface {
		GetPageNum() int32
		GetPageSize() int32
	}

	// Pagination 分页器接口
	Pagination interface {
		IPaginationReq
		GetTotal() int64
		SetTotal(total int64)
	}

	// Limiter 限制器接口
	Limiter[T any] interface {
		Limit(limit int) T
	}

	// PageQuery 分页查询接口
	PageQuery[T any] interface {
		Limit(limit int) T
		Offset(offset int) T
		Where(conds ...gen.Condition) T
		Count() (count int64, err error)
	}
)

// WithPageQuery 分页查询
func WithPageQuery[T any, Q PageQuery[T]](q Q, page Pagination) (T, error) {
	var res T
	res = q.Where()
	if IsNil(q) || IsNil(page) {
		return res, nil
	}
	total, err := q.Count()
	if !IsNil(err) {
		return res, err
	}

	page.SetTotal(total)
	pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
	if pageNum <= 1 {
		res = q.Limit(int(pageSize))
	} else {
		res = q.Offset(int((pageNum - 1) * pageSize))
		// 通过反射调用Limit方法
		res = Limit(res, int(pageSize)).(T)
	}
	return res, nil
}

// Limit 查询限制器
func Limit(t any, limit int) any {
	// 通过反射调用Limit方法
	call := reflect.ValueOf(t).MethodByName("Limit").Call([]reflect.Value{reflect.ValueOf(limit)})
	return call[0].Interface()
}

// NewPage 创建一个分页器
func NewPage(pageNum, pageSize int32) Pagination {
	return &page{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

// NewPagination 获取分页器
func NewPagination(page IPaginationReq) Pagination {
	if IsNil(page) {
		return nil
	}
	return NewPage(page.GetPageNum(), page.GetPageSize())
}

// GetPageNum 获取页码
func (l *page) GetPageNum() int32 {
	return l.PageNum
}

// GetPageSize 获取每页数量
func (l *page) GetPageSize() int32 {
	return l.PageSize
}

// GetTotal 获取总条数
func (l *page) GetTotal() int64 {
	return l.total
}

// SetTotal 设置总条数
func (l *page) SetTotal(total int64) {
	l.total = total
}
