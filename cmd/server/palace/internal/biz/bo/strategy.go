package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateStrategyParams struct {
		// 策略组ID
		GroupId uint32 `json:"group_id"`
		// 策略模板id
		TemplateId uint32 `json:"template_id"`
		// 备注
		Remark string `json:"remark"`
		// 状态
		Status vobj.Status `json:"status"`
		// 采样率
		Step uint32 `json:"step"`
		// 数据源id
		DatasourceIds []uint32 `json:"datasource_ids"`
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

	UpdateStrategyParams struct {
		ID          uint32 `json:"id"`
		UpdateParam CreateStrategyParams
		TeamID      uint32 `json:"teamID"`
	}

	QueryStrategyListParams struct {
		Keyword    string `json:"keyword"`
		Page       types.Pagination
		Alert      string
		Status     vobj.Status
		SourceType vobj.TemplateSourceType
		TeamID     uint32 `json:"teamID"`
	}

	GetStrategyDetailParams struct {
		ID     uint32 `json:"id"`
		TeamID uint32 `json:"teamID"`
	}

	DelStrategyParams struct {
		ID     uint32 `json:"id"`
		TeamID uint32 `json:"teamID"`
	}

	UpdateStrategyStatusParams struct {
		Ids    []uint32 `json:"ids"`
		TeamID uint32   `json:"teamID"`
		Status vobj.Status
	}

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

	CopyStrategyParams struct {
		StrategyID uint32 `json:"strategyID"`
		TeamID     uint32 `json:"teamID"`
	}
)
