// Package bo is the business logic object
package bo

func NewPageRequestBo(page int32, pageSize int32) *PageRequestBo {
	return &PageRequestBo{
		Page:     page,
		PageSize: pageSize,
	}
}

type PageRequestBo struct {
	Page     int32
	PageSize int32
	total    int64
}

func (p *PageRequestBo) Offset() int {
	return int((p.Page - 1) * p.PageSize)
}

func (p *PageRequestBo) Limit() int {
	return int(p.PageSize)
}

func (p *PageRequestBo) WithTotal(total int64) *PageRequestBo {
	p.total = total
	return p
}

func NewPageResponseBo[T any](pageRequestBo *PageRequestBo, data []T) *PageResponseBo[T] {
	return &PageResponseBo[T]{
		items:         data,
		total:         pageRequestBo.total,
		PageRequestBo: pageRequestBo,
	}
}

type PageResponseBo[T any] struct {
	items []T
	total int64
	*PageRequestBo
}

func (p *PageResponseBo[T]) GetItems() []T {
	return p.items
}

func (p *PageResponseBo[T]) GetTotal() int64 {
	return p.total
}

func (p *PageResponseBo[T]) GetPage() int32 {
	return p.Page
}

func (p *PageResponseBo[T]) GetPageSize() int32 {
	return p.PageSize
}
