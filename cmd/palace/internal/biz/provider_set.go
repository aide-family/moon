package biz

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ProviderSetBiz is biz providers.
var ProviderSetBiz = wire.NewSet(
	NewAuthBiz,
	NewPermissionBiz,
	NewResourceBiz,
	NewUserBiz,
	NewDashboardBiz,
	NewServerBiz,
	NewDict,
	NewTeam,
	NewTeamHook,
	NewMessage,
	NewSystem,
	NewTeamNotice,
	NewTeamDatasource,
	NewTeamStrategy,
	NewTeamStrategyGroupBiz,
	NewTeamStrategyMetric,
	NewLogs,
	NewRealtime,
	NewTeamDatasourceQuery,
)

func NewMetricDatasourceItem(datasourceMetricDo do.DatasourceMetric) *common.MetricDatasourceItem {
	teamDo := datasourceMetricDo.GetTeam()
	return &common.MetricDatasourceItem{
		Team: &common.TeamItem{
			TeamId: teamDo.GetID(),
			Uuid:   teamDo.GetUUID().String(),
		},
		Driver:          common.MetricDatasourceDriver(datasourceMetricDo.GetDriver().GetValue()),
		Prometheus:      &common.MetricDatasourceItem_Prometheus{},
		VictoriaMetrics: &common.MetricDatasourceItem_VictoriaMetrics{},
		Enable:          false,
		Id:              0,
		Name:            datasourceMetricDo.GetName(),
		ScrapeInterval:  &durationpb.Duration{},
	}
}
