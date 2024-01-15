package service

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/api"
	pb "prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/biz"
	"prometheus-manager/pkg/strategy"
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

func (s *LoadService) Evaluate(_ context.Context, req *pb.EvaluateRequest) (*pb.EvaluateReply, error) {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	for _, group := range req.GetGroupList() {
		for _, strategyInfo := range group.GetStrategies() {
			eg.Go(func() error {
				d := strategy.NewDatasource(strategy.PrometheusDatasource, strategyInfo.GetEndpoint())
				expr := strategyInfo.GetExpr()
				queryResponse, err := d.Query(context.Background(), expr, time.Now().Unix())
				if err != nil {
					s.log.Error(err)
					return err
				}
				duration := strategyInfo.GetDuration()
				alarmInfo := strategy.NewAlarm(&strategy.Group{
					Name: group.GetName(),
					Id:   group.GetId(),
				}, &strategy.Rule{
					Id:          strategyInfo.GetId(),
					Alert:       strategyInfo.GetAlert(),
					Expr:        expr,
					For:         strconv.Itoa(int(duration.GetValue())) + duration.GetUnit(),
					Labels:      strategyInfo.GetLabels(),
					Annotations: strategyInfo.GetAnnotations(),
				}, queryResponse.Data.Result)
				if err = s.alarmBiz.SendAlarm(alarmInfo); err != nil {
					s.log.Error("load service Evaluate", "group", group.GetName(), "strategy", strategyInfo.GetAlert(), "error", err)
					return err

				}
				return nil
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &pb.EvaluateReply{}, nil
}
