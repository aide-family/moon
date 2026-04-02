package do

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// MachineInfo stores the reported machine hardware profile.
// MachineUUID is the deduplication key and also the primary key for upsert semantics.
type MachineInfo struct {
	ID        snowflake.ID   `gorm:"column:id;primaryKey"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`

	MachineUUID string                 `gorm:"column:machine_uuid;size:191"`
	HostName    string                 `gorm:"column:host_name;size:191"`
	Source      enum.MachineInfoSource `gorm:"column:source;index;default:1"`
	Info        string                 `gorm:"column:info;type:text"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (MachineInfo) TableName() string { return "machine_infos" }

func (m *MachineInfo) BeforeCreate(tx *gorm.DB) error {
	if m.ID == 0 {
		node, err := snowflake.NewNode(hello.NodeID())
		if err != nil {
			return err
		}
		m.ID = node.Generate()
	}
	return nil
}
