package model

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysOAuthUser = "sys_oauth_users"

// SysOAuthUser mapped from table <sys_oauth_users>
type SysOAuthUser struct {
	AllFieldModel
	OAuthID   uint32        `gorm:"column:oauth_id;type:int unsigned;index:uk__oauth_id__sys_user_id__app,unique" json:"oauth_id"`
	SysUserID uint32        `gorm:"column:sys_user_id;type:int unsigned;not null;comment:关联用户id;index:uk__oauth_id__sys_user_id__app,unique" json:"sys_user_id"`
	Row       string        `gorm:"column:row;type:text;comment:github用户信息" json:"row"`
	APP       vobj.OAuthAPP `gorm:"column:app;type:tinyint;not null;comment:oauth应用;index:uk__oauth_id__sys_user_id__app,unique" json:"app"`
}

// String json string
func (c *SysOAuthUser) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis缓存实现
func (c *SysOAuthUser) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis缓存实现
func (c *SysOAuthUser) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysUser's table name
func (*SysOAuthUser) TableName() string {
	return tableNameSysOAuthUser
}
