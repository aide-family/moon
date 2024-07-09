package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateTemplateStrategyParams struct {
		// 策略名称
		Alert string `json:"alert"`
		// 告警表达式
		Expr string `json:"expr"`
		// 策略状态
		Status vobj.Status `json:"status"`
		// 备注
		Remark string `json:"remark"`
		// 标签
		Labels vobj.Labels `json:"labels"`
		// 注解
		Annotations vobj.Annotations `json:"annotations"`
		// 模板等级
		// 告警等级数据
		StrategyLevelTemplates []*CreateStrategyLevelTemplate `gorm:"foreignKey:StrategyTemplateID" json:"strategyLevelTemplates"`

		//策略模板类型
		CategoriesIDs []uint32 `json:"categoriesIds"`
	}

	UpdateTemplateStrategyParams struct {
		ID   uint32                       `json:"id"`
		Data CreateTemplateStrategyParams `json:"data"`
	}

	QueryTemplateStrategyListParams struct {
		Keyword string `json:"keyword"`
		Page    types.Pagination
		Alert   string
		Status  vobj.Status
	}

	CreateStrategyAlarmLevel struct {
		//*model.StrategyAlarmLevel
	}

	UpdateStrategyAlarmLevel struct {
		//*model.StrategyAlarmLevel
	}

	QueryStrategyAlarmLevelListParams struct {
		Page    types.Pagination
		Keyword string
		Status  vobj.Status
	}

	CreateStrategyLevelTemplate struct {
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
		Condition string `json:"condition"`
		// 阈值
		Threshold float64 `json:"threshold"`
		// 告警等级 对应sys_dict字典id
		LevelID uint32 `json:"LevelID"`
		// 状态
		Status vobj.Status `json:"status"`
	}
)
