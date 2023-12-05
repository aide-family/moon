package endpoint

import (
	"context"
	"encoding/json"

	"github.com/aide-cloud/universal/cipher"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.EndpointRepo = (*endpointRepoImpl)(nil)

type endpointRepoImpl struct {
	repository.UnimplementedEndpointRepo
	log  *log.Helper
	data *data.Data
}

type Map map[string]*bo.EndpointBO

const endpointKey = "prometheus:endpoint"

func (l *endpointRepoImpl) Append(ctx context.Context, endpoints []*bo.EndpointBO) error {
	l.log.Info("endpoints:", endpoints)
	client := l.data.Client()
	// 写入redis hash表中
	args := make([]interface{}, 0, len(endpoints))
	for _, endpoint := range endpoints {
		key := generateKey(endpoint)
		args = append(args, key, endpoint)
	}

	return client.HSet(ctx, endpointKey, args).Err()
}

func (l *endpointRepoImpl) Delete(ctx context.Context, endpoints []*bo.EndpointBO) error {
	client := l.data.Client()
	keys := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		key := generateKey(endpoint)
		keys = append(keys, key)
	}

	return client.HDel(ctx, endpointKey, keys...).Err()
}

func (l *endpointRepoImpl) List(ctx context.Context) ([]*bo.EndpointBO, error) {
	endpointMapInfo, err := l.data.Client().HGetAll(ctx, endpointKey).Result()
	if err != nil {
		return nil, err
	}

	list := make([]*bo.EndpointBO, 0)
	for _, endpointMapString := range endpointMapInfo {
		var endpointBO bo.EndpointBO
		if err = json.Unmarshal([]byte(endpointMapString), &endpointBO); err != nil {
			return nil, err
		}
		list = append(list, &endpointBO)
	}

	return list, nil
}

func generateKey(endpoint *bo.EndpointBO) string {
	return cipher.MD5(endpoint.Uuid)
}

func NewEndpointRepo(data *data.Data, logger log.Logger) repository.EndpointRepo {
	return &endpointRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "endpointRepo")),
		data: data,
	}
}
