package do

import (
	"errors"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type StrategyGroup struct {
	BaseModel
	DeletedAt    gorm.DeletedAt              `gorm:"column:deleted_at;uniqueIndex:idx__strategy_groups__namespace_uid__deleted_at__name"`
	NamespaceUID snowflake.ID                `gorm:"column:namespace_uid;uniqueIndex:idx__strategy_groups__namespace_uid__deleted_at__name"`
	Name         string                      `gorm:"column:name;type:varchar(100);uniqueIndex:idx__strategy_groups__namespace_uid__deleted_at__name"`
	Remark       string                      `gorm:"column:remark;type:varchar(100);default:''"`
	Metadata     *safety.Map[string, string] `gorm:"column:metadata;type:json;"`
	Status       enum.GlobalStatus           `gorm:"column:status;type:tinyint;default:0"`
}

func (StrategyGroup) TableName() string {
	return "strategy_groups"
}

func (s *StrategyGroup) WithNamespace(namespace snowflake.ID) *StrategyGroup {
	s.NamespaceUID = namespace
	return s
}

func (s *StrategyGroup) BeforeCreate(tx *gorm.DB) (err error) {
	if s.NamespaceUID == 0 {
		return errors.New("namespace uid is required")
	}
	if s.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
