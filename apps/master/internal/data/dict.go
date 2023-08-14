package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/dal/model"
)

type (
	DictRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IDictV1Repo = (*DictRepo)(nil)

func NewDictRepo(data *Data, logger log.Logger) *DictRepo {
	return &DictRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Dict"))}
}

func (l *DictRepo) V1(_ context.Context) string {
	return "DictRepo.V1"
}

func (l *DictRepo) CreateDict(ctx context.Context, m *model.PromDict) error {
	//TODO implement me
	panic("implement me")
}

func (l *DictRepo) UpdateDictById(ctx context.Context, id int32, m *model.PromDict) error {
	//TODO implement me
	panic("implement me")
}

func (l *DictRepo) DeleteDictById(ctx context.Context, id int32) error {
	//TODO implement me
	panic("implement me")
}

func (l *DictRepo) GetDictById(ctx context.Context, id int32) (*model.PromDict, error) {
	//TODO implement me
	panic("implement me")
}

func (l *DictRepo) ListDict(ctx context.Context, req *pb.ListDictRequest) ([]*model.PromDict, int64, error) {
	//TODO implement me
	panic("implement me")
}
