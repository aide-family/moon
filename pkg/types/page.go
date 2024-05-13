package types

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
