package bo

import (
	"fmt"

	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

type (
	// CreateStrategyParams 创建策略请求参数
	CreateStrategyParams struct {
		// 策略组ID
		GroupID uint32 `json:"group_id"`
		// 策略模板id
		TemplateID uint32 `json:"template_id"`
		// 备注
		Remark string `json:"remark"`
		// 状态
		Status vobj.Status `json:"status"`
		// 数据源id
		DatasourceIDs []uint32 `json:"datasource_ids"`
		// 模板来源
		TemplateSource vobj.StrategyTemplateSource `json:"source_type"`
		// 策略名称
		Name   string `json:"name"`
		TeamID uint32 `json:"teamID"`
		// 标签
		Labels *vobj.Labels `json:"labels"`
		// 注解
		Annotations *vobj.Annotations `json:"annotations"`
		// 告警表达式
		Expr string `json:"expr"`
		// 策略类型
		CategoriesIds []uint32 `json:"categoriesIds"`
		// 告警组
		AlarmGroupIds []uint32 `json:"alarmGroupIds"`
		// 策略类型
		StrategyType vobj.StrategyType `json:"strategyType"`
		// Metric策略等级
		MetricLevels []*CreateStrategyMetricLevel `json:"metricLevels"`
		// 事件策略等级
		EventLevels []*CreateStrategyEventLevel `json:"mqLevels"`
		// 域名证书等级
		DomainLevels []*CreateStrategyDomainLevel `json:"domainLevels"`
		// 端口证书等级
		PortLevels []*CreateStrategyPortLevel `json:"portLevels"`
		// HTTP策略等级
		HTTPLevels []*CreateStrategyHTTPLevel `json:"httpLevels"`
	}

	// UpdateStrategyParams 更新策略请求参数
	UpdateStrategyParams struct {
		ID          uint32 `json:"id"`
		UpdateParam *CreateStrategyParams
	}

	// QueryStrategyListParams 查询策略列表请求参数
	QueryStrategyListParams struct {
		Keyword       string
		Page          types.Pagination
		Alert         string
		Status        vobj.Status
		SourceType    vobj.StrategyTemplateSource
		StrategyTypes []vobj.StrategyType
	}

	// UpdateStrategyStatusParams 更新策略状态请求参数
	UpdateStrategyStatusParams struct {
		Ids    []uint32 `json:"ids"`
		Status vobj.Status
	}

	// CreateStrategyMetricLevel 创建metric策略等级
	CreateStrategyMetricLevel struct {
		ID uint32 `json:"id"`
		// 所属策略模板id
		StrategyTemplateID uint32 `json:"strategyTemplateID"`
		// 持续时间
		Duration int64 `json:"duration"`
		// 持续次数
		Count uint32 `json:"count"`
		// 持续事件类型
		SustainType vobj.Sustain `json:"sustainType"`
		// 执行频率
		Interval int64 `json:"interval"`
		// 条件
		Condition vobj.Condition `json:"condition"`
		// 阈值
		Threshold float64 `json:"threshold"`
		// 告警等级 对应sys_dict字典id
		LevelID uint32 `json:"LevelID"`
		// 告警页面
		AlarmPageIds []uint32 `json:"alarmPageIds"`
		// 告警组
		AlarmGroupIds []uint32 `json:"alarmGroupIds"`
		// 策略ID
		StrategyID uint32 `json:"strategyID"`
		// 策略标签
		LabelNotices []*StrategyLabelNotice `json:"labelNotices"`
	}

	// CreateStrategyEventLevel 创建事件策略等级
	CreateStrategyEventLevel struct {
		ID uint32 `json:"id"`
		// 值
		Value string `json:"value"`
		// 条件
		Condition vobj.EventCondition `json:"condition"`
		// 数据类型
		EventDataType vobj.EventDataType `json:"mqDataType"`
		// 告警等级 对应sys_dict字典id
		LevelID uint32 `json:"levelID"`
		// 状态
		Status vobj.Status `json:"status"`
		// 告警页面
		AlarmPageIds []uint32 `json:"alarmPageIds"`
		// 告警组
		AlarmGroupIds []uint32 `json:"alarmGroupIds"`
		// 策略ID
		StrategyID uint32 `json:"strategyID"`
		// PathKey
		PathKey string `json:"pathKey"`
	}

	// CreateStrategyDomainLevel 创建域名证书监控策略
	CreateStrategyDomainLevel struct {
		// 状态
		Status vobj.Status `json:"status"`
		// 策略标签
		LabelNotices []*StrategyLabelNotice `json:"labelNotices"`
		// 告警组
		AlarmGroupIds []uint32 `json:"alarmGroupIds"`

		Condition vobj.Condition `json:"condition"`
		// 阈值
		Threshold int64 `json:"threshold"`
		// 策略等级ID
		LevelID uint32 `json:"levelID"`
		// 告警页面ID
		AlarmPageIds []uint32 `json:"alarmPageIds"`
	}

	// CreateStrategyPortLevel 创建端口监控策略
	CreateStrategyPortLevel struct {
		// 状态
		Status vobj.Status `json:"status"`
		// 策略标签
		LabelNotices []*StrategyLabelNotice `json:"labelNotices"`
		// 告警组
		AlarmGroupIds []uint32 `json:"alarmGroupIds"`
		// 阈值
		Threshold int64 `json:"threshold"`
		// 端口
		Port uint32 `json:"port"`
		// 策略等级ID
		LevelID uint32 `json:"levelID"`
		// 告警页面ID
		AlarmPageIds []uint32 `json:"alarmPageIds"`
	}

	// CreateStrategyHTTPLevel 创建http监控策略
	CreateStrategyHTTPLevel struct {
		// 状态
		Status vobj.Status `json:"status"`
		// 策略标签
		LabelNotices []*StrategyLabelNotice `json:"labelNotices"`
		// 告警组
		AlarmGroupIds []uint32 `json:"alarmGroupIds"`
		// 告警页面
		AlarmPageIds []uint32 `json:"alarmPageIds"`
		// 响应时间 s
		ResponseTime uint32 `json:"responseTime"`
		// 状态码
		StatusCode string `json:"statusCodes"`
		// 请求体
		Body string `json:"body"`
		// 查询参数
		QueryParams string `json:"queryParams"`
		// 请求方式
		Method string `json:"method"`
		// 状态码判断条件
		StatusCodeCondition vobj.Condition `json:"condition"`
		// 响应时间判断条件
		ResponseTimeCondition vobj.Condition `json:"responseTimeCondition"`
		// 请求头
		Headers []*HeaderItem `json:"headers"`
		// 策略等级ID
		LevelID uint32 `json:"levelID"`
	}

	// HeaderItem 请求头
	HeaderItem struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	// CreateStrategyGroupParams 创建策略组请求参数
	CreateStrategyGroupParams struct {
		// 策略组名称
		Name string `json:"name,omitempty"`
		// 策略组说明信息
		Remark string `json:"remark,omitempty"`
		// 策略组状态
		Status vobj.Status `json:"status,omitempty"`
		// 策略分组类型
		CategoriesIds []uint32 `json:"categoriesIds,omitempty"`
	}

	// UpdateStrategyGroupStatusParams 更新策略组状态请求参数
	UpdateStrategyGroupStatusParams struct {
		IDs    []uint32 `json:"ids"`
		Status vobj.Status
	}

	// UpdateStrategyGroupParams 更新策略组请求参数
	UpdateStrategyGroupParams struct {
		ID          uint32 `json:"id"`
		UpdateParam *CreateStrategyGroupParams
	}

	// DelStrategyGroupParams 删除策略组请求参数
	DelStrategyGroupParams struct {
		ID uint32 `json:"id"`
	}

	// QueryStrategyGroupListParams 查询策略组列表请求参数
	QueryStrategyGroupListParams struct {
		Keyword       string `json:"keyword"`
		Page          types.Pagination
		Name          string
		Status        vobj.Status
		CategoriesIds []uint32 `json:"categoriesIds"`
	}

	// GetStrategyCountParams 查询策略总数参数
	GetStrategyCountParams struct {
		StrategyGroupIds []uint32 `json:"strategyGroupIds"`
		Status           vobj.Status
	}

	// StrategyCountModel 策略数量统计  策略总数,策略开启总数接收model
	StrategyCountModel struct {
		GroupID uint32 `gorm:"column:group_id"`
		// 总数
		Total uint64 `gorm:"column:total"`
	}
	// StrategyCountMap 策略总数map
	StrategyCountMap struct {
		// 策略开启总数
		StrategyCountMap map[uint32]*StrategyCountModel `json:"strategyCountMap"`
		// 策略总数
		StrategyEnableMap map[uint32]*StrategyCountModel `json:"strategyEnableMap"`
	}

	// StrategyLabelNotice 策略标签
	StrategyLabelNotice struct {
		// 标签名称
		Name string `json:"name"`
		// 标签值
		Value string `json:"value"`
		// 告警组
		AlarmGroupIds []uint32 `json:"alarmGroupIds"`
	}

	// StrategyLevelDetailModel 策略等级明细
	StrategyLevelDetailModel struct {
		// LevelMap
		LevelMap map[uint32][]*bizmodel.StrategyLevel `json:"levelMap"`
	}

	// TeamStrategyLevelModel 策略等级
	TeamStrategyLevelModel struct {
		MetricsLevel *bizmodel.StrategyMetricLevel `json:"metricsLeve"`
		// MQ
		MQLevel *bizmodel.StrategyEventLevel `json:"mqLevel"`
	}
)

// GetStrategyCountMap 获取策略总数
func (s *StrategyCountMap) GetStrategyCountMap(strategyGroupIds uint32) uint64 {
	if types.IsNil(s) {
		return 0
	}
	if v, ok := s.StrategyCountMap[strategyGroupIds]; ok {
		return v.Total
	}
	return 0
}

// GetStrategyEnableMap 获取策略开启总数
func (s *StrategyCountMap) GetStrategyEnableMap(strategyGroupIds uint32) uint64 {
	if types.IsNil(s) {
		return 0
	}
	if v, ok := s.StrategyEnableMap[strategyGroupIds]; ok {
		return v.Total
	}
	return 0
}

var _ watch.Indexer = (*Strategy)(nil)

type (
	// LabelNotices 策略标签通知对象
	LabelNotices struct {
		Key   string `json:"key,omitempty"`
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
		// 等级ID
		LevelID uint32 `json:"levelID,omitempty"`
		// 策略名称
		Alert string `json:"alert,omitempty"`
		// 策略语句
		Expr string `json:"expr,omitempty"`
		// 多数据源持续类型
		MultiDatasourceSustainType vobj.MultiDatasourceSustain `json:"multiDatasourceSustainType,omitempty"`
		// 策略标签
		Labels *vobj.Labels `json:"labels"`
		// 策略注解
		Annotations *vobj.Annotations `json:"annotations"`
		// 数据源
		Datasource []*Datasource `json:"datasource,omitempty"`
		// 策略状态
		Status vobj.Status `json:"status,omitempty"`
		// 团队ID
		TeamID       uint32                     `json:"teamID,omitempty"`
		StrategyType vobj.StrategyType          `json:"strategyType,omitempty"`
		MetricLevel  *CreateStrategyMetricLevel `json:"metricLevels,omitempty"`
		EventLevel   *CreateStrategyEventLevel  `json:"mqLevels,omitempty"`
		// 策略持续时间
		For int64 `json:"for,omitempty"`
		// 持续次数
		Count uint32 `json:"count,omitempty"`
		// 持续的类型
		SustainType vobj.Sustain `json:"sustainType,omitempty"`

		// 执行频率
		Interval int64 `json:"interval,omitempty"`
		// 策略采样率
		Step uint32 `json:"step,omitempty"`
		// 判断条件
		Condition vobj.Condition `json:"condition,omitempty"`
		// 阈值
		Threshold float64 `json:"threshold,omitempty"`
	}

	// Datasource 数据源明细
	Datasource struct {
		// 数据源类型
		Category vobj.DatasourceType `json:"category,omitempty"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storage_type,omitempty"`
		// 数据源配置
		Config *datasource.Config `json:"config,omitempty"`
		// 数据源地址
		Endpoint string `json:"endpoint,omitempty"`
		// 数据源ID
		ID uint32 `json:"id,omitempty"`
		// 状态
		Status vobj.Status
	}
)

// String 字符串
func (s *Strategy) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 策略唯一索引
func (s *Strategy) Index() string {
	if types.IsNil(s) {
		return "-"
	}
	return fmt.Sprintf("%d", s.ID)
}

// Message 策略转消息
func (s *Strategy) Message() *watch.Message {
	if s.StrategyType.IsEvent() {
		return watch.NewMessage(s, vobj.TopicEventStrategy)
	}
	return watch.NewMessage(s, vobj.TopicStrategy)
}
