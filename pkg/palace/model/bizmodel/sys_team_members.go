package bizmodel

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

const TableNameSysTeamMember = "sys_team_members"

// SysTeamMember mapped from table <sys_team_members>
type SysTeamMember struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null" json:"deleted_at"`
	UserID    uint32                `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id,priority:1;comment:系统用户ID" json:"user_id"` // 系统用户ID
	TeamID    uint32                `gorm:"column:team_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id,priority:2;comment:团队ID" json:"team_id"`   // 团队ID
	Status    vobj.Status           `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                              // 状态
	Role      vobj.Role             `gorm:"column:role;type:int;not null;comment:是否是管理员" json:"role"`                                                              // 是否是管理员
	TeamRoles []*SysTeamRole        `gorm:"many2many:sys_team_member_roles" json:"team_roles"`
}

// String json string
func (c *SysTeamMember) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *SysTeamMember) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *SysTeamMember) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *SysTeamMember) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *SysTeamMember) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *SysTeamMember) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName SysTeamMember's table name
func (*SysTeamMember) TableName() string {
	return TableNameSysTeamMember
}
