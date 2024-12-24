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
	"google.golang.org/protobuf/types/known/durationpb"
)

var _ IStrategy = (*StrategyDomain)(nil)

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
	Threshold float64 `json:"threshold,omitempty"`
	// 策略标签
	Labels *vobj.Labels `json:"labels,omitempty"`
	// 策略注解
	Annotations *vobj.Annotations `json:"annotations,omitempty"`
	// 域名
	Domain string `json:"domain,omitempty"`
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

// IsCompletelyMeet 判断策略是否完全满足条件
func (s *StrategyDomain) IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool) {
	if !s.Status.IsEnable() {
		return nil, false
	}
	for _, point := range values {
		// 域名证书检测、小于等于阈值都是满足条件的
		if s.Type.IsDomaincertificate() && point.Value <= s.Threshold {
			return nil, true
		}
		// 端口检测、等于阈值才是满足条件的 1开启， 0关闭
		if s.Type.IsDomainport() && point.Value == s.Threshold {
			return nil, true
		}
	}
	return nil, false
}

// BuilderAlarmBaseInfo 构建告警基础信息
func (s *StrategyDomain) BuilderAlarmBaseInfo() *Alarm {
	s.Labels.Append(vobj.StrategyID, strconv.FormatUint(uint64(s.ID), 10))
	s.Labels.Append(vobj.LevelID, strconv.FormatUint(uint64(s.LevelID), 10))
	s.Labels.Append(vobj.TeamID, strconv.FormatUint(uint64(s.TeamID), 10))
	s.Labels.Append(vobj.Domain, s.Domain)
	s.Labels.Append(vobj.DomainPort, strconv.FormatUint(uint64(s.Port), 10))

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

// Eval 策略评估
func (s *StrategyDomain) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	if !s.Status.IsEnable() {
		return nil, nil
	}
	if s.Type.IsDomainport() {
		return datasource.EndpointPortEval(ctx, s.Domain, s.Port, 10*time.Second)
	}
	return datasource.DomainEval(ctx, s.Domain, s.Port, 10*time.Second)
}

// GetTeamID 获取团队ID
func (s *StrategyDomain) GetTeamID() uint32 {
	return s.TeamID
}

// GetStatus 获取策略状态
func (s *StrategyDomain) GetStatus() vobj.Status {
	return s.Status
}

// GetReceiverGroupIDs 获取接收者组ID列表
func (s *StrategyDomain) GetReceiverGroupIDs() []uint32 {
	return s.ReceiverGroupIDs
}

// GetLabelNotices 获取自定义接收者匹配对象
func (s *StrategyDomain) GetLabelNotices() []*LabelNotices {
	return s.LabelNotices
}

// GetAnnotations 获取策略注解
func (s *StrategyDomain) GetAnnotations() map[string]string {
	return s.Annotations.Map()
}

// GetInterval 获取执行频率
func (s *StrategyDomain) GetInterval() *types.Duration {
	return types.NewDuration(durationpb.New(5 * time.Second))
}
