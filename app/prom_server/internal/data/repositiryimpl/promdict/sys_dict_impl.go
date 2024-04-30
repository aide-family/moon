package promdict

import (
	"context"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/orm"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
)

func NewSysDictRepo(data *data.Data, logger log.Logger) repository.SysDictRepo {
	return &sysDictRepoImpl{
		data: data,
		log:  log.NewHelper(logger),
	}
}

type sysDictRepoImpl struct {
	data *data.Data
	log  *log.Helper
}

func (l *sysDictRepoImpl) query(ctx context.Context) orm.SysDictQuery {
	return orm.NewSysDictQuery(l.data.DBWithContext(ctx))
}

func (l *sysDictRepoImpl) CreateDict(ctx context.Context, dict *bo.CreateSysDictBo) (*do.SysDict, error) {
	q := l.query(ctx)
	return q.Create().
		SetName(dict.Name).
		SetCategory(dict.Category).
		SetColor(dict.Color).
		SetRemark(dict.Remark).
		SetStatus(dict.Status).
		Save(ctx)
}

func (l *sysDictRepoImpl) UpdateDictById(ctx context.Context, id uint32, dict *bo.UpdateSysDictBo) (*do.SysDict, error) {
	q := l.query(ctx)
	return q.Where(do.SysDictFieldID.Eq(int(id))).
		Update().
		SetName(dict.Name).
		SetCategory(dict.Category).
		SetColor(dict.Color).
		SetRemark(dict.Remark).
		SetStatus(dict.Status).
		Save(ctx)
}

func (l *sysDictRepoImpl) BatchUpdateDictStatusByIds(ctx context.Context, status vobj.Status, ids []uint32) error {
	q := l.query(ctx)
	intIds := slices.To(ids, func(id uint32) int { return int(id) })
	q = q.Where(do.SysDictFieldID.In(intIds...))

	return q.Mutation().Update().SetStatus(status).Exec(ctx)
}

func (l *sysDictRepoImpl) DeleteDictByIds(ctx context.Context, ids ...uint32) error {
	q := l.query(ctx)
	intIds := slices.To(ids, func(id uint32) int { return int(id) })
	q = q.Where(do.SysDictFieldID.In(intIds...))
	_, err := q.Delete(ctx)
	return err
}

func (l *sysDictRepoImpl) GetDictById(ctx context.Context, id uint32) (*do.SysDict, error) {
	q := l.query(ctx)
	return q.Where(do.SysDictFieldID.Eq(int(id))).First(ctx)
}

func (l *sysDictRepoImpl) GetDictByIds(ctx context.Context, ids ...uint32) ([]*do.SysDict, error) {
	q := l.query(ctx)
	intIds := slices.To(ids, func(id uint32) int { return int(id) })
	return q.Where(do.SysDictFieldID.In(intIds...)).Find(ctx)
}

func (l *sysDictRepoImpl) ListDict(ctx context.Context, params *bo.ListSysDictBo) ([]*do.SysDict, error) {
	q := l.query(ctx)
	return l.setQuery(q, params).List(ctx, params.Page)
}

func (l *sysDictRepoImpl) SelectDict(ctx context.Context, params *bo.SelectSysDictBo) ([]*do.SysDict, error) {
	q := l.query(ctx).Select(do.SysDictFieldID.String(), do.SysDictFieldName.String(), do.SysDictFieldColor.String())
	return l.setQuery(q, params).Find(ctx)
}

func (l *sysDictRepoImpl) setQuery(q orm.SysDictQuery, params *bo.ListSysDictBo) orm.SysDictQuery {
	var wheres []string

	q = q.Order(do.SysDictFieldCreatedAt.Desc()).Order(do.SysDictFieldUpdatedAt.Desc())
	if pkg.IsNil(params) {
		return q.Where(wheres...)
	}

	if params.Keyword != "" {
		wheres = append(wheres, do.SysDictFieldName.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, do.SysDictFieldStatus.Eq(int(params.Status)))
	}
	if !params.Category.IsUnknown() {
		wheres = append(wheres, do.SysDictFieldCategory.Eq(int(params.Category)))
	}

	return q.Where(wheres...)
}
