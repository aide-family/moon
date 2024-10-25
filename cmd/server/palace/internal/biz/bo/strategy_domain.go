package bo

import (
	"fmt"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*StrategyDomain)(nil)

// StrategyDomain 证书策略
type StrategyDomain struct {
	// 接收者 （告警组ID列表）
	ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
	// 自定义接收者匹配对象
	LabelNotices []*LabelNotices `json:"labelNotices,omitempty"`
	// 策略ID
	ID uint32 `json:"id,omitempty"`
	// 策略等级ID
	LevelID uint32 `json:"levelId,omitempty"`
	// 团队ID
	TeamID uint32 `json:"teamId,omitempty"`
	// 状态
	Status vobj.Status `json:"status,omitempty"`
	// 策略名称
	Alert string `json:"alert,omitempty"`
	// 阈值
	Threshold int64 `json:"threshold,omitempty"`
	// 策略标签
	Labels *vobj.Labels `json:"labels,omitempty"`
	// 策略注解
	Annotations vobj.Annotations `json:"annotations,omitempty"`
	// 域名
	Domain string `json:"domain,omitempty"`
	// 超时时间
	Timeout uint32 `json:"timeout,omitempty"`
	// 执行频率
	Interval int64 `json:"interval,omitempty"`
	// 端口
	Port uint32 `json:"port,omitempty"`
	// 类型
	Type vobj.StrategyType `json:"type,omitempty"`
}

// String 策略转字符串
func (s *StrategyDomain) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 策略唯一索引
func (s *StrategyDomain) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0:domain"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d:%s", s.TeamID, s.ID, s.LevelID, s.Domain)
}

// Message 策略转消息
func (s *StrategyDomain) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}
