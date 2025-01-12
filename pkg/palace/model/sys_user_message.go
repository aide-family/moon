package model

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysUserMessage = "sys_user_messages"

// SysUserMessage mapped from table <sys_user_messages>
type SysUserMessage struct {
	AllFieldModel
	Content  string               `gorm:"column:name;type:varchar(255);not null;comment:菜单名称" json:"content"`
	Category vobj.UserMessageType `gorm:"column:category;type:tinyint;not null;comment:消息类型" json:"category"`
	UserID   uint32               `gorm:"column:user_id;type:int unsigned;not null;comment:用户ID;index:idx__user_msg__user_id,priority:1;" json:"userID"`
	Biz      vobj.BizType         `gorm:"column:biz;type:tinyint;not null;comment:业务类型" json:"biz"`
	BizID    uint32               `gorm:"column:biz_id;type:int unsigned;not null;comment:业务ID" json:"bizID"`

	User *SysUser `gorm:"foreignKey:UserID" json:"user"`
}

// String json string
func (c *SysUserMessage) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysUserMessage) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysUserMessage) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysAPI's table name
func (*SysUserMessage) TableName() string {
	return tableNameSysUserMessage
}

// GetContent get content
func (c *SysUserMessage) GetContent() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Content
}

// GetCategory get category
func (c *SysUserMessage) GetCategory() vobj.UserMessageType {
	if types.IsNil(c) {
		return vobj.UserMessageTypeUnknown
	}
	return c.Category
}

// GetUserID get user id
func (c *SysUserMessage) GetUserID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.UserID
}

// GetBiz get biz
func (c *SysUserMessage) GetBiz() vobj.BizType {
	if types.IsNil(c) {
		return vobj.BizTypeUnknown
	}
	return c.Biz
}

// GetBizID get biz id
func (c *SysUserMessage) GetBizID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.BizID
}

// GetUser get user
func (c *SysUserMessage) GetUser() *SysUser {
	if types.IsNil(c) {
		return nil
	}
	return c.User
}
