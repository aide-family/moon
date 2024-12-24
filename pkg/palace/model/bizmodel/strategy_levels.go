package bizmodel

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameStrategyLevels = "strategy_levels"

// StrategyLevel mapped from table <StrategyLevel>
type StrategyLevel struct {
	model.AllFieldModel
	// StrategyType 策略类型
	StrategyType vobj.StrategyType `gorm:"column:strategy_type;type:int;not null;comment:策略类型" json:"strategy_type"`
	// 策略等级json
	RawInfo    string    `gorm:"column:raw_info;type:text;not null;comment:策略等级json"`
	StrategyID uint32    `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__strategy_id__levels_id" json:"strategyID"`
	Strategy   *Strategy `gorm:"foreignKey:StrategyID"`
	// 告警页面 + 告警等级
	DictList []*SysDict `gorm:"many2many:strategy_level_dict_list" json:"dictList"`
	// 告警组列表
	AlarmGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_levels_alarm_groups"`

	dictMap       map[uint32]*SysDict
	alarmGroupMap map[uint32]*AlarmNoticeGroup

	// 映射数据
	StrategyMetricsLevelList []*StrategyMetricLevel `gorm:"-" json:"strategyMetricsLevelList,omitempty"`
	StrategyEventLevelList   []*StrategyEventLevel  `gorm:"-" json:"strategyMQLevelList,omitempty"`
	StrategyDomainList       []*StrategyDomainLevel `gorm:"-" json:"strategyDomainList,omitempty"`
	StrategyPortList         []*StrategyPortLevel   `gorm:"-" json:"strategyPortList,omitempty"`
	StrategyHTTPList         []*StrategyHTTPLevel   `gorm:"-" json:"strategyHTTPList,omitempty"`
	StrategyPingList         []*StrategyPingLevel   `gorm:"-" json:"strategyPingList,omitempty"`
}

// AfterFind get strategy level
func (c *StrategyLevel) AfterFind(_ *gorm.DB) (err error) {
	if c.RawInfo == "" {
		return nil
	}
	c.dictMap = types.ToMap(c.DictList, func(item *SysDict) uint32 {
		return item.GetID()
	})
	c.alarmGroupMap = types.ToMap(c.AlarmGroups, func(item *AlarmNoticeGroup) uint32 {
		return item.GetID()
	})

	switch c.StrategyType {
	case vobj.StrategyTypeMetric:
		c.StrategyMetricsLevelList = c.getStrategyMetricLevel()
	case vobj.StrategyTypeEvent:
		c.StrategyEventLevelList = c.getStrategyEventLevel()
	case vobj.StrategyTypeDomainCertificate:
		c.StrategyDomainList = c.getStrategyDoMain()
	case vobj.StrategyTypeHTTP:
		c.StrategyHTTPList = c.getStrategyHTTP()
	case vobj.StrategyTypePing:
		c.StrategyPingList = c.getStrategyPing()
	case vobj.StrategyTypeDomainPort:
		c.StrategyPortList = c.getStrategyPort()
	default:
	}
	return
}

// String json string
func (c *StrategyLevel) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategyLevel's table name
func (*StrategyLevel) TableName() string {
	return tableNameStrategyLevels
}

// getStrategyMetricLevel get strategy metric level
func (c *StrategyLevel) getStrategyMetricLevel() []*StrategyMetricLevel {
	metricsLevels := make([]*StrategyMetricLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &metricsLevels)
	if err != nil {
		log.Warnw("method", "get strategy metricLevel", "err", err)
		return nil
	}
	return types.SliceTo(metricsLevels, func(item *StrategyMetricLevel) *StrategyMetricLevel {
		item.Level = c.dictMap[item.Level.GetID()]
		item.AlarmPageList = types.SliceTo(item.AlarmPageList, func(dictItem *SysDict) *SysDict {
			return c.dictMap[dictItem.GetID()]
		})
		item.AlarmGroupList = types.SliceTo(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) *AlarmNoticeGroup {
			return c.alarmGroupMap[alarmGroupItem.GetID()]
		})
		item.LabelNoticeList = types.SliceTo(item.LabelNoticeList, func(labelNoticeItem *StrategyMetricsLabelNotice) *StrategyMetricsLabelNotice {
			return &StrategyMetricsLabelNotice{
				Name:  labelNoticeItem.Name,
				Value: labelNoticeItem.Value,
				AlarmGroups: types.SliceTo(labelNoticeItem.AlarmGroups, func(groupItem *AlarmNoticeGroup) *AlarmNoticeGroup {
					return c.alarmGroupMap[groupItem.GetID()]
				}),
			}
		})
		return item
	})
}

// GetStrategyMQLevel get strategy mq level
func (c *StrategyLevel) getStrategyEventLevel() []*StrategyEventLevel {
	eventLevels := make([]*StrategyEventLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &eventLevels)
	if err != nil {
		log.Warnw("method", "get strategy eventLevel", "err", err)
		return nil
	}
	return types.SliceTo(eventLevels, func(item *StrategyEventLevel) *StrategyEventLevel {
		item.Level = c.dictMap[item.Level.GetID()]
		item.AlarmPageList = types.SliceTo(item.AlarmPageList, func(dictItem *SysDict) *SysDict {
			return c.dictMap[dictItem.GetID()]
		})
		item.AlarmGroupList = types.SliceTo(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) *AlarmNoticeGroup {
			return c.alarmGroupMap[alarmGroupItem.GetID()]
		})
		return item
	})
}

// GetStrategyDoMain get strategy domain
func (c *StrategyLevel) getStrategyDoMain() []*StrategyDomainLevel {
	domains := make([]*StrategyDomainLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &domains)
	if err != nil {
		log.Warnw("method", "get strategy domain", "err", err)
		return nil
	}
	return types.SliceTo(domains, func(item *StrategyDomainLevel) *StrategyDomainLevel {
		item.Level = c.dictMap[item.Level.GetID()]
		item.AlarmPageList = types.SliceTo(item.AlarmPageList, func(dictItem *SysDict) *SysDict {
			return c.dictMap[dictItem.GetID()]
		})
		item.AlarmGroupList = types.SliceTo(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) *AlarmNoticeGroup {
			return c.alarmGroupMap[alarmGroupItem.GetID()]
		})
		return item
	})
}

// GetStrategyHTTP get strategy http
func (c *StrategyLevel) getStrategyHTTP() []*StrategyHTTPLevel {
	strategyHTTPS := make([]*StrategyHTTPLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &strategyHTTPS)
	if err != nil {
		log.Warnw("method", "get strategy http", "err", err)
		return nil
	}
	return types.SliceTo(strategyHTTPS, func(item *StrategyHTTPLevel) *StrategyHTTPLevel {
		item.Level = c.dictMap[item.Level.GetID()]
		item.AlarmPageList = types.SliceTo(item.AlarmPageList, func(dictItem *SysDict) *SysDict {
			return c.dictMap[dictItem.GetID()]
		})
		item.AlarmGroupList = types.SliceTo(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) *AlarmNoticeGroup {
			return c.alarmGroupMap[alarmGroupItem.GetID()]
		})
		return item
	})
}

// GetStrategyPing get strategy ping
func (c *StrategyLevel) getStrategyPing() []*StrategyPingLevel {
	pings := make([]*StrategyPingLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &pings)
	if err != nil {
		log.Warnw("method", "get strategy ping", "err", err)
		return nil
	}
	return types.SliceTo(pings, func(item *StrategyPingLevel) *StrategyPingLevel {
		item.Level = c.dictMap[item.Level.GetID()]
		item.AlarmPageList = types.SliceTo(item.AlarmPageList, func(dictItem *SysDict) *SysDict {
			return c.dictMap[dictItem.GetID()]
		})
		item.AlarmGroupList = types.SliceTo(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) *AlarmNoticeGroup {
			return c.alarmGroupMap[alarmGroupItem.GetID()]
		})
		return item
	})
}

// getStrategyPort get strategy port
func (c *StrategyLevel) getStrategyPort() []*StrategyPortLevel {
	ports := make([]*StrategyPortLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &ports)
	if err != nil {
		log.Warnw("method", "get strategy port", "err", err)
		return nil
	}
	return types.SliceTo(ports, func(item *StrategyPortLevel) *StrategyPortLevel {
		item.Level = c.dictMap[item.Level.GetID()]
		item.AlarmPageList = types.SliceTo(item.AlarmPageList, func(dictItem *SysDict) *SysDict {
			return c.dictMap[dictItem.GetID()]
		})
		item.AlarmGroupList = types.SliceTo(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) *AlarmNoticeGroup {
			return c.alarmGroupMap[alarmGroupItem.GetID()]
		})
		return item
	})
}
