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
	Q                            = new(Query)
	PromCombo                    *promCombo
	PromNode                     *promNode
	PromNodeDir                  *promNodeDir
	PromNodeDirFile              *promNodeDirFile
	PromNodeDirFileGroup         *promNodeDirFileGroup
	PromNodeDirFileGroupStrategy *promNodeDirFileGroupStrategy
	PromRule                     *promRule
)

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	*Q = *Use(db, opts...)
	PromCombo = &Q.PromCombo
	PromNode = &Q.PromNode
	PromNodeDir = &Q.PromNodeDir
	PromNodeDirFile = &Q.PromNodeDirFile
	PromNodeDirFileGroup = &Q.PromNodeDirFileGroup
	PromNodeDirFileGroupStrategy = &Q.PromNodeDirFileGroupStrategy
	PromRule = &Q.PromRule
}

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:                           db,
		PromCombo:                    newPromCombo(db, opts...),
		PromNode:                     newPromNode(db, opts...),
		PromNodeDir:                  newPromNodeDir(db, opts...),
		PromNodeDirFile:              newPromNodeDirFile(db, opts...),
		PromNodeDirFileGroup:         newPromNodeDirFileGroup(db, opts...),
		PromNodeDirFileGroupStrategy: newPromNodeDirFileGroupStrategy(db, opts...),
		PromRule:                     newPromRule(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	PromCombo                    promCombo
	PromNode                     promNode
	PromNodeDir                  promNodeDir
	PromNodeDirFile              promNodeDirFile
	PromNodeDirFileGroup         promNodeDirFileGroup
	PromNodeDirFileGroupStrategy promNodeDirFileGroupStrategy
	PromRule                     promRule
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:                           db,
		PromCombo:                    q.PromCombo.clone(db),
		PromNode:                     q.PromNode.clone(db),
		PromNodeDir:                  q.PromNodeDir.clone(db),
		PromNodeDirFile:              q.PromNodeDirFile.clone(db),
		PromNodeDirFileGroup:         q.PromNodeDirFileGroup.clone(db),
		PromNodeDirFileGroupStrategy: q.PromNodeDirFileGroupStrategy.clone(db),
		PromRule:                     q.PromRule.clone(db),
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
		db:                           db,
		PromCombo:                    q.PromCombo.replaceDB(db),
		PromNode:                     q.PromNode.replaceDB(db),
		PromNodeDir:                  q.PromNodeDir.replaceDB(db),
		PromNodeDirFile:              q.PromNodeDirFile.replaceDB(db),
		PromNodeDirFileGroup:         q.PromNodeDirFileGroup.replaceDB(db),
		PromNodeDirFileGroupStrategy: q.PromNodeDirFileGroupStrategy.replaceDB(db),
		PromRule:                     q.PromRule.replaceDB(db),
	}
}

type queryCtx struct {
	PromCombo                    IPromComboDo
	PromNode                     IPromNodeDo
	PromNodeDir                  IPromNodeDirDo
	PromNodeDirFile              IPromNodeDirFileDo
	PromNodeDirFileGroup         IPromNodeDirFileGroupDo
	PromNodeDirFileGroupStrategy IPromNodeDirFileGroupStrategyDo
	PromRule                     IPromRuleDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		PromCombo:                    q.PromCombo.WithContext(ctx),
		PromNode:                     q.PromNode.WithContext(ctx),
		PromNodeDir:                  q.PromNodeDir.WithContext(ctx),
		PromNodeDirFile:              q.PromNodeDirFile.WithContext(ctx),
		PromNodeDirFileGroup:         q.PromNodeDirFileGroup.WithContext(ctx),
		PromNodeDirFileGroupStrategy: q.PromNodeDirFileGroupStrategy.WithContext(ctx),
		PromRule:                     q.PromRule.WithContext(ctx),
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
