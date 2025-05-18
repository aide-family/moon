package biz

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
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
	if validate.IsNil(datasourceMetricDo) {
		return nil
	}
	item := &common.MetricDatasourceItem{
		Team:   nil,
		Driver: common.MetricDatasourceDriver(datasourceMetricDo.GetDriver().GetValue()),
		Config: &common.MetricDatasourceItem_Config{
			Endpoint:  datasourceMetricDo.GetEndpoint(),
			BasicAuth: nil,
			Headers: slices.Map(datasourceMetricDo.GetHeaders(), func(header *kv.KV) *common.KeyValueItem {
				return &common.KeyValueItem{
					Key:   header.Key,
					Value: header.Value,
				}
			}),
			Ca:     datasourceMetricDo.GetCA(),
			Tls:    nil,
			Method: common.DatasourceQueryMethod(datasourceMetricDo.GetQueryMethod().GetValue()),
		},
		Enable:         datasourceMetricDo.GetStatus().IsEnable(),
		Id:             datasourceMetricDo.GetID(),
		Name:           datasourceMetricDo.GetName(),
		ScrapeInterval: durationpb.New(datasourceMetricDo.GetScrapeInterval()),
	}
	if teamDo := datasourceMetricDo.GetTeam(); validate.IsNotNil(teamDo) {
		item.Team = &common.TeamItem{
			TeamId: teamDo.GetID(),
			Uuid:   teamDo.GetUUID().String(),
		}
	}
	if basicAuth := datasourceMetricDo.GetBasicAuth(); validate.IsNotNil(basicAuth) {
		item.Config.BasicAuth = &common.BasicAuth{
			Username: basicAuth.GetUsername(),
			Password: basicAuth.GetPassword(),
		}
	}
	if tls := datasourceMetricDo.GetTLS(); validate.IsNotNil(tls) {
		item.Config.Tls = &common.TLS{
			ServerName: tls.GetServerName(),
			ClientCert: tls.GetClientCert(),
			ClientKey:  tls.GetClientKey(),
			SkipVerify: tls.GetSkipVerify(),
		}
	}
	return item
}
