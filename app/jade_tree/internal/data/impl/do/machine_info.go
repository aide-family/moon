package do

import (
	"time"

	"github.com/aide-family/jade_tree/pkg/machine"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/hello"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// MachineInfo stores the reported machine hardware profile.
// Rows are deduplicated by the composite of machine_uuid, host_name, and local_ip (machine-id may collide across hosts).
type MachineInfo struct {
	ID        snowflake.ID   `gorm:"column:id;primaryKey"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`

	MachineUUID   string                 `gorm:"column:machine_uuid;size:191;uniqueIndex:idx_machine_identity"`
	HostName      string                 `gorm:"column:host_name;size:191;uniqueIndex:idx_machine_identity"`
	LocalIP       string                 `gorm:"column:local_ip;size:64;uniqueIndex:idx_machine_identity"`
	Source        enum.MachineInfoSource `gorm:"column:source;index;default:1"`
	OSType        string                 `gorm:"column:os_type;size:64;index"`
	AgentEndpoint string                 `gorm:"column:agent_endpoint;size:255"`
	AgentVersion  string                 `gorm:"column:agent_version;size:64;index"`
	Info          *machine.MachineInfo   `gorm:"column:info;type:text"`

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
