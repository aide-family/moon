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

var _ IStrategy = (*StrategyEndpoint)(nil)

type (
	// StrategyEndpoint 端点响应时间、状态码策略
	StrategyEndpoint struct {
		// 类型
		Type vobj.StrategyType `json:"type,omitempty"`
		// url 地址
		Url string `json:"url,omitempty"`
		// 超时时间
		Timeout uint32 `json:"timeout,omitempty"`
		// 状态码 200 404 500
		StatusCode uint32 `json:"statusCode,omitempty"`
		// 请求头
		Headers map[string]string `json:"headers,omitempty"`
		// 请求体
		Body string `json:"body,omitempty"`
		// 请求方式
		Method vobj.HttpMethod `json:"method,omitempty"`
		// 相应时间阈值
		Threshold float64 `json:"threshold,omitempty"`
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
		Interval *types.Duration `json:"interval,omitempty"`
		// 策略级别ID
		LevelID uint32 `json:"levelId,omitempty"`
		// 策略ID
		ID uint32 `json:"id,omitempty"`
	}
)

func (e *StrategyEndpoint) String() string {
	bs, _ := types.Marshal(e)
	return string(bs)
}

func (e *StrategyEndpoint) Index() string {
	if types.IsNil(e) {
		return "houyi:strategy:0:endpoint"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d:%s", e.TeamID, e.ID, e.LevelID, types.MD5(e.Url))
}

func (e *StrategyEndpoint) Message() *watch.Message {
	return watch.NewMessage(e, vobj.TopicStrategy)
}

func (e *StrategyEndpoint) BuilderAlarmBaseInfo() *Alarm {
	e.Labels.Append(vobj.StrategyID, strconv.FormatUint(uint64(e.ID), 10))
	e.Labels.Append(vobj.LevelID, strconv.FormatUint(uint64(e.LevelID), 10))
	e.Labels.Append(vobj.TeamID, strconv.FormatUint(uint64(e.TeamID), 10))
	e.Labels.Append(vobj.StrategyHttpPath, e.Url)
	e.Labels.Append(vobj.StrategyHttpMethod, e.Method.String())

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

func (e *StrategyEndpoint) GetTeamID() uint32 {
	return e.TeamID
}

func (e *StrategyEndpoint) GetStatus() vobj.Status {
	return e.Status
}

func (e *StrategyEndpoint) GetReceiverGroupIDs() []uint32 {
	return e.ReceiverGroupIDs
}

func (e *StrategyEndpoint) GetLabelNotices() []*LabelNotices {
	return e.LabelNotices
}

func (e *StrategyEndpoint) GetAnnotations() map[string]string {
	return e.Annotations
}

func (e *StrategyEndpoint) GetInterval() *types.Duration {
	return e.Interval
}

func (e *StrategyEndpoint) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	if !e.Status.IsEnable() {
		return nil, nil
	}
	return datasource.EndpointDuration(ctx, e.Url, e.Method, e.Headers, e.Body, time.Duration(e.Timeout)), nil
}

func (e *StrategyEndpoint) IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool) {
	if len(values) == 0 || !e.Status.IsEnable() {
		return nil, false
	}
	if len(values) != 2 {
		return nil, false
	}

	code := values[1].Value
	duration := values[0].Value
	extJson := map[string]any{
		"code":     code,
		"duration": duration,
	}
	if e.StatusCode != 0 && (float64(e.StatusCode) == code || code == 0) {
		return extJson, true
	}
	if e.Threshold != 0 && (duration >= e.Threshold || duration == 0) {
		return extJson, true
	}

	return extJson, false
}
