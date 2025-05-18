package biz

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
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
	NewTimeEngine,
)

func NewMetricDatasourceItem(datasourceMetricDo do.DatasourceMetric) *common.MetricDatasourceItem {
	teamDo := datasourceMetricDo.GetTeam()
	return &common.MetricDatasourceItem{
		Team: &common.TeamItem{
			TeamId: teamDo.GetID(),
			Uuid:   teamDo.GetUUID().String(),
		},
		Driver: common.MetricDatasourceDriver(datasourceMetricDo.GetDriver().GetValue()),
		Config: &common.MetricDatasourceItem_Config{
			Endpoint: datasourceMetricDo.GetEndpoint(),
			BasicAuth: &common.BasicAuth{
				Username: datasourceMetricDo.GetBasicAuth().GetUsername(),
				Password: datasourceMetricDo.GetBasicAuth().GetPassword(),
			},
			Headers: slices.Map(datasourceMetricDo.GetHeaders(), func(header *kv.KV) *common.KeyValueItem {
				return &common.KeyValueItem{
					Key:   header.Key,
					Value: header.Value,
				}
			}),
			Ca:     datasourceMetricDo.GetCA(),
			Tls:    &common.TLS{},
			Method: common.DatasourceQueryMethod(datasourceMetricDo.GetQueryMethod().GetValue()),
		},
		Enable:         datasourceMetricDo.GetStatus().IsEnable(),
		Id:             datasourceMetricDo.GetID(),
		Name:           datasourceMetricDo.GetName(),
		ScrapeInterval: durationpb.New(datasourceMetricDo.GetScrapeInterval()),
	}
}
