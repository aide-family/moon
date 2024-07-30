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
		TeamID      uint32 `json:"teamID"`
	}

	// QueryStrategyListParams 查询策略列表请求参数
	QueryStrategyListParams struct {
		Keyword    string `json:"keyword"`
		Page       types.Pagination
		Alert      string
		Status     vobj.Status
		SourceType vobj.TemplateSourceType
		TeamID     uint32 `json:"teamID"`
	}

	// GetStrategyDetailParams 获取策略详情请求参数
	GetStrategyDetailParams struct {
		ID     uint32 `json:"id"`
		TeamID uint32 `json:"teamID"`
	}

	// DelStrategyParams 删除策略请求参数
	DelStrategyParams struct {
		ID     uint32 `json:"id"`
		TeamID uint32 `json:"teamID"`
	}

	// UpdateStrategyStatusParams 更新策略状态请求参数
	UpdateStrategyStatusParams struct {
		Ids    []uint32 `json:"ids"`
		TeamID uint32   `json:"teamID"`
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

	// CopyStrategyParams 复制策略请求参数
	CopyStrategyParams struct {
		StrategyID uint32 `json:"strategyID"`
		TeamID     uint32 `json:"teamID"`
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
		TeamID        uint32   `json:"teamID"`
	}

	// UpdateStrategyGroupStatusParams 更新策略组状态请求参数
	UpdateStrategyGroupStatusParams struct {
		IDs    []uint32 `json:"ids"`
		TeamID uint32   `json:"teamID"`
		Status vobj.Status
	}

	// UpdateStrategyGroupParams 更新策略组请求参数
	UpdateStrategyGroupParams struct {
		ID          uint32 `json:"id"`
		UpdateParam CreateStrategyGroupParams
		TeamID      uint32 `json:"teamID"`
	}

	// GetStrategyGroupDetailParams 获取策略组详情请求参数
	GetStrategyGroupDetailParams struct {
		ID     uint32 `json:"id"`
		TeamID uint32 `json:"teamID"`
	}

	// DelStrategyGroupParams 删除策略组请求参数
	DelStrategyGroupParams struct {
		ID     uint32 `json:"id"`
		TeamID uint32 `json:"teamID"`
	}

	// QueryStrategyGroupListParams 查询策略组列表请求参数
	QueryStrategyGroupListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
		Name    string
		Status  vobj.Status
		TeamID  uint32 `json:"teamID"`
	}
)
