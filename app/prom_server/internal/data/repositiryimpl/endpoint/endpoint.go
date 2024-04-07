package endpoint

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/strategy"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.EndpointRepo = (*endpointRepoImpl)(nil)

type endpointRepoImpl struct {
	repository.UnimplementedEndpointRepo
	log  *log.Helper
	data *data.Data
}

func (l *endpointRepoImpl) GetByParams(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.EndpointBO, error) {
	var list []*do.Endpoint
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Find(&list).Error; err != nil {
		return nil, err
	}
	return slices.To(list, func(item *do.Endpoint) *bo.EndpointBO { return bo.EndpointModelToBO(item) }), nil
}

func (l *endpointRepoImpl) Append(ctx context.Context, endpoint *bo.CreateEndpointReq) (*bo.EndpointBO, error) {
	newModelData := &do.Endpoint{
		Name:      endpoint.Name,
		Endpoint:  endpoint.Endpoint,
		Remark:    endpoint.Remark,
		BasicAuth: strategy.NewBasicAuth(endpoint.Username, endpoint.Password),
	}
	if err := l.data.DB().WithContext(ctx).Create(newModelData).Error; err != nil {
		return nil, err
	}
	return bo.EndpointModelToBO(newModelData), nil
}

func (l *endpointRepoImpl) Update(ctx context.Context, endpoint *bo.EndpointBO) (*bo.EndpointBO, error) {
	newModelData := endpoint.ToModel()
	var detail do.Endpoint
	// 查询详情
	if err := l.data.DB().WithContext(ctx).First(&detail, endpoint.Id).Error; err != nil {
		return nil, err
	}
	// 执行更新
	if err := l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(detail.ID)).Updates(newModelData).Error; err != nil {
		return nil, err
	}

	return bo.EndpointModelToBO(newModelData), nil
}

func (l *endpointRepoImpl) UpdateStatus(ctx context.Context, ids []uint32, status vobj.Status) error {
	if len(ids) == 0 {
		return nil
	}
	return l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(ids...)).Updates(&do.Endpoint{Status: status}).Error
}

func (l *endpointRepoImpl) Delete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	return l.data.DB().WithContext(ctx).Scopes(basescopes.InIds(ids...)).Delete(&do.Endpoint{}).Error
}

func (l *endpointRepoImpl) List(ctx context.Context, pagination bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.EndpointBO, error) {
	var endpointList []*do.Endpoint
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, bo.Page(pagination))...).Find(&endpointList).Error; err != nil {
		return nil, err
	}
	if pagination != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Scopes(scopes...).Model(&do.Endpoint{}).Count(&total).Error; err != nil {
			return nil, err
		}
		pagination.SetTotal(total)
	}
	boList := slices.To(endpointList, func(endpoint *do.Endpoint) *bo.EndpointBO {
		return bo.EndpointModelToBO(endpoint)
	})

	return boList, nil
}

func (l *endpointRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.EndpointBO, error) {
	var detail do.Endpoint
	if err := l.data.DB().WithContext(ctx).Scopes(scopes...).First(&detail).Error; err != nil {
		return nil, err
	}
	return bo.EndpointModelToBO(&detail), nil
}

func NewEndpointRepo(data *data.Data, logger log.Logger) repository.EndpointRepo {
	return &endpointRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.endpointRepo")),
		data: data,
	}
}
