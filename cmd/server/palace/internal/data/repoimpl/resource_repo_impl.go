package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/query"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewResourceRepo(data *data.Data) repo.ResourceRepo {
	return &resourceRepoImpl{
		data: data,
	}
}

type resourceRepoImpl struct {
	data *data.Data
}

func (l *resourceRepoImpl) GetById(ctx context.Context, id uint32) (*model.SysAPI, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysAPI.Where(query.SysAPI.ID.Eq(id)).First()
}

func (l *resourceRepoImpl) FindByPage(ctx context.Context, params *bo.QueryResourceListParams) ([]*model.SysAPI, error) {
	q := query.Use(l.data.GetMainDB(ctx)).SysAPI.WithContext(ctx)

	if !types.TextIsNull(params.Keyword) {
		q = q.Or(query.SysAPI.Name.Like(params.Keyword), query.SysAPI.Path.Like(params.Keyword))
	}
	if !types.IsNil(params.Page) {
		page := params.Page
		total, err := q.Count()
		if err != nil {
			return nil, err
		}
		page.SetTotal(int(total))
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			q = q.Limit(pageSize)
		} else {
			q = q.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return q.Order(query.SysAPI.ID.Desc()).Find()
}

func (l *resourceRepoImpl) UpdateStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).SysAPI.Where(query.SysAPI.ID.In(ids...)).Update(query.SysAPI.Status, status)
	return err
}

func (l *resourceRepoImpl) FindSelectByPage(ctx context.Context, params *bo.QueryResourceListParams) ([]*model.SysAPI, error) {
	q := query.Use(l.data.GetMainDB(ctx)).SysAPI.WithContext(ctx)

	if !types.TextIsNull(params.Keyword) {
		q = q.Or(query.SysAPI.Name.Like(params.Keyword), query.SysAPI.Path.Like(params.Keyword))
	}
	if !types.IsNil(params.Page) {
		page := params.Page
		total, err := q.Count()
		if err != nil {
			return nil, err
		}
		page.SetTotal(int(total))
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			q = q.Limit(pageSize)
		} else {
			q = q.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return q.Select(query.SysAPI.ID, query.SysAPI.Name, query.SysAPI.Status, query.SysAPI.DeletedAt).Order(query.SysAPI.ID.Desc()).Find()
}
