package model

import (
	"encoding/json"
)

const tableNameSysGithubUser = "sys_github_users"

// SysGithubUser mapped from table <sys_github_users>
type SysGithubUser struct {
	AllFieldModel
	GithubUserID uint32 `gorm:"column:github_user_id;type:int unsigned;index" json:"github_user_id"`
	SysUserID    uint32 `gorm:"column:sys_user_id;type:int unsigned;not null;comment:关联用户id" json:"sys_user_id"`
	Row          string `gorm:"column:row;type:text;comment:github用户信息" json:"row"`
}

// String json string
func (c *SysGithubUser) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis缓存实现
func (c *SysGithubUser) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis缓存实现
func (c *SysGithubUser) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysUser's table name
func (*SysGithubUser) TableName() string {
	return tableNameSysGithubUser
}
