package do

import (
	"errors"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type Datasource struct {
	ID           uint32                      `gorm:"column:id;primaryKey;autoIncrement"`
	UID          snowflake.ID                `gorm:"column:uid;uniqueIndex"`
	CreatedAt    time.Time                   `gorm:"column:created_at;"`
	UpdatedAt    time.Time                   `gorm:"column:updated_at;"`
	Creator      snowflake.ID                `gorm:"column:creator;index"`
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

func (d *Datasource) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	if d.Creator == 0 {
		d.Creator = contextx.GetUserUID(ctx)
	}
	if d.NamespaceUID == 0 {
		return errors.New("namespace uid is required")
	}
	if d.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
