package bo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/houyi/mq"
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"

	"google.golang.org/protobuf/types/known/durationpb"
)

var _ IStrategyEvent = (*StrategyEvent)(nil)

// IStrategyEvent 事件策略统一接口
type IStrategyEvent interface {
	IStrategy

	// SetValue 设置数据
	SetValue(msg *mq.Msg) IStrategyEvent
	// GetDatasource 获取数据源
	GetDatasource() []*EventDatasource
	// GetTopic 获取主题
	GetTopic() string
}

// StrategyEvent 事件策略明细
type StrategyEvent struct {
	// 策略类型
	StrategyType vobj.StrategyType `json:"strategyType,omitempty"`
	// 团队ID
	TeamID uint32 `json:"teamId,omitempty"`
	// 接收者 （告警组ID列表）
	ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
	// 策略ID
	ID uint32 `json:"id,omitempty"`
	// 策略等级ID
	LevelID uint32 `json:"levelId,omitempty"`
	// 策略名称
	Alert string `json:"alert,omitempty"`
	// 策略语句
	Expr string `json:"expr,omitempty"`
	// 阈值
	Threshold string `json:"threshold,omitempty"`
	// 判断条件
	Condition vobj.EventCondition `json:"condition,omitempty"`
	// 数据类型
	DataType vobj.EventDataType `json:"dataType,omitempty"`
	// 数据 Key
	DataKey string `json:"dataKey,omitempty"`
	// 数据源
	Datasource []*EventDatasource `json:"datasource,omitempty"`
	// 策略状态
	Status vobj.Status `json:"status,omitempty"`
	// 策略标签
	Labels *label.Labels `json:"labels,omitempty"`
	// 策略注解
	Annotations *label.Annotations `json:"annotations,omitempty"`

	msg *mq.Msg
}

// GetTopic 获取主题
func (s *StrategyEvent) GetTopic() string {
	return s.Expr
}

// GetDatasource 获取数据源
func (s *StrategyEvent) GetDatasource() []*EventDatasource {
	if types.IsNil(s) {
		return nil
	}
	return s.Datasource
}

// SetValue 设置数据
func (s *StrategyEvent) SetValue(msg *mq.Msg) IStrategyEvent {
	if types.IsNil(msg) {
		return nil
	}
	s.msg = msg
	return s
}

// String 实现 fmt.Stringer 接口
func (s *StrategyEvent) String() string {
	if types.IsNil(s) {
		return "{}"
	}
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 实现 watch.Indexer 接口
func (s *StrategyEvent) Index() string {
	if types.IsNil(s) {
		return "houyi:event_strategy:0:0:0"
	}
	return types.TextJoin("houyi:event_strategy:", strconv.Itoa(int(s.TeamID)), ":", strconv.Itoa(int(s.ID)), ":", strconv.Itoa(int(s.LevelID)))
}

// Message 实现 watch.Message 接口
func (s *StrategyEvent) Message() *watch.Message {
	if types.IsNil(s) {
		return nil
	}
	return watch.NewMessage(s, vobj.TopicEventStrategy)
}

// BuilderAlarmBaseInfo 实现 IStrategy 接口
func (s *StrategyEvent) BuilderAlarmBaseInfo() *Alarm {
	s.Labels.Append(label.StrategyID, strconv.FormatUint(uint64(s.ID), 10))
	s.Labels.Append(label.LevelID, strconv.FormatUint(uint64(s.LevelID), 10))
	s.Labels.Append(label.TeamID, strconv.FormatUint(uint64(s.TeamID), 10))
	s.Labels.Append(label.StrategyEventExpr, s.Expr)

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

// GetTeamID 实现 IStrategy 接口
func (s *StrategyEvent) GetTeamID() uint32 {
	if types.IsNil(s) {
		return 0
	}
	return s.TeamID
}

// GetStatus 实现 IStrategy 接口
func (s *StrategyEvent) GetStatus() vobj.Status {
	if types.IsNil(s) {
		return vobj.StatusUnknown
	}
	return s.Status
}

// GetReceiverGroupIDs 实现 IStrategy 接口
func (s *StrategyEvent) GetReceiverGroupIDs() []uint32 {
	if types.IsNil(s) {
		return nil
	}
	return s.ReceiverGroupIDs
}

// GetLabelNotices 实现 IStrategy 接口
func (s *StrategyEvent) GetLabelNotices() []*LabelNotices {
	if types.IsNil(s) {
		return nil
	}
	return nil
}

// GetAnnotations 实现 IStrategy 接口
func (s *StrategyEvent) GetAnnotations() map[string]string {
	if types.IsNil(s) {
		return nil
	}
	return s.Annotations.Map()
}

// GetInterval 实现 IStrategy 接口
func (s *StrategyEvent) GetInterval() *types.Duration {
	return types.NewDuration(durationpb.New(60 * time.Second))
}

// Eval 实现 IStrategy 接口
func (s *StrategyEvent) Eval(_ context.Context) (map[watch.Indexer]*datasource.Point, error) {
	return map[watch.Indexer]*datasource.Point{
		s.Labels: {
			Labels: s.Labels.Map(),
			Values: []*datasource.Value{
				{
					Value:     s.isCompletelyMeet(),
					Timestamp: s.getEventTime().Unix(),
					Ext: map[string]any{
						label.StrategyEventInfo: string(s.msg.Data),
					},
				},
			},
		},
	}, nil
}

func (s *StrategyEvent) isCompletelyMeet() float64 {
	has := s.Condition.Judge(s.msg.Data, s.DataType, s.DataKey, s.Threshold)
	if has {
		return 1
	}
	return 0
}

// getEventTime 获取告警时间
func (s *StrategyEvent) getEventTime() *types.Time {
	if s.msg.Timestamp == nil {
		return types.NewTime(time.Now())
	}
	return s.msg.Timestamp
}

// IsCompletelyMeet 实现 IStrategy 接口
func (s *StrategyEvent) IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool) {
	if types.IsNil(s) {
		return nil, false
	}
	if len(values) == 0 || !s.Status.IsEnable() {
		return nil, false
	}
	if len(values) != 1 {
		return nil, false
	}
	value := values[0]
	return values[0].Ext, value.Value == 1
}
