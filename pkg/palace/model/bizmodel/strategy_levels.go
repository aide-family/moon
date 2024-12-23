package bizmodel

import (
	"encoding/json"

	"gorm.io/gorm"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameStrategyLevels = "strategy_levels"

// StrategyLevels mapped from table <StrategyLevels>
type StrategyLevels struct {
	model.AllFieldModel
	// StrategyType 策略类型
	StrategyType vobj.StrategyType `gorm:"column:strategy_type;type:int;not null;comment:策略类型" json:"strategy_type"`
	// 策略等级json
	RawInfo    string    `gorm:"column:raw_info;type:text;not null;comment:策略等级json"`
	StrategyID uint32    `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__strategy_id__levels_id" json:"strategyID"`
	Strategy   *Strategy `gorm:"foreignKey:StrategyID"`

	// 映射数据
	StrategyMetricsLevelList []*StrategyMetricsLevel `gorm:"-" json:"strategyMetricsLevelList,omitempty"`
	StrategyMQLevelList      []*StrategyEventLevel   `gorm:"-" json:"strategyMQLevelList,omitempty"`
	StrategyDomainList       []*StrategyDomain       `gorm:"-" json:"strategyDomainList,omitempty"`
	StrategyPortList         []*StrategyPort         `gorm:"-" json:"strategyPortList,omitempty"`
	StrategyHTTPList         []*StrategyHTTP         `gorm:"-" json:"strategyHTTPList,omitempty"`
	StrategyPingList         []*StrategyPing         `gorm:"-" json:"strategyPingList,omitempty"`
}

// AfterFind get strategy level
func (c *StrategyLevels) AfterFind(tx *gorm.DB) (err error) {
	if c.RawInfo == "" {
		return nil
	}
	switch c.StrategyType {
	case vobj.StrategyTypeMetric:
		c.StrategyMetricsLevelList = c.getStrategyMetricLevel()
	case vobj.StrategyTypeMQ:
		c.StrategyMQLevelList = c.getStrategyMQLevel()
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
func (c *StrategyLevels) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyLevels) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyLevels) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategyLevels's table name
func (*StrategyLevels) TableName() string {
	return tableNameStrategyLevels
}

// getStrategyMetricLevel get strategy metric level
func (c *StrategyLevels) getStrategyMetricLevel() []*StrategyMetricsLevel {
	metricsLevels := make([]*StrategyMetricsLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &metricsLevels)
	if err != nil {
		panic("get strategy metricLevel err" + err.Error())
	}
	return metricsLevels
}

// GetStrategyMQLevel get strategy mq level
func (c *StrategyLevels) getStrategyMQLevel() []*StrategyEventLevel {
	mqLevels := make([]*StrategyEventLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &mqLevels)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return mqLevels
}

// GetStrategyDoMain get strategy domain
func (c *StrategyLevels) getStrategyDoMain() []*StrategyDomain {
	domains := make([]*StrategyDomain, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &domains)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return domains
}

// GetStrategyHTTP get strategy http
func (c *StrategyLevels) getStrategyHTTP() []*StrategyHTTP {
	strategyHTTPS := make([]*StrategyHTTP, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &strategyHTTPS)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return strategyHTTPS
}

// GetStrategyPing get strategy ping
func (c *StrategyLevels) getStrategyPing() []*StrategyPing {
	pings := make([]*StrategyPing, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &pings)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return pings
}

// getStrategyPort get strategy port
func (c *StrategyLevels) getStrategyPort() []*StrategyPort {
	ports := make([]*StrategyPort, 0)
	err := json.Unmarshal([]byte(c.RawInfo), &ports)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return ports
}
