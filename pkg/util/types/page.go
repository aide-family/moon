package types

var _ Pagination = (*page)(nil)

type (
	page struct {
		PageNum  int32 `json:"pageNum"`
		PageSize int32 `json:"pageSize"`
		total    int64
	}

	IPaginationReq interface {
		GetPageNum() int32
		GetPageSize() int32
	}

	// Pagination 分页器
	Pagination interface {
		IPaginationReq
		GetTotal() int64
		SetTotal(total int64)
	}

	// PageQuery 分页查询
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
	page.SetTotal(total)
	pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
	if pageNum <= 1 {
		q.Limit(int(pageSize))
	} else {
		q.Offset(int((pageNum - 1) * pageSize))
		q.Limit(int(pageSize))
	}
	return nil
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
