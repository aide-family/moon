package bo

import (
	"sync"

	"gorm.io/gorm"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
)

var _ Pagination = (*pageImpl)(nil)

type Pagination interface {
	GetCurr() int32
	GetRespCurr() int32
	GetSize() int32
	SetTotal(total int64)
	GetTotal() int64
}

type pageImpl struct {
	Curr  int32 `json:"curr"`
	Size  int32 `json:"size"`
	Total int64 `json:"total"`
	lock  sync.RWMutex
}

var (
	defaultCurr int32 = 1
	defaultSize int32 = 10
)

// WithDefaultCurr is used to set default curr
func WithDefaultCurr(curr int32) {
	defaultCurr = curr
}

// WithDefaultSize is used to set default size
func WithDefaultSize(size int32) {
	defaultSize = size
}

func (p *pageImpl) GetCurr() int32 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	curr := defaultCurr
	if p != nil && p.Curr > 0 {
		curr = p.Curr
	}
	return curr - 1
}

func (p *pageImpl) GetRespCurr() int32 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	total := p.GetTotal()
	size := p.GetSize()
	if total < int64(size) {
		return 1
	}
	return p.GetCurr() + 1
}

func (p *pageImpl) GetSize() int32 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	size := defaultSize
	if p != nil && p.Size > 0 {
		size = p.Size
	}
	return size
}

func (p *pageImpl) SetTotal(total int64) {
	if p != nil {
		p.lock.Lock()
		defer p.lock.Unlock()
		p.Total = total
	}
}

func (p *pageImpl) GetTotal() int64 {
	if p == nil {
		return 0
	}
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.Total
}

func NewPage(curr, size int32) Pagination {
	return &pageImpl{
		Curr: curr,
		Size: size,
	}
}

// Page 分页
func Page(pgInfo Pagination) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if pgInfo == nil {
			return db
		}
		return db.Offset(int((pgInfo.GetCurr()) * pgInfo.GetSize())).Limit(int(pgInfo.GetSize()))
	}
}
