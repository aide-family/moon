package repoimpl

import (
	"context"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewTeamResourceRepository 创建资源仓库
func NewTeamResourceRepository(data *data.Data) repository.TeamResource {
	return &teamResourceRepositoryImpl{
		data: data,
	}
}

type teamResourceRepositoryImpl struct {
	data *data.Data
}

func (l *teamResourceRepositoryImpl) CheckPath(ctx context.Context, s string) error {
	// 1. 检查主库API是否存在且开启
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	mainApiDo, err := mainQuery.SysAPI.WithContext(ctx).Where(mainQuery.SysAPI.Path.Eq(s)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nForbidden(ctx)
		}
		return err
	}
	if !mainApiDo.Status.IsEnable() {
		return merr.ErrorI18nForbidden(ctx)
	}

	// 2. 检查业务库API是否存在且开启
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}

	bizApiDo, err := bizQuery.SysTeamAPI.WithContext(ctx).Where(bizQuery.SysTeamAPI.Path.Eq(s)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nForbidden(ctx)
		}
		return err
	}
	if !bizApiDo.Status.IsEnable() {
		return merr.ErrorI18nForbidden(ctx)
	}
	return nil
}

func (l *teamResourceRepositoryImpl) GetByID(ctx context.Context, id uint32) (imodel.IResource, error) {
	mainQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return mainQuery.SysTeamAPI.WithContext(ctx).Where(mainQuery.SysTeamAPI.ID.Eq(id)).First()
}

func (l *teamResourceRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryResourceListParams) ([]imodel.IResource, error) {
	mainQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	apiQuery := mainQuery.SysTeamAPI.WithContext(ctx)

	var wheres []gen.Condition

	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, apiQuery.Or(mainQuery.SysTeamAPI.Name.Like(params.Keyword), mainQuery.SysTeamAPI.Path.Like(params.Keyword)))
	}
	apiQuery = apiQuery.Where(wheres...)
	if !params.Status.IsUnknown() {
		apiQuery = apiQuery.Where(mainQuery.SysTeamAPI.Status.Eq(params.Status.GetValue()))
	}
	if !params.IsAll {
		if err := types.WithPageQuery[bizquery.ISysTeamAPIDo](apiQuery, params.Page); err != nil {
			return nil, err
		}
	}

	list, err := apiQuery.Order(mainQuery.SysTeamAPI.ID.Desc()).Find()
	if !types.IsNil(err) {
		return nil, err
	}
	return types.SliceTo(list, func(api *bizmodel.SysTeamAPI) imodel.IResource { return api }), nil
}

func (l *teamResourceRepositoryImpl) UpdateStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	mainQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = mainQuery.SysTeamAPI.WithContext(ctx).Where(mainQuery.SysTeamAPI.ID.In(ids...)).Update(mainQuery.SysTeamAPI.Status, status)
	return err
}

func (l *teamResourceRepositoryImpl) FindSelectByPage(ctx context.Context, params *bo.QueryResourceListParams) ([]imodel.IResource, error) {
	mainQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	apiQuery := mainQuery.SysTeamAPI.WithContext(ctx)

	if !types.TextIsNull(params.Keyword) {
		apiQuery = apiQuery.Or(mainQuery.SysTeamAPI.Name.Like(params.Keyword), mainQuery.SysTeamAPI.Path.Like(params.Keyword))
	}
	if !params.Status.IsUnknown() {
		apiQuery = apiQuery.Where(mainQuery.SysTeamAPI.Status.Eq(params.Status.GetValue()))
	}
	if !params.IsAll {
		if err := types.WithPageQuery[bizquery.ISysTeamAPIDo](apiQuery, params.Page); err != nil {
			return nil, err
		}
	}

	list, err := apiQuery.Select(mainQuery.SysTeamAPI.ID, mainQuery.SysTeamAPI.Name, mainQuery.SysTeamAPI.Status, mainQuery.SysTeamAPI.DeletedAt).Order(mainQuery.SysTeamAPI.ID.Desc()).Find()
	if !types.IsNil(err) {
		return nil, err
	}
	return types.SliceTo(list, func(api *bizmodel.SysTeamAPI) imodel.IResource { return api }), nil
}
