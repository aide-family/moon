package endpoint

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/basescopes"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.EndpointRepo = (*endpointRepoImpl)(nil)

type endpointRepoImpl struct {
	repository.UnimplementedEndpointRepo
	log  *log.Helper
	data *data.Data

	query.IAction[model.Endpoint]
}

func (l *endpointRepoImpl) Append(ctx context.Context, endpoint *bo.EndpointBO) (*bo.EndpointBO, error) {
	newModelData := endpoint.ToModel()
	if err := l.WithContext(ctx).Create(newModelData); err != nil {
		return nil, err
	}
	return bo.EndpointModelToBO(newModelData), nil
}

func (l *endpointRepoImpl) Update(ctx context.Context, endpoint *bo.EndpointBO) (*bo.EndpointBO, error) {
	newModelData := endpoint.ToModel()
	// 查询详情
	detail, err := l.WithContext(ctx).FirstByID(newModelData.ID)
	if err != nil {
		return nil, err
	}
	// 执行更新
	if err = l.WithContext(ctx).UpdateByID(detail.ID, newModelData); err != nil {
		return nil, err
	}

	return bo.EndpointModelToBO(newModelData), nil
}

func (l *endpointRepoImpl) UpdateStatus(ctx context.Context, ids []uint32, status valueobj.Status) error {
	if len(ids) == 0 {
		return nil
	}
	return l.WithContext(ctx).Update(&model.Endpoint{Status: status}, basescopes.InIds(ids...))
}

func (l *endpointRepoImpl) Delete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}
	return l.WithContext(ctx).Delete(basescopes.InIds(ids...))
}

func (l *endpointRepoImpl) List(ctx context.Context, pagination query.Pagination, scopes ...query.ScopeMethod) ([]*bo.EndpointBO, error) {
	endpointList, err := l.WithContext(ctx).List(pagination, scopes...)
	if err != nil {
		return nil, err
	}
	boList := slices.To(endpointList, func(endpoint *model.Endpoint) *bo.EndpointBO {
		return bo.EndpointModelToBO(endpoint)
	})

	return boList, nil
}

func (l *endpointRepoImpl) Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.EndpointBO, error) {
	detail, err := l.WithContext(ctx).First(scopes...)
	if err != nil {
		return nil, err
	}
	return bo.EndpointModelToBO(detail), nil
}

func NewEndpointRepo(data *data.Data, logger log.Logger) repository.EndpointRepo {
	return &endpointRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.endpointRepo")),
		data: data,
		IAction: query.NewAction[model.Endpoint](
			query.WithDB[model.Endpoint](data.DB()),
		),
	}
}
