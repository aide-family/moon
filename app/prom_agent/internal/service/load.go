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
			Id:                  1,
			Name:                "test-alarm-group",
			Categories:          nil,
			Status:              0,
			Remark:              "",
			CreatedAt:           0,
			UpdatedAt:           0,
			DeletedAt:           0,
			StrategyCount:       0,
			EnableStrategyCount: 0,
			Strategies: []*api.PromStrategyV1{{
				Id:    1,
				Alert: "UpAlertTest",
				Expr:  "up == 1",
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
				Status:         0,
				GroupId:        1,
				GroupInfo:      nil,
				AlarmLevelId:   1,
				AlarmLevelInfo: nil,
				AlarmPageIds:   nil,
				AlarmPageInfo:  nil,
				CategoryIds:    nil,
				CategoryInfo:   nil,
				CreatedAt:      0,
				UpdatedAt:      0,
				DeletedAt:      0,
				Remark:         "",
				DataSource: &api.PrometheusServerSelectItem{
					Value:    1,
					Label:    "PrometheusServer",
					Status:   0,
					Remark:   "",
					Endpoint: "https://prom-server.aide-cloud.cn/",
				},
				DataSourceId: 1,
				MaxSuppress:  nil,
				SendInterval: nil,
				SendRecover:  false,
			}},
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
