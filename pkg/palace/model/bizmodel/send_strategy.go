package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSendStrategy = "send_strategies"

// SendStrategy 发送策略， 用于控制发送的消息完成抑制或聚合动作
type SendStrategy struct {
	model.AllFieldModel
	Name string `gorm:"column:name;type:varchar(64);not null;comment:策略名称" json:"name"`
	// 标签组
	Labels vobj.Labels `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	// 类型
	SendType vobj.SendType `gorm:"column:send_type;type:int;not null;comment:类型" json:"send_type"`
	Remark   string        `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status   vobj.Status   `gorm:"column:status;type:int;not null;comment:策略状态" json:"status"`
	// 持续时间，0为永久有效
	Duration *types.Duration `gorm:"column:duration;type:bigint(20);not null;comment:持续时间" json:"duration"`
}

// String json string
func (c *SendStrategy) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SendStrategy) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SendStrategy) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SendStrategy's table name
func (*SendStrategy) TableName() string {
	return tableNameSendStrategy
}
