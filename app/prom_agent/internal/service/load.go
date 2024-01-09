package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/agent"
)

type LoadService struct {
	pb.UnimplementedLoadServer

	log *log.Helper
}

func NewLoadService(logger log.Logger) *LoadService {
	return &LoadService{
		log: log.NewHelper(log.With(logger, "module", "service.load")),
	}
}

func (s *LoadService) StrategyGroupAll(ctx context.Context, req *pb.StrategyGroupAllRequest) (*pb.StrategyGroupAllReply, error) {
	s.log.Info("load service StrategyGroupAll")
	return &pb.StrategyGroupAllReply{
		Items: []*api.PromGroup{{
			Id:   1,
			Name: "test-alarm-group",
			// TODO 放进告警label里面
			Categories: nil,
			Strategies: []*api.PromStrategyV1{
				{
					Id:           1,
					Alert:        "UpAlertTest",
					Expr:         "up == 1",
					AlarmLevelId: 1,
					Duration: &api.Duration{
						Value: 30,
						Unit:  "s",
					},
					Labels: map[string]string{
						"job": "prometheus",
					},
					Annotations: map[string]string{
						"summary":     "up == 1",
						"description": "{{ $label.instance }} up == 1, value {{ $value }}",
					},
					DataSource: &api.PrometheusServerSelectItem{
						Value:    1,
						Label:    "PrometheusServer",
						Endpoint: "https://prom-server.aide-cloud.cn/",
					},
				},
				{
					Id:           2,
					Alert:        "RateCpuTotal",
					Expr:         "rate(process_cpu_seconds_total[5m])",
					AlarmLevelId: 2,
					Duration: &api.Duration{
						Value: 30,
						Unit:  "s",
					},
					Labels: map[string]string{
						"job": "prometheus",
					},
					Annotations: map[string]string{
						"summary":     "rate(process_cpu_seconds_total[5m])",
						"description": "{{ $label.instance }} cpu rate total, value {{ $value }}",
					},
					DataSource: &api.PrometheusServerSelectItem{
						Value:    1,
						Label:    "PrometheusServer",
						Endpoint: "https://prom-server.aide-cloud.cn/",
					},
				},
			},
		}},
		Page: &api.PageReply{
			Curr:  1,
			Size:  10,
			Total: 1,
		},
	}, nil
}

func (s *LoadService) StrategyGroupDiff(ctx context.Context, req *pb.StrategyGroupDiffRequest) (*pb.StrategyGroupDiffReply, error) {
	s.log.Info("load service StrategyGroupDiff")
	return &pb.StrategyGroupDiffReply{}, nil
}
