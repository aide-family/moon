package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data/microserver"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service/build"

	"github.com/go-kratos/kratos/v2/log"
)

// NewAlertRepository 实例化alert
func NewAlertRepository(data *data.Data, palaceCli *microserver.PalaceConn) repository.Alert {
	return &alertRepositoryImpl{data: data, palaceCli: palaceCli}
}

type alertRepositoryImpl struct {
	data      *data.Data
	palaceCli *microserver.PalaceConn
}

func (a *alertRepositoryImpl) PushAlarm(ctx context.Context, alarm *bo.Alarm) error {
	in := build.NewAlarmBuilder(alarm).ToAPI()
	pushAlarm, err := a.palaceCli.PushAlarm(ctx, in)
	if err != nil {
		log.Errorw("method", "PushAlarm", "err", err)
		return err
	}
	log.Debugw("pushAlarm", pushAlarm)
	return nil
}

func (a *alertRepositoryImpl) SaveAlarm(_ context.Context, alarm *bo.Alarm) error {
	return a.data.GetAlertQueue().Push(alarm.Message())
}
