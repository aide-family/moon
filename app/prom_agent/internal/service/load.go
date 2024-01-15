package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/biz"
)

type LoadService struct {
	pb.UnimplementedLoadServer

	log      *log.Helper
	alarmBiz *biz.AlarmBiz
}

func NewLoadService(alarmBiz *biz.AlarmBiz, logger log.Logger) *LoadService {
	return &LoadService{
		log:      log.NewHelper(log.With(logger, "module", "service.load")),
		alarmBiz: alarmBiz,
	}
}

func (s *LoadService) StrategyGroupAll(ctx context.Context, req *pb.StrategyGroupAllRequest) (*pb.StrategyGroupAllReply, error) {
	s.log.Info("load service StrategyGroupAll")
	return &pb.StrategyGroupAllReply{
		GroupList: []*pb.GroupSimple{
			{
				Id:   1,
				Name: "test",
				Strategies: []*pb.StrategySimple{
					{
						Id:    1,
						Alert: "cpu_90",
						Expr:  "(1 - avg(rate(process_cpu_seconds_total{}[1m])) by (instance))*100 > 90",
						Duration: &api.Duration{
							Value: 3,
							Unit:  "m",
						},
						Labels: map[string]string{
							"test":  "test",
							"test2": "test2",
						},
						Annotations: map[string]string{
							"summary": "test",
							"message": "test",
						},
						GroupId:      1,
						AlarmLevelId: 1,
					},
				},
			},
		},
	}, nil
}

func (s *LoadService) StrategyGroupDiff(ctx context.Context, req *pb.StrategyGroupDiffRequest) (*pb.StrategyGroupDiffReply, error) {
	s.log.Info("load service StrategyGroupDiff")
	return &pb.StrategyGroupDiffReply{}, nil
}

func (s *LoadService) Evaluate(ctx context.Context, req *pb.EvaluateRequest) (*pb.EvaluateReply, error) {
	s.log.Info("load service Evaluate")
	reqBytes, _ := json.Marshal(req)
	s.log.Infof("%s", string(reqBytes))
	return &pb.EvaluateReply{}, nil
}
