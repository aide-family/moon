package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ imodel.ISendTemplate = (*SysSendTemplate)(nil)

// tableNameSysSendTemplate 发送模板表
const tableNameSysSendTemplate = "sys_send_template"

// SysSendTemplate 发送模板表
type SysSendTemplate struct {
	AllFieldModel
	// Name 发送模板名称
	Name string `gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx__p__name__send_template,priority:1;comment:发送模板名称"`
	// Content 模板内容
	Content string `gorm:"column:content;type:text;not null;comment:模板内容" json:"content"`
	// SendType 发送模板类型
	SendType vobj.AlarmSendType `gorm:"column:send_type;type:tinyint;not null;uniqueIndex:idx__p__name__send_template,priority:2;index:idx__send_template,priority:1;comment:发送模板类型"`
	Status   vobj.Status        `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`
	// Remark 模板备注
	Remark string `gorm:"column:remark;type:varchar(500);not null;comment:模板备注"`
}

// GetStatus get status
func (c *SysSendTemplate) GetStatus() vobj.Status {
	if types.IsNil(c) {
		return vobj.StatusUnknown
	}
	return c.Status
}

// GetName get name
func (c *SysSendTemplate) GetName() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Name
}

// GetContent get content
func (c *SysSendTemplate) GetContent() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Content
}

// GetSendType get send type
func (c *SysSendTemplate) GetSendType() vobj.AlarmSendType {
	if types.IsNil(c) {
		return vobj.AlarmSendTypeUnknown
	}
	return c.SendType
}

// GetRemark get remark
func (c *SysSendTemplate) GetRemark() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Remark
}

// MarshalBinary marshal binary
func (c *SysSendTemplate) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// UnmarshalBinary unmarshal binary
func (c *SysSendTemplate) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// TableName SysSendTemplate's table name
func (*SysSendTemplate) TableName() string {
	return tableNameSysSendTemplate
}
