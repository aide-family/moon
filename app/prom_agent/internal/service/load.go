package service

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
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

func (s *LoadService) Evaluate(_ context.Context, req *pb.EvaluateRequest) (*pb.EvaluateReply, error) {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	for _, group := range req.GetGroupList() {
		for _, strategyItem := range group.GetStrategies() {
			strategyInfo := &*strategyItem
			eg.Go(func() error {
				d := strategy.NewDatasource(strategy.PrometheusDatasource, strategyInfo.GetEndpoint())
				expr := strategyInfo.GetExpr()
				queryResponse, err := d.Query(context.Background(), expr, time.Now().Unix())
				if err != nil {
					s.log.Error(err)
					return err
				}
				duration := strategyInfo.GetDuration()
				ruleLabels := strategyInfo.GetLabels()
				ruleLabels[strategy.MetricLevelId] = strconv.Itoa(int(strategyInfo.GetAlarmLevelId()))
				alarmInfo := strategy.NewAlarm(&strategy.Group{
					Name: group.GetName(),
					Id:   group.GetId(),
				}, &strategy.Rule{
					Id:          strategyInfo.GetId(),
					Alert:       strategyInfo.GetAlert(),
					Expr:        expr,
					For:         strconv.Itoa(int(duration.GetValue())) + duration.GetUnit(),
					Labels:      ruleLabels,
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
