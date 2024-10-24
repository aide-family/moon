package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

// HTTP监控策略定义， 用于监控指定URL的响应时间、状态码
const tableNameStrategyHttp = "strategy_http"

type StrategyHttp struct {
	model.AllFieldModel
	Name        string                `gorm:"column:alert;type:varchar(64);not null;comment:策略名称;uniqueIndex:idx__strategy__http__group_id__name,priority:1" json:"name"`
	GroupID     uint32                `gorm:"column:group_id;type:int unsigned;not null;comment:策略规则组ID;uniqueIndex:idx__strategy__http__group_id__name,priority:2" json:"group_id"`
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;uniqueIndex:idx__strategy__http__group_id__name,priority:3" json:"deleted_at"`
	Labels      *vobj.Labels          `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	Annotations vobj.Annotations      `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	Remark      string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status      vobj.Status           `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
	// 告警等级ID
	LevelID uint32 `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级ID" json:"level_id"`
	// 超时时间 seconds
	Timeout uint32 `gorm:"column:timeout;type:int unsigned;not null;comment:超时时间seconds" json:"timeout"`
	// 执行频率
	Interval uint32 `gorm:"column:interval;type:int unsigned;not null;comment:执行频率seconds" json:"interval"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_domain_alarm_groups;" json:"alarm_groups"`
	// StrategyType 策略类型
	StrategyType vobj.StrategyType `gorm:"column:strategy_type;type:int;not null;comment:策略类型" json:"strategy_type"`
	// 策略组
	Group *StrategyGroup `gorm:"foreignKey:GroupID" json:"group"`

	// URL
	URL string `gorm:"column:url;type:varchar(512);not null;comment:URL" json:"url"`
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
