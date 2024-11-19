package bo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-kratos/kratos/v2/log"
)

var _ IStrategy = (*StrategyMetric)(nil)

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
		IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool)
	}

	LabelNotices struct {
		// label key
		Key string `json:"key,omitempty"`
		// label value 支持正则
		Value string `json:"value,omitempty"`
		// 接收者 （告警组ID列表）
		ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
	}

	// StrategyMetric 策略明细
	StrategyMetric struct {
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
		Annotations *vobj.Annotations `json:"annotations,omitempty"`
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
)

func (s *StrategyMetric) GetInterval() *types.Duration {
	return s.Interval
}

func (s *StrategyMetric) BuilderAlarmBaseInfo() *Alarm {
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

func (s *StrategyMetric) GetTeamID() uint32 {
	return s.TeamID
}

func (s *StrategyMetric) GetStatus() vobj.Status {
	return s.Status
}

func (s *StrategyMetric) GetReceiverGroupIDs() []uint32 {
	return s.ReceiverGroupIDs
}

func (s *StrategyMetric) GetLabelNotices() []*LabelNotices {
	return s.LabelNotices
}

func (s *StrategyMetric) GetAnnotations() map[string]string {
	return s.Annotations.Map()
}

func (s *StrategyMetric) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	if !s.Status.IsEnable() {
		return nil, nil
	}
	datasourceCliList, err := s.getDatasourceCliList()
	if err != nil {
		return nil, err
	}
	return datasource.MetricEval(datasourceCliList...)(ctx, s.Expr, s.For)
}

func (s *StrategyMetric) getDatasourceCliList() ([]datasource.MetricDatasource, error) {
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

func (s *StrategyMetric) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 策略唯一索引
func (s *StrategyMetric) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d", s.TeamID, s.ID, s.LevelID)
}

// Message 策略转消息
func (s *StrategyMetric) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}

// IsCompletelyMeet 判断策略是否完全满足条件
func (s *StrategyMetric) IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool) {
	points := types.SliceTo(values, func(v *datasource.Value) float64 { return v.Value })
	judge := s.SustainType.Judge(s.Condition, s.Count, s.Threshold)
	return nil, judge(points)
}
