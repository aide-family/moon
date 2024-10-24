package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

// 证书过期、端口开闭监控策略
const tableNameStrategyDomain = "strategy_domain"

type StrategyDomain struct {
	model.AllFieldModel
	Name        string                `gorm:"column:alert;type:varchar(64);not null;comment:策略名称;uniqueIndex:idx__strategy__domain__group_id__name,priority:1" json:"name"`
	GroupID     uint32                `gorm:"column:group_id;type:int unsigned;not null;comment:策略规则组ID;uniqueIndex:idx__strategy__domain__group_id__name,priority:2" json:"group_id"`
	DeletedAt   soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;uniqueIndex:idx__strategy__domain__group_id__name,priority:3" json:"deleted_at"`
	Labels      *vobj.Labels          `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	Annotations vobj.Annotations      `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	Remark      string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status      vobj.Status           `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
	// 告警等级ID
	LevelID uint32 `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级ID" json:"level_id"`
	// 域名或者IP
	Domain string `gorm:"column:domain;type:varchar(255);not null;comment:域名或者IP" json:"domain"`
	// 端口
	Port uint16 `gorm:"column:port;type:int unsigned;not null;comment:端口" json:"port"`
	// 超时时间 seconds
	Timeout uint32 `gorm:"column:timeout;type:int unsigned;not null;comment:超时时间seconds" json:"timeout"`
	// 执行频率
	Interval uint32 `gorm:"column:interval;type:int unsigned;not null;comment:执行频率seconds" json:"interval"`
	// 阈值 （证书类型就是剩余天数，端口就是0：关闭，1：开启）
	Threshold uint32 `gorm:"column:threshold;type:int unsigned;not null;comment:阈值" json:"threshold"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_domain_alarm_groups;" json:"alarm_groups"`
	// StrategyType 策略类型
	StrategyType vobj.StrategyType `gorm:"column:strategy_type;type:int;not null;comment:策略类型" json:"strategy_type"`
	// 策略组
	Group *StrategyGroup `gorm:"foreignKey:GroupID" json:"group"`
}

// TableName 表名
func (*StrategyDomain) TableName() string {
	return tableNameStrategyDomain
}

// String 字符串
func (s *StrategyDomain) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyDomain) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyDomain) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
