package bo

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
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
		// 采样率
		Step uint32 `json:"step"`
		// 数据源id
		DatasourceIDs []uint32 `json:"datasource_ids"`
		// 模板来源
		SourceType vobj.TemplateSourceType `json:"source_type"`
		// 策略名称
		Name   string `json:"name"`
		TeamID uint32 `json:"teamID"`
		// 策略等级
		StrategyLevel []*CreateStrategyLevel `json:"strategyLevel"`
		// 标签
		Labels *vobj.Labels `json:"labels"`
		// 注解
		Annotations vobj.Annotations `json:"annotations"`
	}

	// UpdateStrategyParams 更新策略请求参数
	UpdateStrategyParams struct {
		ID          uint32 `json:"id"`
		UpdateParam CreateStrategyParams
	}

	// QueryStrategyListParams 查询策略列表请求参数
	QueryStrategyListParams struct {
		Keyword    string `json:"keyword"`
		Page       types.Pagination
		Alert      string
		Status     vobj.Status
		SourceType vobj.TemplateSourceType
	}

	// UpdateStrategyStatusParams 更新策略状态请求参数
	UpdateStrategyStatusParams struct {
		Ids    []uint32 `json:"ids"`
		Status vobj.Status
	}

	// CreateStrategyLevel 策略模板策略等级
	CreateStrategyLevel struct {
		// 所属策略模板id
		StrategyTemplateID uint32 `json:"strategyTemplateID"`
		// 持续时间
		Duration *types.Duration `json:"duration"`
		// 持续次数
		Count uint32 `json:"count"`
		// 持续事件类型
		SustainType vobj.Sustain `json:"sustainType"`

		// 执行频率
		Interval *types.Duration `json:"interval"`
		// 条件
		Condition vobj.Condition `json:"condition"`
		// 阈值
		Threshold float64 `json:"threshold"`
		// 告警等级 对应sys_dict字典id
		LevelID uint32 `json:"LevelID"`
		// 状态
		Status vobj.Status `json:"status"`
	}

	// CreateStrategyGroupParams 创建策略组请求参数
	CreateStrategyGroupParams struct {
		// 策略组名称
		Name string `json:"name,omitempty"`
		// 策略组说明信息
		Remark string `json:"remark,omitempty"`
		// 策略组状态
		Status api.Status `json:"status,omitempty"`
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
		UpdateParam CreateStrategyGroupParams
	}

	// DelStrategyGroupParams 删除策略组请求参数
	DelStrategyGroupParams struct {
		ID uint32 `json:"id"`
	}

	// QueryStrategyGroupListParams 查询策略组列表请求参数
	QueryStrategyGroupListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
		Name    string
		Status  vobj.Status
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
)
