package repoimpl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/merr"
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

func (s *strategyRepositoryImpl) getDatasourceCliList(strategy *bo.Strategy) ([]datasource.MetricDatasource, error) {
	datasourceList := strategy.Datasource
	datasourceCliList := make([]datasource.MetricDatasource, 0, len(datasourceList))
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
			Id:          datasourceItem.ID,
		}
		newDatasource, err := datasource.NewDatasource(cfg).Metric()
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
	strategy.Labels.Append(vobj.LevelID, fmt.Sprintf("%d", strategy.LevelID))
	strategy.Labels.Append(vobj.TeamID, fmt.Sprintf("%d", strategy.TeamID))

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
	alarmInfo := builderAlarmBaseInfo(strategy)
	var alerts []*bo.Alert
	// 获取存在的告警标识列表
	alertsStr, _ := s.data.GetCacher().Get(ctx, strategy.Index())
	// 移除策略， 直接生成告警恢复事件
	if !strategy.Status.IsEnable() {
		existAlerts := strings.Split(alertsStr, ",")
		if len(existAlerts) == 0 {
			return alarmInfo, nil
		}
		for _, existAlert := range existAlerts {
			getResolvedAlert, err := s.getResolvedAlert(ctx, existAlert)
			if err != nil {
				log.Warnw("method", "NewAlertWithAlertStrInfo", "error", err)
				continue
			}
			alerts = append(alerts, getResolvedAlert)
		}
		alarmInfo.Alerts = alerts
		alarmInfo.Status = vobj.AlertStatusResolved
		s.data.GetCacher().Delete(ctx, strategy.Index())
		return alarmInfo, nil
	}
	datasourceCliList, err := s.getDatasourceCliList(strategy)
	if err != nil {
		return nil, err
	}
	evalPoints, err := datasource.MetricEval(datasourceCliList...)(ctx, strategy.Expr, strategy.For)
	if err != nil {
		log.Warnw("method", "Eval", "error", err)
		return nil, err
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

		labels.AppendMap(alarmInfo.CommonLabels.Map()).AppendMap(point.Labels)
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
			Labels:       labels,      // 合并label
			Annotations:  annotations, // 填充
			StartsAt:     types.NewTimeByUnix(endPointValue.Timestamp),
			EndsAt:       nil,
			GeneratorURL: "", // TODO 生成事件图表链接
			Fingerprint:  "", // TODO 指纹生成逻辑
			Value:        endPointValue.Value,
		}
		alert = s.getFiringAlert(ctx, alert)
		alerts = append(alerts, alert)
		alarmInfo.CommonLabels = findCommonKeys([]*vobj.Labels{alarmInfo.CommonLabels, labels}...)
	}

	alertIndexList := types.SliceToWithFilter(alerts, func(alert *bo.Alert) (string, bool) {
		return alert.Index(), alert.Status.IsFiring()
	})

	if !types.TextIsNull(alertsStr) {
		existAlerts := strings.Split(alertsStr, ",")
		alertIndexListMap := make(map[string]struct{}, len(alerts))
		for _, alertItem := range alerts {
			alertIndexListMap[alertItem.Index()] = struct{}{}
		}
		for _, existAlert := range existAlerts {
			if _, ok := alertIndexListMap[existAlert]; !ok {
				getResolvedAlert, err := s.getResolvedAlert(ctx, existAlert)
				if err != nil {
					log.Warnw("method", "NewAlertWithAlertStrInfo", "error", err)
					continue
				}
				alerts = append(alerts, getResolvedAlert)
			}
		}
	}

	if len(alerts) == 0 {
		// 删除缓存
		s.data.GetCacher().Delete(ctx, strategy.Index())
		s.data.GetCacher().Delete(ctx, alarmInfo.Index())
		return alarmInfo, nil
	}
	if len(alertIndexList) > 0 {
		// 缓存告警指纹， 用于完全消失时候的告警恢复
		if err := s.data.GetCacher().Set(ctx, strategy.Index(), strings.Join(alertIndexList, ","), 0); err != nil {
			log.Warnw("method", "storage.put", "error", err)
		}
	} else {
		s.data.GetCacher().Delete(ctx, strategy.Index())
	}
	alarmInfo.Alerts = alerts
	return alarmInfo, nil
}

func (s *strategyRepositoryImpl) getFiringAlert(ctx context.Context, alert *bo.Alert) *bo.Alert {
	// 获取已存在的告警
	resolvedAlertStr, err := s.data.GetCacher().Get(ctx, alert.Index())
	if err != nil {
		log.Warnw("method", "storage.get", "error", err)
	} else {
		firingAlert, err := bo.NewAlertWithAlertStrInfo(resolvedAlertStr)
		if err != nil {
			log.Warnw("method", "bo.NewAlertWithAlertStrInfo", "error", err)
		} else {
			alert.StartsAt = firingAlert.StartsAt
		}
	}

	// 更新最新的告警数据值
	if err := s.data.GetCacher().Set(ctx, alert.Index(), alert.String(), 0); err != nil {
		log.Warnw("method", "storage.put", "error", err)
		// TODO 存在争议， 不确定是否要把缓存失败的数据推出去
		// 如果不推， 会导致告警丢失，如果推送，会导致此事件没有告警恢复
		// 基于此原因， 选择推出去
	}
	return alert
}

func (s *strategyRepositoryImpl) getResolvedAlert(ctx context.Context, uniqueKey string) (*bo.Alert, error) {
	// 获取存在的告警信息
	existAlertStr, err := s.data.GetCacher().Get(ctx, uniqueKey)
	if err != nil {
		log.Warnw("method", "storage.get", "error", err)
		return nil, err
	}

	resolvedAlert, err := bo.NewAlertWithAlertStrInfo(existAlertStr)
	if err != nil {
		log.Warnw("method", "NewAlertWithAlertStrInfo", "error", err)
		return nil, err
	}
	// 获取存在的告警信息， 并且更新为告警恢复状态
	resolvedAlert.Status = vobj.AlertStatusResolved
	resolvedAlert.EndsAt = types.NewTimeByUnix(time.Now().Unix())
	// 删除缓存
	s.data.GetCacher().Delete(ctx, uniqueKey)
	return resolvedAlert, nil
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
