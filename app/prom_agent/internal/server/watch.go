package server

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/api"
	"prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/biz/bo"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/service"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/strategy"
)

var _ transport.Server = (*Watch)(nil)

type Watch struct {
	conf   *conf.WatchProm
	log    *log.Helper
	exitCh chan struct{}
	ticker *time.Ticker

	alerting     strategy.Alerter
	loadService  *service.LoadService
	alarmService *service.AlarmService

	groups *sync.Map
}

func NewWatch(
	c *conf.WatchProm,
	loadService *service.LoadService,
	alarmService *service.AlarmService,
	logger log.Logger,
) *Watch {
	return &Watch{
		conf:         c,
		exitCh:       make(chan struct{}, 1),
		ticker:       time.NewTicker(c.GetInterval().AsDuration()),
		log:          log.NewHelper(log.With(logger, "module", "server.watch")),
		alerting:     strategy.NewAlerting(logger),
		loadService:  loadService,
		alarmService: alarmService,
		groups:       new(sync.Map),
	}
}

func (w *Watch) Start(ctx context.Context) error {
	if err := w.loadGroups(ctx); err != nil {
		w.log.Errorf("[Watch] load groups error: %v", err)
		return err
	}
	go func() {
		defer after.Recover(w.log)
		for {
			select {
			case <-w.exitCh:
				w.shutdown()
				return
			case <-w.ticker.C:
				w.eval()
			}
		}
	}()
	w.log.Info("[Watch] server started")
	return nil
}

func (w *Watch) eval() {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	w.groups.Range(func(key, value any) bool {
		eg.Go(func() error {
			group, ok := value.(*strategy.Group)
			if !ok {
				return nil
			}
			alarmGroup := &agent.AlarmGroup{
				StrategyGroupId: key.(uint32),
				Alarms:          make([]*agent.AlarmInfo, 0, len(group.Rules)),
			}
			for _, rule := range group.Rules {
				results, err := w.alerting.Eval(context.Background(), group, rule)
				if err != nil {
					w.log.Warnf("[Watch] server tick error: %v", err)
					// TODO push 错误规则数据
					continue
				}
				// 更新告警缓存
				w.log.Info("[Watch] server tick results: %v", len(results))
				alarmList := make([]*agent.AlarmInfo, 0, len(results))
				for _, result := range results {
					timeUnix := result.Value[0].(float64)
					metricValue, _ := strconv.ParseFloat(result.Value[1].(string), 64)
					alarmItem := &agent.AlarmInfo{
						Metric:     result.Metric,
						Value:      []float64{timeUnix, metricValue},
						StrategyId: rule.Labels.StrategyId(),
						Expr:       rule.Expr,
						Duration:   bo.BuildApiDuration(rule.For),
					}
					alarmList = append(alarmList, alarmItem)
				}
				alarmGroup.Alarms = append(alarmGroup.Alarms, alarmList...)
			}

			_, err := w.alarmService.Push(context.Background(), &agent.PushRequest{
				Group: alarmGroup,
			})
			if err != nil {
				w.log.Warnf("[Watch] server tick error: %v", err)
			}
			return nil
		})
		return true
	})
	_ = eg.Wait()
}

func (w *Watch) loadGroups(ctx context.Context) error {
	curr := int32(1)
	size := int32(100)
	// 启动之前, 加载全部策略
	for {
		strategyGroupAllReply, err := w.loadService.StrategyGroupAll(ctx, &agent.StrategyGroupAllRequest{Page: &api.PageRequest{
			Curr: curr,
			Size: size,
		}})
		if err != nil {
			return err
		}
		pg := strategyGroupAllReply.GetPage()
		list := strategyGroupAllReply.GetItems()
		for _, groupItem := range list {
			ruleGroupItem := &strategy.Group{
				Name:  groupItem.GetName(),
				Rules: make([]*strategy.Rule, 0, len(groupItem.GetStrategies())),
			}

			for _, strategyItem := range groupItem.GetStrategies() {
				duration := strategyItem.GetDuration()
				labels := strategyItem.GetLabels()
				annotations := strategyItem.GetAnnotations()
				labels[strategy.LabelKeyStrategyId] = strconv.FormatInt(int64(strategyItem.GetId()), 10)
				labels[strategy.LabelKeyLevelId] = strconv.FormatInt(int64(strategyItem.GetAlarmLevelId()), 10)

				ruleItem := &strategy.Rule{
					Alert:       strategyItem.GetAlert(),
					Expr:        strategyItem.GetExpr(),
					For:         strconv.FormatInt(duration.GetValue(), 10) + duration.GetUnit(),
					Labels:      labels,
					Annotations: annotations,
				}
				ruleItem.SetEndpoint(strategyItem.GetDataSource().GetEndpoint())
				ruleGroupItem.Rules = append(ruleGroupItem.Rules, ruleItem)
			}
			w.groups.Store(groupItem.GetId(), ruleGroupItem)
		}
		if pg.GetTotal() == 0 || len(list) < int(size) {
			// 如果没有更多数据, 则退出
			break
		}
	}
	return nil
}

func (w *Watch) Stop(_ context.Context) error {
	w.log.Info("[Watch] server stopping")
	close(w.exitCh)
	return nil
}

func (w *Watch) shutdown() {
	w.groups = nil
	w.ticker.Stop()
	w.log.Info("[Watch] server stopped")
}
