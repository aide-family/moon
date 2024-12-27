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

var _ IStrategy = (*StrategyHTTP)(nil)

type (
	// StrategyHTTP 端点响应时间、状态码策略
	StrategyHTTP struct {
		// 类型
		StrategyType vobj.StrategyType `json:"strategyType,omitempty"`
		// url 地址
		URL string `json:"url,omitempty"`
		// 状态码 200 404 500
		StatusCode string `json:"statusCode,omitempty"`
		// 状态码匹配模式
		StatusCodeCondition vobj.Condition `json:"statusCodeCondition,omitempty"`
		// 请求头
		Headers map[string]string `json:"headers,omitempty"`
		// 请求体
		Body string `json:"body,omitempty"`
		// 请求方式
		Method vobj.HTTPMethod `json:"method,omitempty"`
		// 响应时间阈值
		ResponseTime float64 `json:"responseTime,omitempty"`
		// 响应时间阈值条件
		ResponseTimeCondition vobj.Condition `json:"responseTimeCondition,omitempty"`
		// 策略标签
		Labels *vobj.Labels `json:"labels,omitempty"`
		// 策略注解
		Annotations *vobj.Annotations `json:"annotations,omitempty"`
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
		// 策略级别ID
		LevelID uint32 `json:"levelId,omitempty"`
		// 策略ID
		ID uint32 `json:"id,omitempty"`
	}
)

// String 将策略端点转换为字符串
func (e *StrategyHTTP) String() string {
	bs, _ := types.Marshal(e)
	return string(bs)
}

// Index 生成策略索引
func (e *StrategyHTTP) Index() string {
	if types.IsNil(e) {
		return "houyi:strategy:0:endpoint"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d:%s", e.TeamID, e.ID, e.LevelID, types.MD5(e.URL))
}

// Message 生成策略消息
func (e *StrategyHTTP) Message() *watch.Message {
	return watch.NewMessage(e, vobj.TopicStrategy)
}

// BuilderAlarmBaseInfo 生成告警基础信息
func (e *StrategyHTTP) BuilderAlarmBaseInfo() *Alarm {
	e.Labels.Append(vobj.StrategyID, strconv.FormatUint(uint64(e.ID), 10))
	e.Labels.Append(vobj.LevelID, strconv.FormatUint(uint64(e.LevelID), 10))
	e.Labels.Append(vobj.TeamID, strconv.FormatUint(uint64(e.TeamID), 10))
	e.Labels.Append(vobj.StrategyHTTPPath, e.URL)
	e.Labels.Append(vobj.StrategyHTTPMethod, e.Method.String())

	return &Alarm{
		Receiver:          strings.Join(types.SliceTo(e.ReceiverGroupIDs, func(id uint32) string { return fmt.Sprintf("team_%d_%d", e.TeamID, id) }), ","),
		Status:            vobj.AlertStatusFiring,
		Alerts:            nil,
		GroupLabels:       e.Labels,
		CommonLabels:      e.Labels,
		CommonAnnotations: e.Annotations,
		ExternalURL:       "",
		Version:           env.Version(),
		GroupKey:          "",
		TruncatedAlerts:   0,
	}
}

// GetTeamID 获取团队ID
func (e *StrategyHTTP) GetTeamID() uint32 {
	return e.TeamID
}

// GetStatus 获取策略状态
func (e *StrategyHTTP) GetStatus() vobj.Status {
	return e.Status
}

// GetReceiverGroupIDs 获取接收者组ID列表
func (e *StrategyHTTP) GetReceiverGroupIDs() []uint32 {
	return e.ReceiverGroupIDs
}

// GetLabelNotices 获取自定义接收者匹配对象
func (e *StrategyHTTP) GetLabelNotices() []*LabelNotices {
	return e.LabelNotices
}

// GetAnnotations 获取策略注解
func (e *StrategyHTTP) GetAnnotations() map[string]string {
	return e.Annotations.Map()
}

// GetInterval 获取执行频率
func (e *StrategyHTTP) GetInterval() *types.Duration {
	return types.NewDuration(durationpb.New(10 * time.Second))
}

// Eval 评估策略
func (e *StrategyHTTP) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	if !e.Status.IsEnable() {
		return nil, nil
	}
	return datasource.EndpointDuration(ctx, e.URL, e.Method, e.Headers, e.Body, 10*time.Second), nil
}

// IsCompletelyMeet 是否完全满足策略条件
func (e *StrategyHTTP) IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool) {
	if len(values) == 0 || !e.Status.IsEnable() {
		return nil, false
	}
	if len(values) != 2 {
		return nil, false
	}

	code := values[1].Value
	duration := values[0].Value
	extJSON := map[string]any{
		"code":     code,
		"duration": duration,
	}

	codeMatch := types.MatchStatusCodes(e.StatusCode, int(code))
	responseTimeMatch := e.ResponseTimeCondition.Judge(e.ResponseTime, duration)

	if e.StatusCodeCondition.IsEQ() && codeMatch {
		codeMatch = true
	}
	if e.StatusCodeCondition.IsNE() && !codeMatch {
		codeMatch = true
	}
	return extJSON, codeMatch && responseTimeMatch
}
