package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

// Ping监控策略定义， 用于监控指定IP的网络延迟、丢包率等
const tableNameStrategyPing = "strategy_ping"

type StrategyPing struct {
	model.AllFieldModel
	Name        string                `gorm:"column:alert;type:varchar(64);not null;comment:策略名称;uniqueIndex:idx__strategy__ping__group_id__name,priority:1" json:"name"`
	GroupID     uint32                `gorm:"column:group_id;type:int unsigned;not null;comment:策略规则组ID;uniqueIndex:idx__strategy__ping__group_id__name,priority:2" json:"group_id"`
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;uniqueIndex:idx__strategy__ping__group_id__name,priority:3" json:"deleted_at"`
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

	// 域名或者ip
	Target string `gorm:"column:target;type:varchar(255);not null;comment:域名或者ip" json:"target"`
	// 总包数
	Total uint32 `gorm:"column:total;type:int unsigned;not null;comment:总包数" json:"total"`
	// 成功包数
	Success uint32 `gorm:"column:success;type:int unsigned;not null;comment:成功包数" json:"success"`
	// 丢包率
	LossRate float64 `gorm:"column:loss_rate;type:float;not null;comment:丢包率" json:"loss_rate"`
	// 平均延迟
	AvgDelay uint32 `gorm:"column:avg_delay;type:int unsigned;not null;comment:平均延迟" json:"avg_delay"`
	// 最大延迟
	MaxDelay uint32 `gorm:"column:max_delay;type:int unsigned;not null;comment:最大延迟" json:"max_delay"`
	// 最小延迟
	MinDelay uint32 `gorm:"column:min_delay;type:int unsigned;not null;comment:最小延迟" json:"min_delay"`
	// 标准差
	StdDev uint32 `gorm:"column:std_dev;type:int unsigned;not null;comment:标准差" json:"std_dev"`
}

// TableName 表名
func (*StrategyPing) TableName() string {
	return tableNameStrategyPing
}

// String 字符串
func (s *StrategyPing) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyPing) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyPing) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
