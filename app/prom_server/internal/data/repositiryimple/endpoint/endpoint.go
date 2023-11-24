package endpoint

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.EndpointRepo = (*endpointRepoImpl)(nil)

type endpointRepoImpl struct {
	repository.UnimplementedEndpointRepo
	log  *log.Helper
	data *data.Data
}

const endpointKey = "prometheus:endpoint"

func (l *endpointRepoImpl) Append(ctx context.Context, endpoints []*dobo.EndpointDO) error {
	client := l.data.Client()
	// 写入redis hash表中
	endpointMap := endpointsToMap(endpoints)

	return client.HSet(ctx, endpointKey, endpointMap).Err()
}

func (l *endpointRepoImpl) Delete(ctx context.Context, endpoints []*dobo.EndpointDO) error {
	client := l.data.Client()
	keys := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		key := generateKey(endpoint)
		keys = append(keys, key)
	}

	return client.HDel(ctx, endpointKey, keys...).Err()
}

func (l *endpointRepoImpl) List(ctx context.Context) ([]*dobo.EndpointDO, error) {
	list := make([]*dobo.EndpointDO, 0)
	endpointMap := make(map[string]*dobo.EndpointDO)
	if err := l.data.Client().HGetAll(ctx, endpointKey).Scan(&endpointMap); err != nil {
		return nil, err
	}
	for _, endpoint := range endpointMap {
		list = append(list, endpoint)
	}

	return list, nil
}

func generateKey(endpoint *dobo.EndpointDO) string {
	return endpoint.Endpoint
}

func endpointsToMap(endpoints []*dobo.EndpointDO) map[string]*dobo.EndpointDO {
	return slices.ToMap(endpoints, generateKey)
}

func NewEndpointRepo(data *data.Data, logger log.Logger) repository.EndpointRepo {
	return &endpointRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "endpointRepo")),
		data: data,
	}
}
