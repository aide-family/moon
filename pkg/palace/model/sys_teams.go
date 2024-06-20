package model

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameSysTeam = "sys_teams"

// SysTeam mapped from table <sys_teams>
type SysTeam struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`
	Name      string                `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:团队空间名" json:"name"`               // 团队空间名
	Status    vobj.Status           `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                       // 状态
	Remark    string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                              // 备注
	Logo      string                `gorm:"column:logo;type:varchar(255);not null;comment:团队logo" json:"logo"`                                              // 团队logo
	LeaderID  uint32                `gorm:"column:leader_id;type:int unsigned;not null;index:sys_teams__sys_users,priority:1;comment:负责人" json:"leader_id"` // 负责人
	CreatorID uint32                `gorm:"column:creator_id;type:int unsigned;not null;comment:创建者" json:"creator_id"`                                     // 创建者
	UUID      string                `gorm:"column:uuid;type:varchar(64);not null" json:"uuid"`
}

// String json string
func (c *SysTeam) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *SysTeam) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *SysTeam) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *SysTeam) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *SysTeam) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *SysTeam) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName SysTeam's table name
func (*SysTeam) TableName() string {
	return TableNameSysTeam
}
