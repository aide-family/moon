package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen/field"
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

func (r *alarmRawRepositoryImpl) GetTeamStrategy(ctx context.Context, params *bo.GetTeamStrategyParams) (*bizmodel.Strategy, error) {
	bizQuery, err := getTeamIdBizQuery(r.data, params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}

	return bizQuery.Strategy.WithContext(ctx).Preload(field.Associations).Where(bizQuery.Strategy.ID.Eq(params.StrategyID)).First()
}

func (r *alarmRawRepositoryImpl) GetStrategyLevel(ctx context.Context, params *bo.GetTeamStrategyLevelParams) (*bizmodel.StrategyLevel, error) {
	bizQuery, err := getTeamIdBizQuery(r.data, params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}

	return bizQuery.StrategyLevel.WithContext(ctx).Preload(field.Associations).Where(bizQuery.StrategyLevel.ID.Eq(params.LevelID)).First()
}

func (r *alarmRawRepositoryImpl) ListDatasource(ctx context.Context, params *bo.GetTeamDatasourceParams) ([]*bizmodel.Datasource, error) {
	bizQuery, err := getTeamIdBizQuery(r.data, params.TeamID)
	if !types.IsNil(err) {
		return nil, err
	}

	return bizQuery.Datasource.WithContext(ctx).Preload(field.Associations).Where(bizQuery.Datasource.ID.In(params.DatasourceIds...)).Find()
}

func (r *alarmRawRepositoryImpl) CreateAlarmRaws(ctx context.Context, param []*bo.CreateAlarmRawParams, teamID uint32) ([]*alarmmodel.AlarmRaw, error) {
	alarmQuery, err := getTeamBizAlarmQuery(teamID, r.data)
	if !types.IsNil(err) {
		return nil, err
	}

	alarmRawModels := types.SliceTo(param, func(item *bo.CreateAlarmRawParams) *alarmmodel.AlarmRaw {
		return &alarmmodel.AlarmRaw{RawInfo: item.RawInfo, Fingerprint: item.Fingerprint}
	})
	err = alarmQuery.AlarmRaw.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(alarmRawModels, len(alarmRawModels))
	if err != nil {
		log.Error("AlarmRaw CreateInBatches err: ", err.Error())
		return nil, err
	}
	return alarmRawModels, nil
}
