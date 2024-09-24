package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewAlarmRawRepository 创建 AlarmRawRepository
func NewAlarmRawRepository(data *data.Data) repository.AlarmRaw {
	return &alarmRawRepositoryImpl{
		data: data,
	}
}

type alarmRawRepositoryImpl struct {
	data *data.Data
}

func (r *alarmRawRepositoryImpl) CreateAlarmRaw(ctx context.Context, param *bo.CreateAlarmRawParams) (*alarmmodel.AlarmRaw, error) {
	bizQuery, err := getBizAlarmQuery(ctx, r.data)
	if !types.IsNil(err) {
		return nil, err
	}
	alarmRaw := &alarmmodel.AlarmRaw{
		RawInfo:     param.RawInfo,
		Fingerprint: param.Fingerprint,
	}
	err = bizQuery.AlarmRaw.WithContext(ctx).Create(alarmRaw)
	if !types.IsNil(err) {
		return nil, err
	}
	return alarmRaw, nil
}
