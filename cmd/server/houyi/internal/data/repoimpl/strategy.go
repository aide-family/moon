package repoimpl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/exp/maps"
)

// NewStrategyRepository 创建策略操作器
func NewStrategyRepository(data *data.Data) repository.Strategy {
	return &strategyRepositoryImpl{data: data}
}

type strategyRepositoryImpl struct {
	data *data.Data
}

// Save 保存策略
func (s *strategyRepositoryImpl) Save(_ context.Context, strategies []bo.IStrategy) error {
	queue := s.data.GetStrategyQueue()
	mqQueue := s.data.GetEventStrategyQueue()
	go func() {
		defer after.RecoverX()
		for _, strategyItem := range strategies {
			item, ok := strategyItem.(bo.IStrategyEvent)
			log.Debugw("strategyItem:", item)
			if ok {
				if err := mqQueue.Push(item.Message()); err != nil {
					log.Errorw("method", "mqQueue.push", "error", err)
				}
				continue
			}
			if err := queue.Push(strategyItem.Message()); err != nil {
				log.Errorw("method", "queue.push", "error", err)
			}
		}
	}()

	return nil
}

func (s *strategyRepositoryImpl) resolvedAlerts(ctx context.Context, strategy bo.IStrategy, alertKeys ...string) (alerts []*bo.Alert) {
	alertsStr, err := s.data.GetCacher().Get(ctx, strategy.Index())
	if err != nil || alertsStr == "" {
		return
	}
	firingKeys := strings.Split(alertsStr, ",")
	// 生成告警恢复事件
	if len(firingKeys) == 0 {
		return
	}
	firingKeyMap := types.ToMap(alertKeys, func(key string) string { return key })
	for _, existAlert := range firingKeys {
		if _, ok := firingKeyMap[existAlert]; ok {
			continue
		}
		resolvedAlert, err := getResolvedAlert(ctx, s.data, existAlert)
		if err != nil {
			continue
		}
		alerts = append(alerts, resolvedAlert)
	}
	return alerts
}

// Eval 评估策略 告警/恢复
func (s *strategyRepositoryImpl) Eval(ctx context.Context, strategy bo.IStrategy) (*bo.Alarm, error) {
	alarmInfo := strategy.BuilderAlarmBaseInfo()
	if !strategy.GetStatus().IsEnable() {
		alerts := s.resolvedAlerts(ctx, strategy)
		alarmInfo.Alerts = alerts
		alarmInfo.Status = vobj.AlertStatusResolved
		_ = s.data.GetCacher().Delete(ctx, strategy.Index())
		_ = s.data.GetCacher().Delete(ctx, alarmInfo.Index())
		return alarmInfo, nil
	}

	evalPoints, err := strategy.Eval(ctx)
	if err != nil {
		log.Warnw("method", "Eval", "error", err)
		return nil, err
	}

	var alerts []*bo.Alert
	receiverGroupIDsMap := types.ToMap(strategy.GetReceiverGroupIDs(), func(id uint32) string { return fmt.Sprintf("team_%d_%d", strategy.GetTeamID(), id) })
	for index, point := range evalPoints {
		labels, ok := index.(*vobj.Labels)
		if !ok {
			continue
		}
		extJSON, ok := strategy.IsCompletelyMeet(point.Values)
		if !ok {
			continue
		}

		// 判断labels里面key值是否满足告警
		for _, notice := range strategy.GetLabelNotices() {
			// 判断key是否存在
			if !labels.Match(notice.Key, notice.Value) {
				continue
			}
			// 加入到通知对象里面
			for _, receiverGroupID := range notice.ReceiverGroupIDs {
				receiverGroupIDStr := fmt.Sprintf("team_%d_%d", strategy.GetTeamID(), receiverGroupID)
				receiverGroupIDsMap[receiverGroupIDStr] = receiverGroupID
			}
		}

		valLength := len(point.Values)
		endPointValue := point.Values[valLength-1]

		labels.AppendMap(alarmInfo.CommonLabels.Map()).AppendMap(point.Labels)
		formatValue := map[string]any{
			"value":  endPointValue.Value,
			"time":   endPointValue.Timestamp,
			"labels": labels.Map(),
			"ext":    extJSON,
		}
		annotations := make(vobj.Annotations, len(strategy.GetAnnotations()))

		for key, annotation := range strategy.GetAnnotations() {
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
		alert = getFiringAlert(ctx, s.data, alert)
		alerts = append(alerts, alert)
		alarmInfo.CommonLabels = findCommonKeys([]*vobj.Labels{alarmInfo.CommonLabels, labels}...)
	}

	firingKeys := types.SliceToWithFilter(alerts, func(alert *bo.Alert) (string, bool) {
		return alert.Index(), alert.Status.IsFiring()
	})
	alarmInfo.Receiver = strings.Join(maps.Keys(receiverGroupIDsMap), ",")

	alerts = append(alerts, s.resolvedAlerts(ctx, strategy, firingKeys...)...)

	if len(alerts) == 0 {
		// 删除缓存
		_ = s.data.GetCacher().Delete(ctx, strategy.Index())
		_ = s.data.GetCacher().Delete(ctx, alarmInfo.Index())
		return alarmInfo, nil
	}
	if len(firingKeys) > 0 {
		// 缓存告警指纹， 用于完全消失时候的告警恢复
		if err := s.data.GetCacher().Set(ctx, strategy.Index(), strings.Join(firingKeys, ","), 0); err != nil {
			log.Warnw("method", "storage.put", "error", err)
		}
	} else {
		_ = s.data.GetCacher().Delete(ctx, strategy.Index())
	}
	alarmInfo.Alerts = alerts
	return alarmInfo, nil
}

func getFiringAlert(ctx context.Context, d *data.Data, alert *bo.Alert) *bo.Alert {
	// 获取已存在的告警
	firingKey, err := d.GetCacher().Get(ctx, alert.Index())
	if err == nil {
		firingAlert, err := bo.NewAlertWithAlertStrInfo(firingKey)
		if err != nil {
			log.Warnw("method", "bo.NewAlertWithAlertStrInfo", "error", err)
		} else {
			alert.StartsAt = firingAlert.StartsAt
		}
	}

	// 更新最新的告警数据值
	if err := d.GetCacher().Set(ctx, alert.Index(), alert.String(), 0); err != nil {
		log.Warnw("method", "storage.put", "error", err)
		// TODO 存在争议， 不确定是否要把缓存失败的数据推出去
		// 如果不推， 会导致告警丢失，如果推送，会导致此事件没有告警恢复
		// 基于此原因， 选择推出去
	}
	return alert
}

func getResolvedAlert(ctx context.Context, d *data.Data, uniqueKey string) (*bo.Alert, error) {
	// 获取存在的告警信息
	var resolvedAlert bo.Alert
	if err := d.GetCacher().GetObject(ctx, uniqueKey, &resolvedAlert); err != nil {
		return nil, err
	}

	// 获取存在的告警信息， 并且更新为告警恢复状态
	resolvedAlert.Status = vobj.AlertStatusResolved
	resolvedAlert.EndsAt = types.NewTimeByUnix(time.Now().Unix())
	// 删除缓存
	_ = d.GetCacher().Delete(ctx, uniqueKey)
	return &resolvedAlert, nil
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
