package bizmodel

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameSysTeamRole = "sys_team_roles"

// SysTeamRole mapped from table <sys_team_roles>
type SysTeamRole struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null" json:"deleted_at"`
	TeamID    uint32                `gorm:"column:team_id;type:int unsigned;not null;comment:团队ID" json:"team_id"` // 团队ID
	Name      string                `gorm:"column:name;type:varchar(64);not null;comment:角色名称" json:"name"`        // 角色名称
	Status    int                   `gorm:"column:status;type:int;not null;comment:状态" json:"status"`              // 状态
	Remark    string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`     // 备注
	Apis      []*SysTeamAPI         `gorm:"many2many:sys_team_role_apis" json:"apis"`
}

// String json string
func (c *SysTeamRole) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *SysTeamRole) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *SysTeamRole) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *SysTeamRole) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *SysTeamRole) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *SysTeamRole) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName SysTeamRole's table name
func (*SysTeamRole) TableName() string {
	return TableNameSysTeamRole
}
