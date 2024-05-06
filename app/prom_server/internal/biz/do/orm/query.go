package orm

import (
	"context"
	"strings"

	"github.com/aide-family/moon/app/prom_server/internal/biz/do/field"
	"github.com/aide-family/moon/pkg"
	"gorm.io/gorm"
)

// NewQuery 创建查询对象
func NewQuery[Model any](db *gorm.DB, opts ...QueryOption[Model]) Query[Model] {
	q := &query[Model]{
		db: db,
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

type (
	ScopeMethod = func(db *gorm.DB) *gorm.DB

	Query[T any] interface {
		Select(fields ...string) Query[T]

		Where(conditions ...string) Query[T]
		WhereMap(conditions map[string]any) Query[T]
		Or(conditions ...string) Query[T]
		OrMap(conditions map[string]any) Query[T]

		Order(field string) Query[T]
		Limit(limit int) Query[T]
		Offset(offset int) Query[T]
		Preload(field string, conditions ...ScopeMethod) Query[T]
		Scope(scopes ...ScopeMethod) Query[T]
		DB() *gorm.DB
		WithContext(ctx context.Context) Query[T]

		Find(ctx context.Context) ([]*T, error)
		List(ctx context.Context, page field.Pagination) ([]*T, error)
		First(ctx context.Context) (*T, error)
		Last(ctx context.Context) (*T, error)
		Take(ctx context.Context) (*T, error)
		Count(ctx context.Context) (int64, error)

		Mutation() Mutation[T]

		Update() Update[T]
		Create() Create[T]
		Delete(ctx context.Context) (int64, error)
	}
	query[T any] struct {
		db *gorm.DB
	}

	QueryOption[T any] func(*query[T])
)

func (l *query[T]) Where(conditions ...string) Query[T] {
	if !pkg.IsNil(conditions) && len(conditions) > 0 {
		l.db = l.db.Where(strings.Join(conditions, " AND "))
	}

	return l
}

func (l *query[T]) WhereMap(condition map[string]any) Query[T] {
	if !pkg.IsNil(condition) && len(condition) > 0 {
		l.db = l.db.Where(condition)
	}

	return l
}

func (l *query[T]) Or(conditions ...string) Query[T] {
	if !pkg.IsNil(conditions) && len(conditions) > 0 {
		l.db = l.db.Or(strings.Join(conditions, " AND "))
	}
	return l
}

func (l *query[T]) OrMap(condition map[string]any) Query[T] {
	if !pkg.IsNil(condition) && len(condition) > 0 {
		l.db = l.db.Or(condition)
	}
	return l
}

func (l *query[T]) Order(field string) Query[T] {
	l.db = l.db.Order(field)
	return l
}

func (l *query[T]) Limit(limit int) Query[T] {
	l.db = l.db.Limit(limit)
	return l
}

func (l *query[T]) Offset(offset int) Query[T] {
	l.db = l.db.Offset(offset)
	return l
}

func (l *query[T]) Preload(field string, conditions ...ScopeMethod) Query[T] {
	if len(conditions) > 0 {
		l.db = l.db.Preload(field, conditions[0])
	} else {
		l.db = l.db.Preload(field)
	}
	return l
}

func (l *query[T]) Scope(scopes ...ScopeMethod) Query[T] {
	l.db = l.db.Scopes(scopes...)
	return l
}

func (l *query[T]) DB() *gorm.DB {
	return l.db
}

func (l *query[T]) WithContext(ctx context.Context) Query[T] {
	l.db = l.db.WithContext(ctx)
	return l
}

func (l *query[T]) Find(ctx context.Context) ([]*T, error) {
	var list []*T
	if err := l.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (l *query[T]) List(ctx context.Context, page field.Pagination) ([]*T, error) {
	var list []*T
	if !pkg.IsNil(page) {
		var total int64
		var m T
		if err := l.db.Model(&m).WithContext(ctx).Count(&total).Error; err != nil {
			return nil, err
		}
		page.SetTotal(total)
	}

	if err := l.db.WithContext(ctx).Scopes(field.Page(page)).Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (l *query[T]) First(ctx context.Context) (*T, error) {
	var dict T
	if err := l.db.WithContext(ctx).First(&dict).Error; err != nil {
		return nil, err
	}
	return &dict, nil
}

func (l *query[T]) Last(ctx context.Context) (*T, error) {
	var dict T
	if err := l.db.WithContext(ctx).Last(&dict).Error; err != nil {
		return nil, err
	}
	return &dict, nil
}

func (l *query[T]) Take(ctx context.Context) (*T, error) {
	var dict T
	if err := l.db.WithContext(ctx).Take(&dict).Error; err != nil {
		return nil, err
	}
	return &dict, nil
}

func (l *query[T]) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := l.db.WithContext(ctx).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (l *query[T]) Select(fields ...string) Query[T] {
	if len(fields) != 0 {
		l.db = l.db.Select(fields)
	}
	return l
}

func (l *query[T]) Mutation() Mutation[T] {
	return NewMutation[T](l)
}

func (l *query[T]) Update() Update[T] {
	return NewMutation[T](l).Update()
}

func (l *query[T]) Create() Create[T] {
	return NewMutation[T](l).Create()
}

func (l *query[T]) Delete(ctx context.Context) (int64, error) {
	return NewMutation[T](l).Delete(ctx)
}
