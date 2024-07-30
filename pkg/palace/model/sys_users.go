package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysUser = "sys_users"

// SysUser mapped from table <sys_users>
type SysUser struct {
	AllFieldModel
	Username string      `gorm:"column:username;type:varchar(64);not null;uniqueIndex:idx__su__username,priority:1;comment:用户名" json:"username"` // 用户名
	Nickname string      `gorm:"column:nickname;type:varchar(64);not null;comment:昵称" json:"nickname"`                                           // 昵称
	Password string      `gorm:"column:password;type:varchar(255);not null;comment:密码" json:"password"`                                          // 密码
	Email    string      `gorm:"column:email;type:varchar(64);not null;uniqueIndex:idx__su__email,priority:1;comment:邮箱" json:"email"`           // 邮箱
	Phone    string      `gorm:"column:phone;type:varchar(64);not null;uniqueIndex:idx__su__phone,priority:1;comment:手机号" json:"phone"`          // 手机号
	Remark   string      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                              // 备注
	Avatar   string      `gorm:"column:avatar;type:varchar(255);not null;comment:头像" json:"avatar"`                                              // 头像
	Salt     string      `gorm:"column:salt;type:varchar(16);not null;comment:盐" json:"salt"`                                                    // 盐
	Gender   vobj.Gender `gorm:"column:gender;type:int;not null;comment:性别" json:"gender"`                                                       // 性别
	Role     vobj.Role   `gorm:"column:role;type:int;not null;comment:系统默认角色类型" json:"role"`                                                     // 系统默认角色类型
	Status   vobj.Status `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                       // 状态
}

// String json string
func (c *SysUser) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis缓存实现
func (c *SysUser) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis缓存实现
func (c *SysUser) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysUser's table name
func (*SysUser) TableName() string {
	return tableNameSysUser
}
