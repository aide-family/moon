package repoimpl

import (
	"context"
	"fmt"

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

// Eval 评估策略
func (s *strategyRepositoryImpl) Eval(ctx context.Context, strategy *bo.Strategy) (*bo.Alarm, error) {
	datasourceList := strategy.Datasource
	if len(datasourceList) == 0 {
		return nil, merr.ErrorNotification("datasource is empty")
	}
	strategy.Labels.Append(vobj.StrategyID, fmt.Sprintf("%d", strategy.ID))
	category := datasourceList[0].Category
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
	datasourceCliList := make([]datasource.Datasource, 0, len(datasourceList))
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
		datasourceCliList = append(datasourceCliList, datasource.NewDatasource(cfg))
	}

	var alerts []*bo.Alert
	for _, cli := range datasourceCliList {
		// TODO async eval
		step := cli.Step()
		if strategy.Step > 0 {
			step = strategy.Step
		}
		evalPoints, err := cli.Eval(ctx, strategy.Expr, step)
		if err != nil {
			log.Warnw("method", "Eval", "error", err)
		}
		if len(evalPoints) == 0 {
			return nil, merr.ErrorNotification("no data")
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
				GeneratorURL: "",
				Fingerprint:  "",
				Value:        endPointValue.Value,
			}
			alerts = append(alerts, alert)
			alarmInfo.CommonLabels = findCommonKeys([]*vobj.Labels{alarmInfo.CommonLabels, labels}...)
		}
	}
	if len(alerts) == 0 {
		return nil, merr.ErrorNotification("no data")
	}
	alarmInfo.Alerts = alerts
	return &alarmInfo, nil
}

func isCompletelyMeet(pointValues []*datasource.Value, strategy *bo.Strategy) bool {
	switch strategy.Condition {
	case vobj.ConditionEQ: // equal
		for _, point := range pointValues {
			if point.Value != strategy.Threshold {
				return false
			}
		}
		return true
	case vobj.ConditionNE: // not equal
		for _, point := range pointValues {
			if point.Value == strategy.Threshold {
				return false
			}
		}
		return true
	case vobj.ConditionGT: // greater than
		for _, point := range pointValues {
			if point.Value <= strategy.Threshold {
				return false
			}
		}
		return true
	case vobj.ConditionLT: // less than
		for _, point := range pointValues {
			if point.Value >= strategy.Threshold {
				return false
			}
		}
		return true
	case vobj.ConditionGTE:
		for _, point := range pointValues {
			if point.Value < strategy.Threshold {
				return false
			}
		}
		return true
	case vobj.ConditionLTE:
		for _, point := range pointValues {
			if point.Value > strategy.Threshold {
				return false
			}
		}
		return true
	default:
		return false
	}
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
