package repoimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
)

// NewStrategyRepository 创建策略操作器
func NewStrategyRepository(data *data.Data) repository.Strategy {
	return &strategyRepositoryImpl{data: data}
}

type strategyRepositoryImpl struct {
	data *data.Data
}

// Save 保存策略
func (s *strategyRepositoryImpl) Save(_ context.Context, strategies []*bo.Strategy) error {
	queue := s.data.GetStrategyQueue()
	go func() {
		defer after.RecoverX()
		for _, strategyItem := range strategies {
			if err := queue.Push(strategyItem.Message()); err != nil {
				log.Errorw("method", "queue.push", "error", err)
			}
		}
	}()

	return nil
}

func (s *strategyRepositoryImpl) getDatasourceCliList(strategy *bo.Strategy) ([]datasource.Datasource, error) {
	datasourceList := strategy.Datasource
	datasourceCliList := make([]datasource.Datasource, 0, len(datasourceList))
	category := datasourceList[0].Category
	for _, datasourceItem := range datasourceList {
		if datasourceItem.Category != category {
			log.Warnw("method", "Eval", "error", "datasource category is not same")
			continue
		}
		cfg := &api.Datasource{
			Category:    api.DatasourceType(datasourceItem.Category),
			StorageType: api.StorageType(datasourceItem.StorageType),
			Config:      datasourceItem.Config,
			Endpoint:    datasourceItem.Endpoint,
		}
		newDatasource, err := datasource.NewDatasource(cfg)
		if err != nil {
			log.Warnw("method", "NewDatasource", "error", err)
			continue
		}
		datasourceCliList = append(datasourceCliList, newDatasource)
	}
	if len(datasourceCliList) == 0 {
		return nil, merr.ErrorNotification("datasource is empty")
	}
	return datasourceCliList, nil
}

func builderAlarmBaseInfo(strategy *bo.Strategy) *bo.Alarm {
	strategy.Labels.Append(vobj.StrategyID, fmt.Sprintf("%d", strategy.ID))

	alarmInfo := bo.Alarm{
		Receiver:          "",
		Status:            vobj.AlertStatusFiring,
		Alerts:            nil,
		GroupLabels:       strategy.Labels,
		CommonLabels:      strategy.Labels,
		CommonAnnotations: strategy.Annotations,
		ExternalURL:       "",
		Version:           env.Version(),
		GroupKey:          "",
		TruncatedAlerts:   0,
	}
	return &alarmInfo
}

// Eval 评估策略 告警/恢复
func (s *strategyRepositoryImpl) Eval(ctx context.Context, strategy *bo.Strategy) (*bo.Alarm, error) {
	datasourceCliList, err := s.getDatasourceCliList(strategy)
	if err != nil {
		return nil, err
	}
	alarmInfo := builderAlarmBaseInfo(strategy)
	var alerts []*bo.Alert
	for _, cli := range datasourceCliList {
		evalPoints, err := cli.Eval(ctx, strategy.Expr, strategy.For)
		if err != nil {
			log.Warnw("method", "Eval", "error", err)
			continue
		}
		if len(evalPoints) == 0 {
			// 生成恢复事件（如果有存在的告警），并补充当前时间为告警恢复事件
			continue
		}

		for index, point := range evalPoints {
			labels, ok := index.(*vobj.Labels)
			if !ok {
				continue
			}
			if !isCompletelyMeet(point.Values, strategy) {
				continue
			}
			valLength := len(point.Values)
			endPointValue := point.Values[valLength-1]

			formatValue := map[string]any{
				"value":  endPointValue.Value,
				"time":   endPointValue.Timestamp,
				"labels": labels.Map(),
			}
			annotations := make(vobj.Annotations, len(strategy.Annotations))

			for key, annotation := range strategy.Annotations {
				annotations[key] = format.Formatter(annotation, formatValue)
			}
			alert := &bo.Alert{
				Status:       vobj.AlertStatusFiring,
				Labels:       point.Labels, // 合并label
				Annotations:  annotations,  // 填充
				StartsAt:     types.NewTimeByUnix(endPointValue.Timestamp),
				EndsAt:       nil,
				GeneratorURL: "", // 生成事件图表链接
				Fingerprint:  "", // TODO 指纹生成逻辑
				Value:        endPointValue.Value,
			}
			alert = s.getFiringAlert(ctx, alert)
			alerts = append(alerts, alert)
			alarmInfo.CommonLabels = findCommonKeys([]*vobj.Labels{alarmInfo.CommonLabels, labels}...)
		}
	}
	alertIndexList := types.SliceToWithFilter(alerts, func(alert *bo.Alert) (string, bool) {
		return alert.Index(), alert.Status.IsFiring()
	})
	// 获取存在的告警标识列表
	alertsStr, _ := s.data.GetCacher().Get(ctx, strategy.Index())
	if !types.TextIsNull(alertsStr) {
		existAlerts := strings.Split(alertsStr, ",")
		alertIndexListMap := make(map[string]struct{}, len(alerts))
		for _, alertItem := range alerts {
			alertIndexListMap[alertItem.Index()] = struct{}{}
		}
		for _, existAlert := range existAlerts {
			if _, ok := alertIndexListMap[existAlert]; !ok {
				// 获取存在的告警信息
				existAlertStr, err := s.data.GetCacher().Get(ctx, existAlert)
				if err != nil {
					continue
				}

				var existAlertItem bo.Alert
				if err := json.Unmarshal([]byte(existAlertStr), &existAlertItem); err != nil {
					continue
				}
				// 获取存在的告警信息， 并且更新为告警恢复状态
				existAlertItem.Status = vobj.AlertStatusResolved
				existAlertItem.EndsAt = types.NewTimeByUnix(time.Now().Unix())
				alerts = append(alerts, &existAlertItem)
			}
		}
	}

	if len(alerts) == 0 {
		return nil, merr.ErrorNotification("no data")
	}
	if len(alertIndexList) > 0 {
		// 缓存告警指纹， 用于完全消失时候的告警恢复
		s.data.GetCacher().Set(ctx, strategy.Index(), strings.Join(alertIndexList, ","), 0)
	}
	alarmInfo.Alerts = alerts
	return alarmInfo, nil
}

func (s *strategyRepositoryImpl) getFiringAlert(ctx context.Context, alert *bo.Alert) *bo.Alert {
	// 获取已存在的告警
	resolvedAlertStr, err := s.data.GetCacher().Get(ctx, alert.Labels.Index())
	if err != nil {
		return alert
	}

	var firingAlert bo.Alert
	if err := json.Unmarshal([]byte(resolvedAlertStr), &firingAlert); err != nil {
		return alert
	}
	alert.StartsAt = firingAlert.StartsAt
	// 更新最新的告警数据值
	if err := s.data.GetCacher().Set(ctx, alert.Labels.Index(), alert.String(), 0); err != nil {
		log.Warnw("method", "storage.put", "error", err)
		// TODO 存在争议， 不确定是否要把缓存失败的数据推出去
		// 如果不推， 会导致告警丢失，如果推送，会导致此事件没有告警恢复
		// 基于此原因， 选择推出去
	}
	return alert
}

func (s *strategyRepositoryImpl) getResolvedAlert(ctx context.Context, labels *vobj.Labels) (*bo.Alert, error) {
	// 获取已存在的告警
	resolvedAlertStr, err := s.data.GetCacher().Get(ctx, labels.Index())
	if err != nil {
		return nil, err
	}

	var resolvedAlert bo.Alert
	if err := json.Unmarshal([]byte(resolvedAlertStr), &resolvedAlert); err != nil {
		return nil, err
	}
	// 删除缓存
	s.data.GetAlertStorage().Remove(labels)
	resolvedAlert.EndsAt = types.NewTime(time.Now())
	return &resolvedAlert, nil
}

func isCompletelyMeet(pointValues []*datasource.Value, strategy *bo.Strategy) bool {
	values := types.SliceTo(pointValues, func(v *datasource.Value) float64 { return v.Value })
	judge := strategy.SustainType.Judge(strategy.Condition, strategy.Count, strategy.Threshold)
	return judge(values)
}

func findCommonKeys(maps ...*vobj.Labels) *vobj.Labels {
	if len(maps) == 0 {
		return nil
	}
	commonMap := maps[0]
	for _, m := range maps[1:] {
		for k, v := range m.Map() {
			commonVal, ok := commonMap.Map()[k]
			if !ok || commonVal != v {
				delete(commonMap.Map(), k)
			}
		}
	}
	return commonMap
}
