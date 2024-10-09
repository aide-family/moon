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
		// CreateAlarmRaw 创建报警规则
		CreateAlarmRaw(ctx context.Context, params *bo.CreateAlarmRawParams) (*alarmmodel.AlarmRaw, error)
		// GetTeamStrategy 获取团队策略
		GetTeamStrategy(ctx context.Context, params *bo.GetTeamStrategyParams) (*bizmodel.Strategy, error)
		// GetStrategyLevel 获取策略等级
		GetStrategyLevel(ctx context.Context, params *bo.GetTeamStrategyLevelParams) (*bizmodel.StrategyLevel, error)
		// ListDatasource 获取数据源
		ListDatasource(ctx context.Context, params *bo.GetTeamDatasourceParams) ([]*bizmodel.Datasource, error)
	}
)
