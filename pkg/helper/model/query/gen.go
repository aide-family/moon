// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"gorm.io/gen"

	"gorm.io/plugin/dbresolver"
)

var (
	Q       = new(Query)
	SysAPI  *sysAPI
	SysTeam *sysTeam
	SysUser *sysUser
)

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	*Q = *Use(db, opts...)
	SysAPI = &Q.SysAPI
	SysTeam = &Q.SysTeam
	SysUser = &Q.SysUser
}

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:      db,
		SysAPI:  newSysAPI(db, opts...),
		SysTeam: newSysTeam(db, opts...),
		SysUser: newSysUser(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	SysAPI  sysAPI
	SysTeam sysTeam
	SysUser sysUser
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:      db,
		SysAPI:  q.SysAPI.clone(db),
		SysTeam: q.SysTeam.clone(db),
		SysUser: q.SysUser.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:      db,
		SysAPI:  q.SysAPI.replaceDB(db),
		SysTeam: q.SysTeam.replaceDB(db),
		SysUser: q.SysUser.replaceDB(db),
	}
}

type queryCtx struct {
	SysAPI  ISysAPIDo
	SysTeam ISysTeamDo
	SysUser ISysUserDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		SysAPI:  q.SysAPI.WithContext(ctx),
		SysTeam: q.SysTeam.WithContext(ctx),
		SysUser: q.SysUser.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	tx := q.db.Begin(opts...)
	return &QueryTx{Query: q.clone(tx), Error: tx.Error}
}

type QueryTx struct {
	*Query
	Error error
}

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}