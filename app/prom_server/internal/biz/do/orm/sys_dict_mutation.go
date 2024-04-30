package orm

import (
	"context"

	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
)

func NewSysDictMutation(query SysDictQuery, opts ...sysDictMutationOption) SysDictMutation {
	m := &sysDictMutation{
		query: query,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type (
	SysDictMutation interface {
		Create() SysDictCreate
		Update() SysDictUpdate
		Delete(ctx context.Context) (int64, error)
	}

	sysDictMutation struct {
		query SysDictQuery
	}

	SysDictCreate interface {
		Save(context.Context, do.SysDictSet) (*do.SysDict, error)
	}

	sysDictCreate struct {
		query SysDictQuery
	}

	SysDictUpdate interface {
		Save(context.Context, do.SysDictSet) (*do.SysDict, error)
		ExecWithRowsAffected(context.Context, do.SysDictSet) (int64, error)
		Exec(context.Context, do.SysDictSet) error
	}

	sysDictUpdate struct {
		query SysDictQuery
	}

	sysDictMutationOption func(*sysDictMutation)
)

var ErrInvalidType = perrors.ErrorInvalidParams("db model invalid type")

func (l *sysDictMutation) Query() SysDictQuery {
	return l.query
}

func (l *sysDictMutation) Delete(ctx context.Context) (int64, error) {
	res := l.Query().WithContext(ctx).DB().Delete(&do.SysDict{})
	return res.RowsAffected, res.Error
}

func (l *sysDictMutation) Create() SysDictCreate {
	return &sysDictCreate{
		query: l.query,
	}
}

func (l *sysDictMutation) Update() SysDictUpdate {
	return &sysDictUpdate{
		query: l.query,
	}
}

func (l *sysDictCreate) Save(ctx context.Context, d do.SysDictSet) (*do.SysDict, error) {
	dictDo, ok := d.(*do.SysDict)
	if !ok {
		return nil, ErrInvalidType
	}
	if err := l.query.DB().WithContext(ctx).Create(dictDo).Error; err != nil {
		return nil, err
	}
	return dictDo, nil
}

func (l *sysDictUpdate) Save(ctx context.Context, d do.SysDictSet) (*do.SysDict, error) {
	dictDo, ok := d.(*do.SysDict)
	if !ok {
		return nil, ErrInvalidType
	}
	if err := l.query.DB().Model(&do.SysDict{}).WithContext(ctx).Updates(dictDo).Error; err != nil {
		return nil, err
	}
	l.query.DB().WithContext(ctx).First(dictDo)
	return dictDo, nil
}

func (l *sysDictUpdate) ExecWithRowsAffected(ctx context.Context, d do.SysDictSet) (int64, error) {
	dictDo, ok := d.(*do.SysDict)
	if !ok {
		return 0, ErrInvalidType
	}
	res := l.query.DB().WithContext(ctx).Updates(dictDo)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (l *sysDictUpdate) Exec(ctx context.Context, d do.SysDictSet) error {
	dictDo, ok := d.(*do.SysDict)
	if !ok {
		return ErrInvalidType
	}
	return l.query.DB().WithContext(ctx).Updates(dictDo).Error
}
