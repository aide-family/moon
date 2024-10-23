package bo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ IStrategy = (*StrategyPing)(nil)

// StrategyPing ping 策略
type StrategyPing struct {
	// 类型
	Type vobj.StrategyType `json:"type,omitempty"`
	// 策略ID
	StrategyID uint32 `json:"strategy_id,omitempty"`
	// 团队ID
	TeamID uint32 `json:"teamId,omitempty"`
	// 状态
	Status vobj.Status `json:"status,omitempty"`
	// 策略名称
	Alert string `json:"alert,omitempty"`
	// 执行频率
	Interval *types.Duration `json:"interval,omitempty"`
	// 策略级别ID
	LevelID uint32 `json:"levelId,omitempty"`
	// 超时时间
	Timeout uint32 `json:"timeout,omitempty"`
	// 策略标签
	Labels *vobj.Labels `json:"labels,omitempty"`
	// 策略注解
	Annotations vobj.Annotations `json:"annotations,omitempty"`
	// 接收者 （告警组ID列表）
	ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
	// 域名或IP
	Address string `json:"address,omitempty"`

	// 总包数
	TotalPackets float64 `json:"totalPackets,omitempty"`
	// 成功包数
	SuccessPackets float64 `json:"successPackets,omitempty"`
	// 丢包率
	LossRate float64 `json:"lossRate,omitempty"`
	// 最小延迟
	MinDelay float64 `json:"minDelay,omitempty"`
	// 最大延迟
	MaxDelay float64 `json:"maxDelay,omitempty"`
	// 平均延迟
	AvgDelay float64 `json:"avgDelay,omitempty"`
	// 标准差
	StdDevDelay float64 `json:"stdDevDelay,omitempty"`
}

func (s *StrategyPing) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

func (s *StrategyPing) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0:ping"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d:%s", s.TeamID, s.StrategyID, s.LevelID, types.MD5(s.Address))
}

func (s *StrategyPing) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}

func (s *StrategyPing) BuilderAlarmBaseInfo() *Alarm {
	s.Labels.Append(vobj.StrategyID, strconv.FormatUint(uint64(s.StrategyID), 10))
	s.Labels.Append(vobj.LevelID, strconv.FormatUint(uint64(s.LevelID), 10))
	s.Labels.Append(vobj.TeamID, strconv.FormatUint(uint64(s.TeamID), 10))
	s.Labels.Append(vobj.Domain, s.Address)

	return &Alarm{
		Receiver:          strings.Join(types.SliceTo(s.ReceiverGroupIDs, func(id uint32) string { return fmt.Sprintf("team_%d_%d", s.TeamID, id) }), ","),
		Status:            vobj.AlertStatusFiring,
		Alerts:            nil,
		GroupLabels:       s.Labels,
		CommonLabels:      s.Labels,
		CommonAnnotations: s.Annotations,
		ExternalURL:       "",
		Version:           env.Version(),
		GroupKey:          "",
		TruncatedAlerts:   0,
	}
}

func (s *StrategyPing) GetTeamID() uint32 {
	return s.TeamID
}

func (s *StrategyPing) GetStatus() vobj.Status {
	return s.Status
}

func (s *StrategyPing) GetReceiverGroupIDs() []uint32 {
	return s.ReceiverGroupIDs
}

func (s *StrategyPing) GetLabelNotices() []*LabelNotices {
	return make([]*LabelNotices, 0)
}

func (s *StrategyPing) GetAnnotations() map[string]string {
	return s.Annotations
}

func (s *StrategyPing) GetInterval() *types.Duration {
	return s.Interval
}

func (s *StrategyPing) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	return datasource.EndpointPing(ctx, s.Address, time.Duration(s.Timeout))
}

// IsCompletelyMeet checks if the ping metrics meet the strategy criteria
func (s *StrategyPing) IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool) {
	const expectedMetricsCount = 7

	// Early validation
	if len(values) != expectedMetricsCount || !s.Status.IsEnable() {
		return nil, false
	}

	// Create metrics map for cleaner data access
	metrics := map[string]float64{
		"totalPackets":   values[0].Value,
		"successPackets": values[1].Value,
		"lossRate":       values[2].Value,
		"minDelay":       values[3].Value,
		"maxDelay":       values[4].Value,
		"avgDelay":       values[5].Value,
		"stdDevDelay":    values[6].Value,
	}

	// Prepare extended info
	extJson := make(map[string]any, len(metrics))
	for k, v := range metrics {
		extJson[k] = v
	}

	// Define threshold checks
	thresholds := []struct {
		configValue float64 // Strategy configuration value
		metricValue float64 // Actual metric value
		condition   string  // Description for debugging
		comparison  func(configVal, metricVal float64) bool
	}{
		{s.TotalPackets, metrics["totalPackets"], "total packets", func(c, m float64) bool { return c > 0 && c > m }},
		{s.LossRate, metrics["lossRate"], "loss rate", func(c, m float64) bool { return c > 0 && c > m }},
		{s.MinDelay, metrics["minDelay"], "min delay", func(c, m float64) bool { return c > 0 && c > m }},
		{s.MaxDelay, metrics["maxDelay"], "max delay", func(c, m float64) bool { return c > 0 && c > m }},
		{s.AvgDelay, metrics["avgDelay"], "average delay", func(c, m float64) bool { return c > 0 && c > m }},
		{s.StdDevDelay, metrics["stdDevDelay"], "standard deviation delay", func(c, m float64) bool { return c > 0 && c > m }},
	}

	// Check each threshold
	for _, check := range thresholds {
		if check.comparison(check.configValue, check.metricValue) {
			return extJson, true
		}
	}

	return extJson, false
}
