package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// HTTP监控策略定义， 用于监控指定URL的响应时间、状态码
const tableNameStrategyHttp = "strategy_http"

type StrategyHttp struct {
	model.AllFieldModel
	// 所属策略
	StrategyID uint32    `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__http__strategy_id__level_id" json:"strategy_id"`
	Strategy   *Strategy `gorm:"foreignKey:StrategyID" json:"strategy"`
	// 告警等级ID
	LevelID uint32   `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级ID" json:"level_id"`
	Level   *SysDict `gorm:"foreignKey:LevelID" json:"level"`
	// 执行频率
	Interval uint32 `gorm:"column:interval;type:int unsigned;not null;comment:执行频率seconds" json:"interval"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_http_alarm_groups;" json:"alarm_groups"`
	// 状态码
	StatusCodes []int `gorm:"column:status_codes;type:JSON;not null;comment:状态码" json:"status_codes"`
	// 响应时间
	ResponseTime int `gorm:"column:response_time;type:int;not null;comment:响应时间" json:"response_time"`
	// 请求头
	Headers map[string]string `gorm:"column:headers;type:JSON;not null;comment:请求头" json:"headers"`
	// 请求body
	Body string `gorm:"column:body;type:varchar(512);not null;comment:请求body" json:"body"`
	// 请求方式
	Method vobj.HttpMethod `gorm:"column:method;type:int;not null;comment:请求方式" json:"method"`
}

// TableName 表名
func (*StrategyHttp) TableName() string {
	return tableNameStrategyHttp
}

// String 字符串
func (s *StrategyHttp) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyHttp) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyHttp) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
