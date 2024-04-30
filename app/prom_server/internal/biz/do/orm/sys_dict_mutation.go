package orm

import (
	"context"

	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

func NewSysDictMutation(query SysDictQuery, opts ...sysDictMutationOption) *SysDictMutation {
	m := &SysDictMutation{
		query: query,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type (
	SysDictMutation struct {
		query SysDictQuery
	}

	SysDictCreate struct {
		*do.SysDict
		query SysDictQuery
	}

	SysDictUpdate struct {
		*do.SysDict
		query SysDictQuery
	}

	sysDictMutationOption func(*SysDictMutation)
)

func (l *SysDictMutation) Query() SysDictQuery {
	return l.query
}

func (l *SysDictMutation) Delete(ctx context.Context) (int64, error) {
	res := l.Query().WithContext(ctx).DB().Delete(&do.SysDict{})
	return res.RowsAffected, res.Error
}

func (l *SysDictMutation) Create() *SysDictCreate {
	return &SysDictCreate{
		SysDict: &do.SysDict{},
		query:   l.query,
	}
}

func (l *SysDictMutation) Update() *SysDictUpdate {
	return &SysDictUpdate{
		SysDict: &do.SysDict{},
		query:   l.query,
	}
}

func (l *SysDictCreate) Save(ctx context.Context) (*do.SysDict, error) {
	if err := l.query.DB().WithContext(ctx).Create(l.SysDict).Error; err != nil {
		return nil, err
	}
	return l.SysDict, nil
}

func (l *SysDictUpdate) Save(ctx context.Context) (*do.SysDict, error) {
	if err := l.query.DB().WithContext(ctx).Updates(l.SysDict).Error; err != nil {
		return nil, err
	}
	l.query.DB().WithContext(ctx).First(l.SysDict)
	return l.SysDict, nil
}

func (l *SysDictUpdate) ExecWithRowsAffected(ctx context.Context) (int64, error) {
	res := l.query.DB().WithContext(ctx).Updates(l.SysDict)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (l *SysDictUpdate) Exec(ctx context.Context) error {
	return l.query.DB().WithContext(ctx).Updates(l.SysDict).Error
}

func (l *SysDictCreate) SetPromStrategies(strategies []*do.PromStrategy) *SysDictCreate {
	if l == nil {
		return nil
	}
	l.PromStrategies = strategies
	return l
}

func (l *SysDictCreate) SetID(id uint32) *SysDictCreate {
	if l == nil {
		return nil
	}
	l.ID = id
	return l
}

func (l *SysDictCreate) SetName(name string) *SysDictCreate {
	if l == nil {
		return nil
	}
	l.Name = name
	return l
}

func (l *SysDictCreate) SetCategory(category vobj.Category) *SysDictCreate {
	if l == nil {
		return nil
	}
	l.Category = category
	return l
}

func (l *SysDictCreate) SetColor(color string) *SysDictCreate {
	if l == nil {
		return nil
	}
	l.Color = color
	return l
}

func (l *SysDictCreate) SetStatus(status vobj.Status) *SysDictCreate {
	if l == nil {
		return nil
	}
	l.Status = status
	return l
}

func (l *SysDictCreate) SetRemark(remark string) *SysDictCreate {
	if l == nil {
		return nil
	}
	l.Remark = remark
	return l
}

// update

func (l *SysDictUpdate) SetPromStrategies(strategies []*do.PromStrategy) *SysDictUpdate {
	if l == nil {
		return nil
	}
	l.PromStrategies = strategies
	return l
}

func (l *SysDictUpdate) SetID(id uint32) *SysDictUpdate {
	if l == nil {
		return nil
	}
	l.ID = id
	return l
}

func (l *SysDictUpdate) SetName(name string) *SysDictUpdate {
	if l == nil {
		return nil
	}
	l.Name = name
	return l
}

func (l *SysDictUpdate) SetCategory(category vobj.Category) *SysDictUpdate {
	if l == nil {
		return nil
	}
	l.Category = category
	return l
}

func (l *SysDictUpdate) SetColor(color string) *SysDictUpdate {
	if l == nil {
		return nil
	}
	l.Color = color
	return l
}

func (l *SysDictUpdate) SetStatus(status vobj.Status) *SysDictUpdate {
	if l == nil {
		return nil
	}
	l.Status = status
	return l
}

func (l *SysDictUpdate) SetRemark(remark string) *SysDictUpdate {
	if l == nil {
		return nil
	}
	l.Remark = remark
	return l
}
