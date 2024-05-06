package orm

import (
	"context"
)

func NewMutation[T any](query Query[T], opts ...MutationOption[T]) Mutation[T] {
	m := &mutation[T]{
		query: query,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type (
	Mutation[T any] interface {
		Create() Create[T]
		Update() Update[T]
		Delete(ctx context.Context) (int64, error)
	}

	mutation[T any] struct {
		query Query[T]
	}

	Create[T any] interface {
		Save(context.Context, *T) (*T, error)
	}

	create[T any] struct {
		query Query[T]
	}

	Update[T any] interface {
		Save(context.Context, *T) (*T, error)
		ExecWithRowsAffected(context.Context, *T) (int64, error)
		Exec(context.Context, *T) error
	}

	update[T any] struct {
		query Query[T]
	}

	MutationOption[T any] func(*mutation[T])
)

func (l *mutation[T]) Query() Query[T] {
	return l.query
}

func (l *mutation[T]) Delete(ctx context.Context) (int64, error) {
	var m T
	res := l.Query().WithContext(ctx).DB().Delete(&m)
	return res.RowsAffected, res.Error
}

func (l *mutation[T]) Create() Create[T] {
	return &create[T]{
		query: l.query,
	}
}

func (l *mutation[T]) Update() Update[T] {
	return &update[T]{
		query: l.query,
	}
}

func (l *create[T]) Save(ctx context.Context, dictDo *T) (*T, error) {
	if err := l.query.DB().WithContext(ctx).Create(dictDo).Error; err != nil {
		return nil, err
	}
	return dictDo, nil
}

func (l *update[T]) Save(ctx context.Context, dictDo *T) (*T, error) {
	var m T
	if err := l.query.DB().Model(&m).WithContext(ctx).Updates(dictDo).Error; err != nil {
		return nil, err
	}
	l.query.DB().WithContext(ctx).First(dictDo)
	return dictDo, nil
}

func (l *update[T]) ExecWithRowsAffected(ctx context.Context, dictDo *T) (int64, error) {
	res := l.query.DB().WithContext(ctx).Updates(dictDo)
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (l *update[T]) Exec(ctx context.Context, dictDo *T) error {
	return l.query.DB().WithContext(ctx).Updates(dictDo).Error
}
