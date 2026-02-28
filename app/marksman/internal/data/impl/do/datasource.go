package do

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type Datasource struct {
	BaseModel
	DeletedAt    gorm.DeletedAt              `gorm:"column:deleted_at;uniqueIndex:idx__datasources__namespace_uid__deleted_at__name"`
	NamespaceUID snowflake.ID                `gorm:"column:namespace_uid;uniqueIndex:idx__datasources__namespace_uid__deleted_at__name"`
	Name         string                      `gorm:"column:name;type:varchar(100);uniqueIndex:idx__datasources__namespace_uid__deleted_at__name"`
	Type         enum.DatasourceType         `gorm:"column:type;type:tinyint;default:0"`
	Driver       enum.DatasourceDriver       `gorm:"column:driver;type:tinyint;default:0"`
	Metadata     *safety.Map[string, string] `gorm:"column:metadata;type:json;"`
	Status       enum.GlobalStatus           `gorm:"column:status;type:tinyint;default:0"`
}

func (Datasource) TableName() string {
	return "datasources"
}

func (d *Datasource) WithNamespace(namespace snowflake.ID) *Datasource {
	d.NamespaceUID = namespace
	return d
}
