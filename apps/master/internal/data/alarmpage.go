package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/biz"
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
