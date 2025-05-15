package bo

import (
	"github.com/aide-family/moon/pkg/util/validate"
)

type PaginationRequest struct {
	Page  uint32
	Limit uint32
	total int64
}

type PaginationReply struct {
	Total uint32
	Page  uint32
	Limit uint32
}

// WithTotal set the total number of items
func (r *PaginationRequest) WithTotal(total int64) *PaginationRequest {
	if validate.IsNil(r) {
		return r
	}
	r.total = total
	return r
}

// Offset calculate the offset
func (r *PaginationRequest) Offset() int {
	if validate.IsNil(r) {
		return 0
	}
	if r.Page == 0 || r.Limit == 0 {
		return 0
	}
	return int((r.Page - 1) * r.Limit)
}

// ToReply convert the pagination request to a pagination reply
func (r *PaginationRequest) ToReply() *PaginationReply {
	if validate.IsNil(r) {
		return nil
	}
	return &PaginationReply{
		Total: uint32(r.total),
		Page:  r.Page,
		Limit: r.Limit,
	}
}

type ListReply[T any] struct {
	*PaginationReply
	Items []T
}
