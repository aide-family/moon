package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm/clause"
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

func (r *alarmRawRepositoryImpl) CreateAlarmRaws(ctx context.Context, param []*bo.CreateAlarmRawParams, teamID uint32) ([]*alarmmodel.AlarmRaw, error) {
	alarmQuery, err := getTeamBizAlarmQuery(teamID, r.data)
	if !types.IsNil(err) {
		return nil, err
	}

	alarmRawModels := types.SliceTo(param, func(item *bo.CreateAlarmRawParams) *alarmmodel.AlarmRaw {
		return &alarmmodel.AlarmRaw{
			Receiver:    item.Receiver,
			RawInfo:     item.RawInfo,
			Fingerprint: item.Fingerprint,
		}
	})
	columns := []string{"receiver", "raw_info"}
	err = alarmQuery.AlarmRaw.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "fingerprint"}},
			DoUpdates: clause.AssignmentColumns(columns),
		}).
		CreateInBatches(alarmRawModels, len(alarmRawModels))
	if err != nil {
		log.Error("AlarmRaw CreateInBatches err: ", err.Error())
		return nil, err
	}
	return alarmRawModels, nil
}
