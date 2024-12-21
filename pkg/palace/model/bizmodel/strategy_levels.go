package bizmodel

import (
	"encoding/json"

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
	RawInfo    *vobj.StrategyRawLevel `gorm:"column:raw_info;type:text;not null;comment:策略等级json"`
	StrategyID uint32                 `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__strategy_id__levels_id" json:"strategyID"`
	Strategy   *Strategy              `gorm:"foreignKey:StrategyID"`
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

// GetStrategyMetricLevel get strategy metric level
func (c *StrategyLevels) GetStrategyMetricLevel() []*StrategyMetricsLevel {
	metricsLevels := make([]*StrategyMetricsLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo.GetRawInfo()), &metricsLevels)
	if err != nil {
		panic("get strategy metricLevel err" + err.Error())
	}
	return metricsLevels
}

// GetStrategyMQLevel get strategy mq level
func (c *StrategyLevels) GetStrategyMQLevel() []*StrategyMQLevel {
	mqLevels := make([]*StrategyMQLevel, 0)
	err := json.Unmarshal([]byte(c.RawInfo.GetRawInfo()), &mqLevels)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return mqLevels
}

// GetStrategyDoMain get strategy domain
func (c *StrategyLevels) GetStrategyDoMain() []*StrategyDomain {
	domains := make([]*StrategyDomain, 0)
	err := json.Unmarshal([]byte(c.RawInfo.GetRawInfo()), &domains)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return domains
}

// GetStrategyHTTP get strategy http
func (c *StrategyLevels) GetStrategyHTTP() []*StrategyHTTP {
	strategyHTTPS := make([]*StrategyHTTP, 0)
	err := json.Unmarshal([]byte(c.RawInfo.GetRawInfo()), &strategyHTTPS)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return strategyHTTPS
}

// GetStrategyPing get strategy ping
func (c *StrategyLevels) GetStrategyPing() []*StrategyPing {
	pings := make([]*StrategyPing, 0)
	err := json.Unmarshal([]byte(c.RawInfo.GetRawInfo()), &pings)
	if err != nil {
		panic("get strategy mqLevel err" + err.Error())
	}
	return pings
}
