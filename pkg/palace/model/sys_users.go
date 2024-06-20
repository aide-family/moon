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

const TableNameSysUser = "sys_users"

// SysUser mapped from table <sys_users>
type SysUser struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`
	Username  string                `gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx__su__username,priority:1;comment:用户名" json:"username"` // 用户名
	Nickname  string                `gorm:"column:nickname;type:varchar(64);not null;comment:昵称" json:"nickname"`                                           // 昵称
	Password  string                `gorm:"column:password;type:varchar(255);not null;comment:密码" json:"password"`                                          // 密码
	Email     string                `gorm:"column:email;type:varchar(64);not null;uniqueIndex:idx__su__email,priority:1;comment:邮箱" json:"email"`           // 邮箱
	Phone     string                `gorm:"column:phone;type:varchar(64);not null;uniqueIndex:idx__su__phone,priority:1;comment:手机号" json:"phone"`          // 手机号
	Remark    string                `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                              // 备注
	Avatar    string                `gorm:"column:avatar;type:varchar(255);not null;comment:头像" json:"avatar"`                                              // 头像
	Salt      string                `gorm:"column:salt;type:varchar(16);not null;comment:盐" json:"salt"`                                                    // 盐
	Gender    vobj.Gender           `gorm:"column:gender;type:int;not null;comment:性别" json:"gender"`                                                       // 性别
	Role      vobj.Role             `gorm:"column:role;type:int;not null;comment:系统默认角色类型" json:"role"`                                                     // 系统默认角色类型
	Status    vobj.Status           `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                       // 状态
}

// String json string
func (c *SysUser) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *SysUser) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *SysUser) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *SysUser) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *SysUser) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *SysUser) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName SysUser's table name
func (*SysUser) TableName() string {
	return TableNameSysUser
}
