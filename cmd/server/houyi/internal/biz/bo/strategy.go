package bo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-kratos/kratos/v2/log"
)

var _ IStrategy = (*Strategy)(nil)
var _ IStrategy = (*DomainStrategy)(nil)

type (
	IStrategy interface {
		watch.Indexer
		Message() *watch.Message
		BuilderAlarmBaseInfo() *Alarm
		GetTeamID() uint32
		GetStatus() vobj.Status
		GetReceiverGroupIDs() []uint32
		GetLabelNotices() []*LabelNotices
		GetAnnotations() map[string]string
		GetInterval() *types.Duration
		Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error)
		IsCompletelyMeet(values []*datasource.Value) bool
	}

	LabelNotices struct {
		// label key
		Key string `json:"key,omitempty"`
		// label value 支持正则
		Value string `json:"value,omitempty"`
		// 接收者 （告警组ID列表）
		ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
	}

	// Strategy 策略明细
	Strategy struct {
		// 接收者 （告警组ID列表）
		ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
		// 自定义接收者匹配对象
		LabelNotices []*LabelNotices `json:"labelNotices,omitempty"`
		// 策略ID
		ID uint32 `json:"id,omitempty"`
		// 策略等级ID
		LevelID uint32 `json:"levelId,omitempty"`
		// 策略名称
		Alert string `json:"alert,omitempty"`
		// 策略语句
		Expr string `json:"expr,omitempty"`
		// 策略持续时间
		For *types.Duration `json:"for,omitempty"`
		// 持续次数
		Count uint32 `json:"count,omitempty"`
		// 持续的类型
		SustainType vobj.Sustain `json:"sustainType,omitempty"`
		// 多数据源持续类型
		MultiDatasourceSustainType vobj.MultiDatasourceSustain `json:"multiDatasourceSustainType,omitempty"`
		// 策略标签
		Labels *vobj.Labels `json:"labels,omitempty"`
		// 策略注解
		Annotations vobj.Annotations `json:"annotations,omitempty"`
		// 执行频率
		Interval *types.Duration `json:"interval,omitempty"`
		// 数据源
		Datasource []*Datasource `json:"datasource,omitempty"`
		// 策略状态
		Status vobj.Status `json:"status,omitempty"`
		// 策略采样率
		Step uint32 `json:"step,omitempty"`
		// 判断条件
		Condition vobj.Condition `json:"condition,omitempty"`
		// 阈值
		Threshold float64 `json:"threshold,omitempty"`
		// 团队ID
		TeamID uint32 `json:"teamId,omitempty"`
	}

	// Datasource 数据源明细
	Datasource struct {
		// 数据源类型
		Category vobj.DatasourceType `json:"category,omitempty"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storage_type,omitempty"`
		// 数据源配置 json
		Config map[string]string `json:"config,omitempty"`
		// 数据源地址
		Endpoint string `json:"endpoint,omitempty"`
		// 数据源ID
		ID uint32 `json:"id,omitempty"`
	}

	// DomainStrategy 证书策略
	DomainStrategy struct {
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
		Annotations vobj.Annotations `json:"annotations,omitempty"`
		// 域名
		Domain string `json:"domain,omitempty"`
		// 超时时间
		Timeout uint32 `json:"timeout,omitempty"`
		// 执行频率
		Interval *types.Duration `json:"interval,omitempty"`
		// 端口
		Port uint32 `json:"port,omitempty"`
		// 类型
		Type vobj.StrategyType `json:"type,omitempty"`
	}
)

func (s *Strategy) GetInterval() *types.Duration {
	return s.Interval
}

func (s *Strategy) BuilderAlarmBaseInfo() *Alarm {
	s.Labels.Append(vobj.StrategyID, strconv.FormatUint(uint64(s.ID), 10))
	s.Labels.Append(vobj.LevelID, strconv.FormatUint(uint64(s.LevelID), 10))
	s.Labels.Append(vobj.TeamID, strconv.FormatUint(uint64(s.TeamID), 10))

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

func (s *Strategy) GetTeamID() uint32 {
	return s.TeamID
}

func (s *Strategy) GetStatus() vobj.Status {
	return s.Status
}

func (s *Strategy) GetReceiverGroupIDs() []uint32 {
	return s.ReceiverGroupIDs
}

func (s *Strategy) GetLabelNotices() []*LabelNotices {
	return s.LabelNotices
}

func (s *Strategy) GetAnnotations() map[string]string {
	return s.Annotations
}

func (s *Strategy) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	datasourceCliList, err := s.getDatasourceCliList()
	if err != nil {
		return nil, err
	}
	return datasource.MetricEval(datasourceCliList...)(ctx, s.Expr, s.For)
}

func (s *Strategy) getDatasourceCliList() ([]datasource.MetricDatasource, error) {
	datasourceList := s.Datasource
	datasourceCliList := make([]datasource.MetricDatasource, 0, len(datasourceList))
	category := datasourceList[0].Category
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
			Id:          datasourceItem.ID,
		}
		newDatasource, err := datasource.NewDatasource(cfg).Metric()
		if err != nil {
			log.Warnw("method", "NewDatasource", "error", err)
			continue
		}
		datasourceCliList = append(datasourceCliList, newDatasource)
	}
	if len(datasourceCliList) == 0 {
		return nil, merr.ErrorNotification("datasource is empty")
	}
	return datasourceCliList, nil
}

func (s *Strategy) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 策略唯一索引
func (s *Strategy) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d", s.TeamID, s.ID, s.LevelID)
}

// Message 策略转消息
func (s *Strategy) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}

// IsCompletelyMeet 判断策略是否完全满足条件
func (s *Strategy) IsCompletelyMeet(values []*datasource.Value) bool {
	points := types.SliceTo(values, func(v *datasource.Value) float64 { return v.Value })
	judge := s.SustainType.Judge(s.Condition, s.Count, s.Threshold)
	return judge(points)
}

// String 策略转字符串
func (s *DomainStrategy) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 策略唯一索引
func (s *DomainStrategy) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0:domain"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d:%s", s.TeamID, s.ID, s.LevelID, s.Domain)
}

// Message 策略转消息
func (s *DomainStrategy) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}

// IsCompletelyMeet 判断策略是否完全满足条件
func (s *DomainStrategy) IsCompletelyMeet(values []*datasource.Value) bool {
	if !s.Status.IsEnable() {
		return false
	}
	for _, point := range values {
		// 域名证书检测、小于等于阈值都是满足条件的
		if s.Type.IsDomaincertificate() && point.Value <= s.Threshold {
			return true
		}
		// 端口检测、等于阈值才是满足条件的 1开启， 0关闭
		if s.Type.IsDomainport() && point.Value == s.Threshold {
			return true
		}
	}
	return false
}

// BuilderAlarmBaseInfo 构建告警基础信息
func (s *DomainStrategy) BuilderAlarmBaseInfo() *Alarm {
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
func (s *DomainStrategy) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	if s.Type.IsDomainport() {
		return datasource.EndpointPortEval(ctx, s.Domain, s.Port, time.Duration(s.Timeout))
	}
	return datasource.DomainEval(ctx, s.Domain, s.Port, time.Duration(s.Timeout))
}

func (s *DomainStrategy) GetTeamID() uint32 {
	return s.TeamID
}

func (s *DomainStrategy) GetStatus() vobj.Status {
	return s.Status
}

func (s *DomainStrategy) GetReceiverGroupIDs() []uint32 {
	return s.ReceiverGroupIDs
}

func (s *DomainStrategy) GetLabelNotices() []*LabelNotices {
	return s.LabelNotices
}

func (s *DomainStrategy) GetAnnotations() map[string]string {
	return s.Annotations
}

func (s *DomainStrategy) GetInterval() *types.Duration {
	return s.Interval
}
