package orm

import (
	"context"
	"strings"

	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/field"
	"github.com/aide-family/moon/pkg"
	"gorm.io/gorm"
)

func NewSysDictQuery(db *gorm.DB, opts ...sysDictQueryOption) SysDictQuery {
	q := &sysDictQuery{
		db: db,
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

type (
	SysDictScopeMethod = func(db *gorm.DB) *gorm.DB

	SysDictQuery interface {
		Select(fields ...string) SysDictQuery

		Where(conditions ...string) SysDictQuery
		WhereMap(conditions map[do.SysDictField]any) SysDictQuery
		Or(conditions ...string) SysDictQuery
		OrMap(conditions map[do.SysDictField]any) SysDictQuery

		Order(field string) SysDictQuery
		Limit(limit int) SysDictQuery
		Offset(offset int) SysDictQuery
		Preload(field do.SysDictWithField, conditions ...SysDictScopeMethod) SysDictQuery
		Scope(scopes ...SysDictScopeMethod) SysDictQuery
		DB() *gorm.DB
		WithContext(ctx context.Context) SysDictQuery

		Find(ctx context.Context) ([]*do.SysDict, error)
		List(ctx context.Context, page field.Pagination) ([]*do.SysDict, error)
		First(ctx context.Context) (*do.SysDict, error)
		Last(ctx context.Context) (*do.SysDict, error)
		Take(ctx context.Context) (*do.SysDict, error)
		Count(ctx context.Context) (int64, error)

		Mutation() SysDictMutation

		Update() SysDictUpdate
		Create() SysDictCreate
		Delete(ctx context.Context) (int64, error)
	}
	sysDictQuery struct {
		db *gorm.DB
	}

	sysDictQueryOption func(*sysDictQuery)
)

func (l *sysDictQuery) Where(conditions ...string) SysDictQuery {
	if !pkg.IsNil(conditions) && len(conditions) > 0 {
		l.db = l.db.Where(strings.Join(conditions, " AND "))
	}

	return l
}

func (l *sysDictQuery) WhereMap(condition map[do.SysDictField]any) SysDictQuery {
	if !pkg.IsNil(condition) && len(condition) > 0 {
		l.db = l.db.Where(condition)
	}

	return l
}

func (l *sysDictQuery) Or(conditions ...string) SysDictQuery {
	if !pkg.IsNil(conditions) && len(conditions) > 0 {
		l.db = l.db.Or(strings.Join(conditions, " AND "))
	}
	return l
}

func (l *sysDictQuery) OrMap(condition map[do.SysDictField]any) SysDictQuery {
	if !pkg.IsNil(condition) && len(condition) > 0 {
		l.db = l.db.Or(condition)
	}
	return l
}

func (l *sysDictQuery) Order(field string) SysDictQuery {
	l.db = l.db.Order(field)
	return l
}

func (l *sysDictQuery) Limit(limit int) SysDictQuery {
	l.db = l.db.Limit(limit)
	return l
}

func (l *sysDictQuery) Offset(offset int) SysDictQuery {
	l.db = l.db.Offset(offset)
	return l
}

func (l *sysDictQuery) Preload(field do.SysDictWithField, conditions ...SysDictScopeMethod) SysDictQuery {
	if len(conditions) > 0 {
		l.db = l.db.Preload(string(field), conditions[0])
	} else {
		l.db = l.db.Preload(string(field))
	}
	return l
}

func (l *sysDictQuery) Scope(scopes ...SysDictScopeMethod) SysDictQuery {
	l.db = l.db.Scopes(scopes...)
	return l
}

func (l *sysDictQuery) DB() *gorm.DB {
	return l.db
}

func (l *sysDictQuery) WithContext(ctx context.Context) SysDictQuery {
	l.db = l.db.WithContext(ctx)
	return l
}

func (l *sysDictQuery) Find(ctx context.Context) ([]*do.SysDict, error) {
	var list []*do.SysDict
	if err := l.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (l *sysDictQuery) List(ctx context.Context, page field.Pagination) ([]*do.SysDict, error) {
	var list []*do.SysDict
	if !pkg.IsNil(page) {
		var total int64
		if err := l.db.Model(&do.SysDict{}).WithContext(ctx).Count(&total).Error; err != nil {
			return nil, err
		}
		page.SetTotal(total)
	}

	if err := l.db.WithContext(ctx).Scopes(field.Page(page)).Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (l *sysDictQuery) First(ctx context.Context) (*do.SysDict, error) {
	var dict do.SysDict
	if err := l.db.WithContext(ctx).First(&dict).Error; err != nil {
		return nil, err
	}
	return &dict, nil
}

func (l *sysDictQuery) Last(ctx context.Context) (*do.SysDict, error) {
	var dict do.SysDict
	if err := l.db.WithContext(ctx).Last(&dict).Error; err != nil {
		return nil, err
	}
	return &dict, nil
}

func (l *sysDictQuery) Take(ctx context.Context) (*do.SysDict, error) {
	var dict do.SysDict
	if err := l.db.WithContext(ctx).Take(&dict).Error; err != nil {
		return nil, err
	}
	return &dict, nil
}

func (l *sysDictQuery) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := l.db.WithContext(ctx).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (l *sysDictQuery) Select(fields ...string) SysDictQuery {
	if len(fields) != 0 {
		l.db = l.db.Select(fields)
	}
	return l
}

func (l *sysDictQuery) Mutation() SysDictMutation {
	return NewSysDictMutation(l)
}

func (l *sysDictQuery) Update() SysDictUpdate {
	return NewSysDictMutation(l).Update()
}

func (l *sysDictQuery) Create() SysDictCreate {
	return NewSysDictMutation(l).Create()
}

func (l *sysDictQuery) Delete(ctx context.Context) (int64, error) {
	return NewSysDictMutation(l).Delete(ctx)
}
