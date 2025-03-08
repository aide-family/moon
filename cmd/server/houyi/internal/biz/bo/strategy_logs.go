package bo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/houyi/logs"
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/durationpb"
)

var _ watch.Indexer = (*LogDatasource)(nil)

var _ IStrategy = (*StrategyLogs)(nil)

type (
	StrategyLogs struct {
		// 类型
		Type vobj.StrategyType `json:"type,omitempty"`
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
		// 策略状态
		Status vobj.Status `json:"status,omitempty"`
		// 策略标签
		Labels *label.Labels `json:"labels,omitempty"`
		// 策略注解
		Annotations *label.Annotations      `json:"annotations,omitempty"`
		ResLog      *datasource.LogResponse `json:"resLog,omitempty"`
		// 自定义接收者匹配对象
		LabelNotices []*LabelNotices `json:"labelNotices,omitempty"`
		// 数据源
		Datasource []*LogDatasource `json:"datasource,omitempty"`
		// 持续次数
		Count uint32 `json:"count,omitempty"`
		// 策略持续时间
		For *types.Duration `json:"for,omitempty"`
	}
	LogDatasource struct {
		TeamID uint32         `json:"team_id"`
		ID     uint32         `json:"id"`
		Status vobj.Status    `json:"status"`
		Conf   *conf.LogQuery `json:"conf"`
	}
)

func (s *StrategyLogs) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

func (s *StrategyLogs) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0:logs:0"
	}
	return types.TextJoin("houyi:strategy:", strconv.Itoa(int(s.TeamID)), ":", strconv.Itoa(int(s.ID)), ":", strconv.Itoa(int(s.LevelID)))
}

func (s *StrategyLogs) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}

func (s *StrategyLogs) BuilderAlarmBaseInfo() *Alarm {
	s.Labels.Append(label.StrategyID, strconv.FormatUint(uint64(s.ID), 10))
	s.Labels.Append(label.LevelID, strconv.FormatUint(uint64(s.LevelID), 10))
	s.Labels.Append(label.TeamID, strconv.FormatUint(uint64(s.TeamID), 10))

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
func (s *StrategyLogs) GetTeamID() uint32 {
	return s.TeamID
}

// GetStatus 获取策略状态
func (s *StrategyLogs) GetStatus() vobj.Status {
	return s.Status
}

// GetReceiverGroupIDs 获取接收者组ID列表
func (s *StrategyLogs) GetReceiverGroupIDs() []uint32 {
	return s.ReceiverGroupIDs
}

// GetLabelNotices 获取自定义接收者匹配对象
func (s *StrategyLogs) GetLabelNotices() []*LabelNotices {
	return s.LabelNotices
}

// GetAnnotations 获取策略注解
func (s *StrategyLogs) GetAnnotations() map[string]string {
	return s.Annotations.Map()
}

// GetInterval 获取执行频率
func (s *StrategyLogs) GetInterval() *types.Duration {
	return types.NewDuration(durationpb.New(5 * time.Second))
}

func (s *StrategyLogs) getDatasourceCliList() ([]datasource.LogDatasource, error) {
	datasourceList := s.Datasource
	datasourceCliList := make([]datasource.LogDatasource, 0, len(datasourceList))
	for _, datasourceItem := range datasourceList {
		logQuery, err := logs.NewLogQuery(datasourceItem.Conf)
		if err != nil {
			log.Warnw("strategy logs datasource query error", datasourceItem.Conf, err)
			continue
		}
		datasourceCliList = append(datasourceCliList, logQuery)
	}
	return datasourceCliList, nil
}

func (s *StrategyLogs) Eval(ctx context.Context) (map[watch.Indexer]*datasource.Point, error) {
	if !s.Status.IsEnable() {
		return nil, nil
	}

	datasourceCliList, err := s.getDatasourceCliList()
	if err != nil {
		return nil, err
	}
	points := make(map[watch.Indexer]*datasource.Point)
	for _, item := range datasourceCliList {
		endAt := time.Now()
		startAt := types.NewTime(endAt.Add(-s.For.Duration.AsDuration()))
		queryRes, err := item.QueryLogs(ctx, s.Expr, startAt.Unix(), endAt.Unix())
		if err != nil {
			log.Warnw("strategy logs datasource query error", "expr", s.Expr, "startAt", startAt, "endAt", endAt, err)
			continue
		}
		datasourceUrl := queryRes.DatasourceUrl
		queryValue := queryRes.Values
		if types.IsNil(queryValue) {
			continue
		}
		labels := label.NewLabels(map[string]string{
			label.DatasourceID: datasourceUrl,
		})
		points[labels] = &datasource.Point{
			Labels: labels.Map(),
			Values: []*datasource.Value{
				{
					Value: float64(len(queryValue)),
					Ext: map[string]any{
						label.StrategyLogInfo: queryRes.Values,
					},
					Timestamp: queryRes.Timestamp,
				},
			},
		}
	}

	return points, nil
}

// IsCompletelyMeet 是否满足条件
func (s *StrategyLogs) IsCompletelyMeet(values []*datasource.Value) (map[string]any, bool) {
	if types.IsNil(s) {
		return nil, false
	}
	if len(values) == 0 || !s.Status.IsEnable() {
		return nil, false
	}

	return values[0].Ext, len(values) >= int(s.Count)
}

func (l *LogDatasource) String() string {
	bs, _ := types.Marshal(l)
	return string(bs)
}

func (l *LogDatasource) Index() string {
	return types.TextJoin(strconv.Itoa(int(l.TeamID)), ":", strconv.Itoa(int(l.ID)))
}
