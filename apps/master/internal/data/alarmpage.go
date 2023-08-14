package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/dal/model"
)

type (
	AlarmPageRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IAlarmPageV1Repo = (*AlarmPageRepo)(nil)

func NewAlarmPageRepo(data *Data, logger log.Logger) *AlarmPageRepo {
	return &AlarmPageRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/AlarmPage"))}
}

func (l *AlarmPageRepo) V1(_ context.Context) string {
	return "AlarmPageRepo.V1"
}

func (l *AlarmPageRepo) CreateAlarmPage(ctx context.Context, m *model.PromAlarmPage) error {
	//TODO implement me
	panic("implement me")
}

func (l *AlarmPageRepo) UpdateAlarmPageById(ctx context.Context, id int32, m *model.PromAlarmPage) error {
	//TODO implement me
	panic("implement me")
}

func (l *AlarmPageRepo) DeleteAlarmPageById(ctx context.Context, id int32) error {
	//TODO implement me
	panic("implement me")
}

func (l *AlarmPageRepo) GetAlarmPageById(ctx context.Context, id int32) (*model.PromAlarmPage, error) {
	//TODO implement me
	panic("implement me")
}

func (l *AlarmPageRepo) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) ([]*model.PromAlarmPage, int64, error) {
	//TODO implement me
	panic("implement me")
}
