package do

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type Strategy struct {
	BaseModel
	DeletedAt        gorm.DeletedAt              `gorm:"column:deleted_at;uniqueIndex:idx__strategies__namespace_uid__deleted_at__strategy_group_uid"`
	NamespaceUID     snowflake.ID                `gorm:"column:namespace_uid;uniqueIndex:idx__strategies__namespace_uid__deleted_at__strategy_group_uid"`
	StrategyGroupUID snowflake.ID                `gorm:"column:strategy_group_uid;uniqueIndex:idx__strategies__namespace_uid__deleted_at__strategy_group_uid"`
	Name             string                      `gorm:"column:name;type:varchar(100);uniqueIndex:idx__strategies__namespace_uid__deleted_at__strategy_group_uid"`
	Remark           string                      `gorm:"column:remark;type:varchar(100);default:''"`
	Type             enum.DatasourceType         `gorm:"column:type;type:tinyint;default:0"`
	Driver           enum.DatasourceDriver       `gorm:"column:driver;type:tinyint;default:0"`
	Metadata         *safety.Map[string, string] `gorm:"column:metadata;type:json;"`
	Status           enum.GlobalStatus           `gorm:"column:status;type:tinyint;default:0"`
}

func (Strategy) TableName() string {
	return "strategies"
}

func (s *Strategy) WithNamespace(namespace snowflake.ID) *Strategy {
	s.NamespaceUID = namespace
	return s
}
