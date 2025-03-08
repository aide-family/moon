package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

const tableNameStrategyLevels = "strategy_levels"

// StrategyLevel mapped from table <StrategyLevel>
type StrategyLevel struct {
	AllFieldModel
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
	StrategyDomainLevelList  []*StrategyDomainLevel `gorm:"-" json:"strategyDomainList,omitempty"`
	StrategyPortLevelList    []*StrategyPortLevel   `gorm:"-" json:"strategyPortList,omitempty"`
	StrategyHTTPLevelList    []*StrategyHTTPLevel   `gorm:"-" json:"strategyHTTPList,omitempty"`
	StrategyPingLevelList    []*StrategyPingLevel   `gorm:"-" json:"strategyPingList,omitempty"`
	StrategyLogsLevelList    []*StrategyLogsLevel   `gorm:"-" json:"strategyLogsList,omitempty"`
}

// GetStrategy 获取策略
func (c *StrategyLevel) GetStrategy() *Strategy {
	if c == nil {
		return nil
	}
	return c.Strategy
}

// GetStrategyMetricsLevelList 获取metric策略等级列表
func (c *StrategyLevel) GetStrategyMetricsLevelList() []*StrategyMetricLevel {
	if c == nil {
		return nil
	}
	return c.StrategyMetricsLevelList
}

// GetStrategyEventLevelList 获取event策略等级列表
func (c *StrategyLevel) GetStrategyEventLevelList() []*StrategyEventLevel {
	if c == nil {
		return nil
	}
	return c.StrategyEventLevelList
}

// GetStrategyDomainLevelList 获取domain策略等级列表
func (c *StrategyLevel) GetStrategyDomainLevelList() []*StrategyDomainLevel {
	if c == nil {
		return nil
	}
	return c.StrategyDomainLevelList
}

// GetStrategyPortLevelList 获取port策略等级列表
func (c *StrategyLevel) GetStrategyPortLevelList() []*StrategyPortLevel {
	if c == nil {
		return nil
	}
	return c.StrategyPortLevelList
}

// GetStrategyHTTPLevelList 获取http策略等级列表
func (c *StrategyLevel) GetStrategyHTTPLevelList() []*StrategyHTTPLevel {
	if c == nil {
		return nil
	}
	return c.StrategyHTTPLevelList
}

// GetStrategyPingLevelList 获取ping策略等级列表
func (c *StrategyLevel) GetStrategyPingLevelList() []*StrategyPingLevel {
	if c == nil {
		return nil
	}
	return c.StrategyPingLevelList
}

// GetStrategyLogLevelList 获取日志策略等级列表
func (c *StrategyLevel) GetStrategyLogLevelList() []*StrategyLogsLevel {
	if c == nil {
		return nil
	}
	return c.StrategyLogsLevelList
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
		c.StrategyDomainLevelList = c.getStrategyDoMain()
	case vobj.StrategyTypeHTTP:
		c.StrategyHTTPLevelList = c.getStrategyHTTP()
	case vobj.StrategyTypePing:
		c.StrategyPingLevelList = c.getStrategyPing()
	case vobj.StrategyTypeDomainPort:
		c.StrategyPortLevelList = c.getStrategyPort()
	case vobj.StrategyTypeLogs:
		c.StrategyLogsLevelList = c.getStrategyLogs()
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

// GetAlarmPageList 获取告警页面列表
func (c *StrategyLevel) GetAlarmPageList() []*SysDict {
	return types.Filter(c.DictList, func(item *SysDict) bool {
		return item.DictType.IsAlarmPage()
	})
}

// GetLevelByID 获取等级
func (c *StrategyLevel) GetLevelByID(id uint32) string {
	switch c.StrategyType {
	case vobj.StrategyTypeMetric:
		levels := c.getStrategyMetricLevel()
		level := types.SliceFind(levels, func(item *StrategyMetricLevel) bool {
			return item.Level.GetID() == id
		})
		return level.String()
	case vobj.StrategyTypeEvent:
		levels := c.getStrategyEventLevel()
		level := types.SliceFind(levels, func(item *StrategyEventLevel) bool {
			return item.Level.GetID() == id
		})
		return level.String()
	case vobj.StrategyTypeDomainCertificate:
		levels := c.getStrategyDoMain()
		level := types.SliceFind(levels, func(item *StrategyDomainLevel) bool {
			return item.Level.GetID() == id
		})
		return level.String()
	case vobj.StrategyTypeHTTP:
		levels := c.getStrategyHTTP()
		level := types.SliceFind(levels, func(item *StrategyHTTPLevel) bool {
			return item.Level.GetID() == id
		})
		return level.String()
	case vobj.StrategyTypePing:
		levels := c.getStrategyPing()
		level := types.SliceFind(levels, func(item *StrategyPingLevel) bool {
			return item.Level.GetID() == id
		})
		return level.String()
	case vobj.StrategyTypeDomainPort:
		levels := c.getStrategyPort()
		level := types.SliceFind(levels, func(item *StrategyPortLevel) bool {
			return item.Level.GetID() == id
		})
		return level.String()
	case vobj.StrategyTypeLogs:
		levels := c.getStrategyLogs()
		level := types.SliceFind(levels, func(item *StrategyLogsLevel) bool {
			return item.Level.GetID() == id
		})
		return level.String()
	default:
		return "{}"
	}
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
		item.AlarmPageList = types.SliceToWithFilter(item.AlarmPageList, func(dictItem *SysDict) (*SysDict, bool) {
			dict, ok := c.dictMap[dictItem.GetID()]
			return dict, ok
		})
		item.AlarmGroupList = types.SliceToWithFilter(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
			group, ok := c.alarmGroupMap[alarmGroupItem.GetID()]
			return group, ok
		})
		item.LabelNoticeList = types.SliceToWithFilter(item.LabelNoticeList, func(labelNoticeItem *StrategyMetricsLabelNotice) (*StrategyMetricsLabelNotice, bool) {
			notice := &StrategyMetricsLabelNotice{
				Name:  labelNoticeItem.Name,
				Value: labelNoticeItem.Value,
				AlarmGroups: types.SliceToWithFilter(labelNoticeItem.AlarmGroups, func(groupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
					group, ok := c.alarmGroupMap[groupItem.GetID()]
					return group, ok
				}),
			}
			return notice, len(notice.AlarmGroups) > 0
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
		item.AlarmPageList = types.SliceToWithFilter(item.AlarmPageList, func(dictItem *SysDict) (*SysDict, bool) {
			dict, ok := c.dictMap[dictItem.GetID()]
			return dict, ok
		})
		item.AlarmGroupList = types.SliceToWithFilter(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
			group, ok := c.alarmGroupMap[alarmGroupItem.GetID()]
			return group, ok
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
		item.AlarmPageList = types.SliceToWithFilter(item.AlarmPageList, func(dictItem *SysDict) (*SysDict, bool) {
			dict, ok := c.dictMap[dictItem.GetID()]
			return dict, ok
		})
		item.AlarmGroupList = types.SliceToWithFilter(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
			group, ok := c.alarmGroupMap[alarmGroupItem.GetID()]
			return group, ok
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
		item.AlarmPageList = types.SliceToWithFilter(item.AlarmPageList, func(dictItem *SysDict) (*SysDict, bool) {
			dict, ok := c.dictMap[dictItem.GetID()]
			return dict, ok
		})
		item.AlarmGroupList = types.SliceToWithFilter(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
			group, ok := c.alarmGroupMap[alarmGroupItem.GetID()]
			return group, ok
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
		item.AlarmPageList = types.SliceToWithFilter(item.AlarmPageList, func(dictItem *SysDict) (*SysDict, bool) {
			dict, ok := c.dictMap[dictItem.GetID()]
			return dict, ok
		})
		item.AlarmGroupList = types.SliceToWithFilter(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
			group, ok := c.alarmGroupMap[alarmGroupItem.GetID()]
			return group, ok
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
		item.AlarmPageList = types.SliceToWithFilter(item.AlarmPageList, func(dictItem *SysDict) (*SysDict, bool) {
			dict, ok := c.dictMap[dictItem.GetID()]
			return dict, ok
		})
		item.AlarmGroupList = types.SliceToWithFilter(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
			group, ok := c.alarmGroupMap[alarmGroupItem.GetID()]
			return group, ok
		})
		return item
	})
}

// getStrategyLogs get strategy logs
func (c *StrategyLevel) getStrategyLogs() []*StrategyLogsLevel {
	ports := make([]*StrategyLogsLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &ports)
	if err != nil {
		log.Warnw("method", "get strategy port", "err", err)
		return nil
	}
	return types.SliceTo(ports, func(item *StrategyLogsLevel) *StrategyLogsLevel {
		item.Level = c.dictMap[item.Level.GetID()]
		item.AlarmPageList = types.SliceToWithFilter(item.AlarmPageList, func(dictItem *SysDict) (*SysDict, bool) {
			dict, ok := c.dictMap[dictItem.GetID()]
			return dict, ok
		})
		item.AlarmGroupList = types.SliceToWithFilter(item.AlarmGroupList, func(alarmGroupItem *AlarmNoticeGroup) (*AlarmNoticeGroup, bool) {
			group, ok := c.alarmGroupMap[alarmGroupItem.GetID()]
			return group, ok
		})
		return item
	})
}
