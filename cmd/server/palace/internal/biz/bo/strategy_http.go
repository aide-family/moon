package bo

import (
	"fmt"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*StrategyEndpoint)(nil)

// StrategyEndpoint 端点响应时间、状态码策略
type StrategyEndpoint struct {
	// 类型
	Type vobj.StrategyType `json:"type,omitempty"`
	// url 地址
	URL string `json:"url,omitempty"`
	// 超时时间
	Timeout uint32 `json:"timeout,omitempty"`
	// 状态码 200 404 500
	StatusCode uint32 `json:"statusCode,omitempty"`
	// 请求头
	Headers map[string]string `json:"headers,omitempty"`
	// 请求体
	Body string `json:"body,omitempty"`
	// 请求方式
	Method vobj.HTTPMethod `json:"method,omitempty"`
	// 相应时间阈值
	Threshold int64 `json:"threshold,omitempty"`
	// 策略标签
	Labels *vobj.Labels `json:"labels,omitempty"`
	// 策略注解
	Annotations vobj.Annotations `json:"annotations,omitempty"`
	// 接收者 （告警组ID列表）
	ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
	// 自定义接收者匹配对象
	LabelNotices []*LabelNotices `json:"labelNotices,omitempty"`
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
	// 策略ID
	ID uint32 `json:"id,omitempty"`
}

// String 字符串
func (e *StrategyEndpoint) String() string {
	bs, _ := types.Marshal(e)
	return string(bs)
}

// Index 索引
func (e *StrategyEndpoint) Index() string {
	if types.IsNil(e) {
		return "houyi:strategy:0:endpoint"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d:%s", e.TeamID, e.ID, e.LevelID, types.MD5(e.URL))
}

// Message 消息
func (e *StrategyEndpoint) Message() *watch.Message {
	return watch.NewMessage(e, vobj.TopicStrategy)
}
