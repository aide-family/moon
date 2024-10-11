package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	// AlarmRaw repository
	AlarmRaw interface {
		// CreateAlarmRaws 创建告警原始数据
		CreateAlarmRaws(ctx context.Context, params []*bo.CreateAlarmRawParams, teamID uint32) ([]*alarmmodel.AlarmRaw, error)
		// GetTeamStrategy 获取团队策略
		GetTeamStrategy(ctx context.Context, params *bo.GetTeamStrategyParams) (*bizmodel.Strategy, error)
		// GetStrategyLevel 获取策略等级
		GetStrategyLevel(ctx context.Context, params *bo.GetTeamStrategyLevelParams) (*bizmodel.StrategyLevel, error)
		// ListDatasource 获取数据源
		ListDatasource(ctx context.Context, params *bo.GetTeamDatasourceParams) ([]*bizmodel.Datasource, error)
	}
)
