package repoimpl

import (
	"context"

	"gorm.io/gen"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewResourceRepository 创建资源仓库
func NewResourceRepository(data *data.Data) repository.Resource {
	return &resourceRepositoryImpl{
		data: data,
	}
}

type resourceRepositoryImpl struct {
	data *data.Data
}

func (l *resourceRepositoryImpl) GetByID(ctx context.Context, id uint32) (*model.SysAPI, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysAPI.Where(query.SysAPI.ID.Eq(id)).First()
}

func (l *resourceRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryResourceListParams) ([]*model.SysAPI, error) {
	q := query.Use(l.data.GetMainDB(ctx)).SysAPI.WithContext(ctx)

	var wheres []gen.Condition

	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, q.Or(query.SysAPI.Name.Like(params.Keyword), query.SysAPI.Path.Like(params.Keyword)))
	}
	q = q.Where(wheres...)
	if err := types.WithPageQuery[query.ISysAPIDo](q, params.Page); err != nil {
		return nil, err
	}
	return q.Order(query.SysAPI.ID.Desc()).Find()
}

func (l *resourceRepositoryImpl) UpdateStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).SysAPI.Where(query.SysAPI.ID.In(ids...)).Update(query.SysAPI.Status, status)
	return err
}

func (l *resourceRepositoryImpl) FindSelectByPage(ctx context.Context, params *bo.QueryResourceListParams) ([]*model.SysAPI, error) {
	q := query.Use(l.data.GetMainDB(ctx)).SysAPI.WithContext(ctx)

	if !types.TextIsNull(params.Keyword) {
		q = q.Or(query.SysAPI.Name.Like(params.Keyword), query.SysAPI.Path.Like(params.Keyword))
	}
	if err := types.WithPageQuery[query.ISysAPIDo](q, params.Page); err != nil {
		return nil, err
	}
	return q.Select(query.SysAPI.ID, query.SysAPI.Name, query.SysAPI.Status, query.SysAPI.DeletedAt).Order(query.SysAPI.ID.Desc()).Find()
}
