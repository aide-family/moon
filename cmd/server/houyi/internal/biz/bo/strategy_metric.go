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
	"google.golang.org/protobuf/types/known/durationpb"
)

var _ IStrategy = (*StrategyMetric)(nil)

type (
	// IStrategy 策略接口
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

	// LabelNotices 自定义接收者匹配对象
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
		// 数据源
		Datasource []*Datasource `json:"datasource,omitempty"`
		// 策略状态
		Status vobj.Status `json:"status,omitempty"`
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

// GetInterval 获取执行频率
func (s *StrategyMetric) GetInterval() *types.Duration {
	return types.NewDuration(durationpb.New(time.Second * 10))
}

// BuilderAlarmBaseInfo 生成告警基础信息
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

// GetTeamID 获取团队ID
func (s *StrategyMetric) GetTeamID() uint32 {
	return s.TeamID
}

// GetStatus 获取策略状态
func (s *StrategyMetric) GetStatus() vobj.Status {
	return s.Status
}

// GetReceiverGroupIDs 获取接收者组ID列表
func (s *StrategyMetric) GetReceiverGroupIDs() []uint32 {
	return s.ReceiverGroupIDs
}

// GetLabelNotices 获取自定义接收者匹配对象
func (s *StrategyMetric) GetLabelNotices() []*LabelNotices {
	return s.LabelNotices
}

// GetAnnotations 获取策略注解
func (s *StrategyMetric) GetAnnotations() map[string]string {
	return s.Annotations.Map()
}

// Eval 评估策略
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

// getDatasourceCliList 获取数据源客户端列表
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

// String 将策略转换为字符串
func (s *StrategyMetric) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 策略唯一索引
func (s *StrategyMetric) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0:0:0"
	}
	return types.TextJoin("houyi:strategy:", strconv.Itoa(int(s.TeamID)), ":", strconv.Itoa(int(s.ID)), ":", strconv.Itoa(int(s.LevelID)))
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
