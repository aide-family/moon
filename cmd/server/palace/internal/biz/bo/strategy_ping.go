package bo

import (
	"fmt"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*StrategyPing)(nil)

// StrategyPing ping 策略
type StrategyPing struct {
	// 类型
	Type vobj.StrategyType `json:"type,omitempty"`
	// 策略ID
	ID uint32 `json:"strategy_id,omitempty"`
	// 团队ID
	TeamID uint32 `json:"teamId,omitempty"`
	// 状态
	Status vobj.Status `json:"status,omitempty"`
	// 策略名称
	Alert string `json:"alert,omitempty"`
	// 执行频率
	Interval int64 `json:"interval,omitempty"`
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
	TotalPackets int64 `json:"totalPackets,omitempty"`
	// 成功包数
	SuccessPackets int64 `json:"successPackets,omitempty"`
	// 丢包率
	LossRate float64 `json:"lossRate,omitempty"`
	// 最小延迟
	MinDelay int64 `json:"minDelay,omitempty"`
	// 最大延迟
	MaxDelay int64 `json:"maxDelay,omitempty"`
	// 平均延迟
	AvgDelay int64 `json:"avgDelay,omitempty"`
	// 标准差
	StdDevDelay int64 `json:"stdDevDelay,omitempty"`
}

func (s *StrategyPing) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

func (s *StrategyPing) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0:ping"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d:%s", s.TeamID, s.ID, s.LevelID, types.MD5(s.Address))
}

func (s *StrategyPing) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}
